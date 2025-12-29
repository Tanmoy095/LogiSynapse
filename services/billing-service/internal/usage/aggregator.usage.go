//services/billing-service/internal/usage/aggregator.go

package usage

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/Tanmoy095/LogiSynapse/services/billing-service/internal/store"
	"github.com/google/uuid"
)

// The Engine Managing Workers and Events
type Aggregator struct {
	mu      sync.Mutex         //Read-write mutex—protects the buckets map (briefly, to avoid slowing everything).
	buckets map[string]*Bucket //Map of usage buckets, keyed by tenant/account ID.
	//key is "TenantID:Type" (e.g., "ABC:SHIPMENT_CREATED"), value is a Bucket (counter).
	eventChan chan UsageEvent //Channel for incoming usage events.
	//Buffered channel (holds up to 1000 events)—like a queue for pending work.
	ctx context.Context // Root context

	quitChan      chan struct{}    //Channel to signal shutdown of the aggregator.
	wg            sync.WaitGroup   //WaitGroup to track active worker goroutines.
	store         store.UsageStore //Persistent store for flushed usage data.
	flushInterval time.Duration    //Interval between flushes to the store. e.g. 30*time.Second
}

func NewAggregator(ctx context.Context, store store.UsageStore, flushInterval time.Duration) *Aggregator {
	return &Aggregator{
		buckets:       make(map[string]*Bucket),
		eventChan:     make(chan UsageEvent, 10000), //Buffered channel for incoming events.
		quitChan:      make(chan struct{}),          //Channel to signal shutdown.
		ctx:           ctx,
		store:         store,
		flushInterval: flushInterval,
	}
}

// Start Launching the background workers to process usage events
func (agg *Aggregator) Start(workers int) {
	for i := 0; i < workers; i++ {
		agg.wg.Add(1)
		go agg.Worker(i)
	}

	//Start flusher goroutine
	agg.wg.Add(1)
	go agg.Flusher()
	fmt.Printf("🚀 Aggregator started with %d workers and flush every %v\n", workers, agg.flushInterval)
}

// Ingest is the public method to add events (Thread-Safe)
func (agg *Aggregator) Ingest(event UsageEvent) {
	// 1. VALIDATION: Prevent negative/zero billing
	if event.Quantity <= 0 {
		fmt.Printf("⚠️ Skipping invalid quantity: %d for event %s\n", event.Quantity, event.ID)
		return
	}
	//case a.eventChan <- e: Tries to send (<-) the event e to the channel eventChan. If successful (channel has space), it adds the event and continues (comment: "Event sent successfully").
	//default: If the send fails (channel full), run this instead—print a warning with the event's ID.
	select {
	case agg.eventChan <- event:
		//Event sent successfully
	default:
		// Channel full: In production, log error or push to Dead Letter Queue
		fmt.Println("⚠️ Aggregator channel full, dropping event:", event.ID)
	}

}

func (agg *Aggregator) Worker(id int) {
	defer agg.wg.Done()
	for {
		select {
		case event := <-agg.eventChan:
			agg.Process(event)
		case <-agg.quitChan:
			//Quit signal received,
			//we must drain of any remaining events before exiting.
			// drain means process all remaining events in the channel
			//if we dont we lose everything currently buffered in eventChan
			for {
				select {
				case event := <-agg.eventChan:
					agg.Process(event)
				default:
					// Channel is empty, exit the worker
					return
				}
			}
		}

	}

}
func (agg *Aggregator) Process(event UsageEvent) {
	key := fmt.Sprintf("%s:%s", event.TenantID, event.Type) //e.g., "Tenant123:SHIPMENT_CREATED"
	//Get or create the bucket (with lock)
	agg.mu.Lock()
	defer agg.mu.Unlock() //unlock after getting/creating bucket
	bucket, exist := agg.buckets[key]
	if !exist {
		bucket = &Bucket{
			TenantID: event.TenantID,
			Type:     event.Type,
			Count:    0,
		}
		agg.buckets[key] = bucket //Add new bucket to map
	}
	//Now increment the bucket count (outside lock to minimize contention)
	// Move Increment INSIDE the lock.
	// This prevents the Flusher from swapping the map while we are updating.
	bucket.Increment(event.Quantity)

}

