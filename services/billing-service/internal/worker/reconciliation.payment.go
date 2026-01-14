//services/billing-service/internal/worker/reconciliation.payment.go

package worker

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/Tanmoy095/LogiSynapse/services/billing-service/internal/payment"
)

/* call Stripe to charge the card
Now bad things can happen:
Internet dies
Server crashes
App restarts
Webhook never arrives
⚠️ But the money may already be taken
So sometimes db says:
❌ PENDING
✅ Money actually paid
This worker exists only to fix that mismatch

The Reconciler checks old PENDING payments and asks Stripe what REALLY happened, then fixes the database.

Every 5 minutes:
Find payments that are:
Still PENDING
Older than 5 minutes (suspicious)
For each payment:
Ask Stripe: “Did this payment succeed or fail?”
Update your database to match Stripe
*/

// It finds payments that are stuck in "PENDING" and syncs them with Stripe.
type Reconciler struct {
	paymentService *payment.PaymentService     // business logic
	attemptStore   payment.PaymentAttemptStore //payment db
	gateway        payment.StripeGateway

	//setting
	batchSize   int //how many to process at once
	workerCount int //how many goroutines to run in parallel
}

func NewReconciler(
	paymentService *payment.PaymentService,
	attemptStore payment.PaymentAttemptStore,
	gateway payment.StripeGateway,
) *Reconciler {
	return &Reconciler{
		paymentService: paymentService,
		attemptStore:   attemptStore,
		gateway:        gateway,
		batchSize:      50, // Process 50 items per tick
		workerCount:    5,  // 5 concurrent goroutines
	}
}

// Start runs the worker loop. blocking call.
func (r *Reconciler) Start(ctx context.Context) {
	ticker := time.NewTicker(5 * time.Minute) // Run every 5 minutes
	defer ticker.Stop()
	log.Println("[Reconciler] Worker started. Polling every 5m.")
	for {
		select {
		case <-ctx.Done():
			log.Println("[Reconciler] Context cancelled, stopping.")
			return
		case <-ticker.C:
			log.Println("[Reconciler] Tick - starting reconciliation cycle.")
			r.processBatch(ctx)
		}
	}
}

// processBatch orchestrates the worker pool
func (r *Reconciler) processBatch(ctx context.Context) {
	// find Zombie
	// "Give me pending attempts created > 5 mins ago"

	attempts, err := r.attemptStore.GetPendingAttempts(ctx, r.batchSize, 5*time.Minute)
	if err != nil {
		log.Printf("[Reconciler] DB Error: %v", err)
		return
	}
	if len(attempts) == 0 { //this means no pending attempts found
		log.Println("[Reconciler] No pending payment attempts found.")
		return
	}
	log.Printf("[Reconciler] Processing %d stuck payments...", len(attempts))

	//  Worker Pool Setup
	jobs := make(chan payment.PaymentAttempt, len(attempts)) //this channel holds the payment attempts to process. buffered to avoid blocking
	var wg sync.WaitGroup
	// Start Workers
	for w := 0; w < r.workerCount; w++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			for attempt := range jobs {
				if err := r.syncAttempt(ctx, attempt); err != nil {
					log.Printf("[Reconciler] Worker %d failed on %s: %v", id, attempt.AttemptID, err)
				}
			}

		}(w)

	}
	// Enqueue Jobs
	//Feed the Workers. means sending payment attempts to jobs channel
	for _, attempt := range attempts {
		jobs <- *attempt //send attempt to jobs channel. here  jobs is buffered so this won't block it will only block if buffer is full.

	}
	close(jobs) //close jobs channel to signal workers no more jobs will be sent
	wg.Wait()   //wait for all workers to finish
	log.Println("[Reconciler] Reconciliation cycle completed.")

}

// syncAttempt is the core logic: DB(Pending) vs Stripe(???)
func (r *Reconciler) syncAttempt(ctx context.Context, attempt payment.PaymentAttempt) error {
	// Case A: We crashed BEFORE we even got a PaymentID from Stripe.
	// We cannot verify this with Stripe because we don't have an ID.
	// Result: Mark FAILED (Safe to retry).
	if attempt.ProviderPaymentID == "" {
		failMsg := "Stuck PENDING with no Provider ID (Crash before API call)"
		code := "system_crash_pre_flight"
		return r.attemptStore.UpdateAttemptStatus(ctx, attempt.AttemptID, payment.PaymentFailed, "", &code, &failMsg)
	}
	// Case B: We have an ID. Ask Stripe what happened.
	realStatus, err := r.gateway.GetPaymentStatus(ctx, attempt.ProviderPaymentID)
	if err != nil {
		return fmt.Errorf("gateway check failed: %w", err)
	}
	log.Printf("[Reconciler] Attempt %s (Local: PENDING) -> Stripe says: %s", attempt.AttemptID, realStatus)
	// Case C: Stripe says SUCCEEDED
	if realStatus == payment.PaymentSucceeded {
		//update local attempt to succeeded
		if err := r.attemptStore.UpdateAttemptStatus(ctx, attempt.AttemptID, payment.PaymentSucceeded, attempt.ProviderPaymentID, nil, nil); err != nil {
			return err
		}
		// 2. Finalize Invoice (The Logic we just fixed!)
		return r.paymentService.FinalizeSuccessfulPayment(
			ctx,
			attempt.InvoiceID,
			attempt.TenantID,
			attempt.AmountCents,
			attempt.Currency,
			attempt.ProviderPaymentID,
		)

	}
	// Case D: Stripe says FAILED
	if realStatus == payment.PaymentFailed {
		msg := "Reconciled from gateway"
		return r.attemptStore.UpdateAttemptStatus(ctx, attempt.AttemptID, payment.PaymentFailed, attempt.ProviderPaymentID, nil, &msg)

	}
	// Case E: Still Processing? Do nothing.
	return nil

}