// Every 30 second, flush() saves deltas    --e.g., if 100 events processed,saves sums to store,resets counters
func (agg *Aggregator) Flusher() {
	defer agg.wg.Done()
	ticker := time.NewTicker(agg.flushInterval) //Ticker triggers flushes at intervals
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C: //On each tick, flush the current usage data
			if err := agg.Flush(agg.ctx); err != nil {
				//In production, log error
				fmt.Println("❌ Aggregator flush error:", err)
			} else {
				fmt.Println("✅ Aggregator flush successful")
			}
		case <-agg.quitChan:
			//Quit signal received, exit flusher
			return

		}
	}
}
func (agg *Aggregator) Stop() {
	fmt.Println("🛑 Stopping Aggregator...")
	close(agg.quitChan) //Signal all goroutines to quit
	agg.wg.Wait()       //Wait for all workers and flusher to finish
	//Final flush before exit
	if err := agg.Flush(context.Background()); err != nil {
		fmt.Println("❌ Final Aggregator flush error:", err)
	} else {
		fmt.Println("✅ Final Aggregator flush successful")
	}
	fmt.Println("🛑 Aggregator stopped.")

}

// Flush swaps current map of buckets with a new empty one, then persists the old data
// and resets counters
func (agg *Aggregator) Flush(ctx context.Context) error {
	agg.mu.Lock()
	if len(agg.buckets) == 0 {
		fmt.Println("ℹ️ No usage data to flush.")
		agg.mu.Unlock()
		return nil

	}
	//Swap maps
	oldBuckets := agg.buckets              //Capture the current full map of Buckets
	agg.buckets = make(map[string]*Bucket) //Replace with a new empty map
	agg.mu.Unlock()                        //Unlock quickly to allow new events to be processed

	// Build the Batch from oldBuckets
	batch := store.FlushBatch{
		BatchID: uuid.New(),
		Records: make([]store.UsageRecord, 0, len(oldBuckets)), //Preallocate slice , 0 means no initial elements, len(oldBuckets) is capacity
	}
	now := time.Now().UTC() // Gets current time in UTC, e.g., 2024-06-01 12:00:00 +0000 UTC

	for _, bucket := range oldBuckets {
		batch.Records = append(batch.Records, store.UsageRecord{
			TenantID:      bucket.TenantID,   //"CompanyX"
			UsageType:     bucket.Type,       //"SHIPMENT_CREATED"
			TotalQuantity: bucket.GetCount(), //e.g., 150
			BillingPeriod: store.BillingPeriod{
				Year:  now.Year(),
				Month: int(now.Month()),
			},
		})
	}
	//Flush batch with new Unique batch id  and for each bucket create usage record and append to batch.Records

	//Now its time to persist the batch to the store
	maxRetries := 3
	for attempt := 1; attempt <= maxRetries; attempt++ {
		err := agg.store.Flush(agg.ctx, batch)
		if err == nil {
			fmt.Printf("✅ Flush successful on attempt %d\n", attempt)
			return nil // Success
		}
		fmt.Printf("Flush attempt %d failed: %v\n", attempt, err)
		time.Sleep(time.Second * time.Duration(attempt)) // Backoff
	}
	//On failure marge back to buckets to avoid data loss
	agg.mu.Lock()
	for key, bucket := range oldBuckets { //Iterate old buckets
		existingBucket, exists := agg.buckets[key] //Check if bucket already exists
		if exists {
			//Merge counts
			existingBucket.Increment(bucket.GetCount()) //Add old count to existing bucket
		} else {
			// Re-add the old bucket
			agg.buckets[key] = bucket //Re-add the old bucket
			//This ensures no data loss on flush failure..
		}
	}
	agg.mu.Unlock()
	return fmt.Errorf("failed to flush usage data after %d attempts", maxRetries)

}
