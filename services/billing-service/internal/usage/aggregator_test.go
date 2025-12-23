// services/billing-service/internal/usage/aggregator_test.go
package usage

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"testing"
	"time"

	billingtypes "github.com/Tanmoy095/LogiSynapse/services/billing-service/internal/billingTypes"
	"github.com/Tanmoy095/LogiSynapse/services/billing-service/internal/store"
	"github.com/google/uuid"
)

type MockUsageStore struct {
	mu           sync.Mutex
	flushedData  map[string]int64 //
	shouldFail   bool             // Simulate failure default false .. Because bool zero value is false
	failureCount int              // Number of times to fail before succeeding
}

func newMockUsageStore() *MockUsageStore {
	return &MockUsageStore{
		flushedData: make(map[string]int64),
	}
}
func (m *MockUsageStore) Flush(ctx context.Context, batch store.FlushBatch) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.shouldFail { // m.shouldFail returns true so if it's true we simulate failure.. here shouldFail is set to true in the test function
		m.failureCount++ // Increment failure count
		return errors.New("simulated DB failure")
	}
	//
	for _, records := range batch.Records {
		// Create a unique key for each TenantID and UsageType combination
		key := fmt.Sprintf("%s:%s", records.TenantID, records.UsageType)
		// Accumulate the TotalQuantity for this key
		m.flushedData[key] = m.flushedData[key] + records.TotalQuantity //the map should look like map["tenantID-usageType"] = totalQuantity
	}

	return nil
}

// TestAggregator_Concurrency_Integrity tests that the Aggregator correctly aggregates usage data under concurrent ingestion
func TestAggregator_Concurrency_Integrity(t *testing.T) {
	mockUsageStore := newMockUsageStore()
	ctx := context.Background()
	agg := NewAggregator(ctx, mockUsageStore, 50*time.Microsecond)
	agg.Start(5) //5 workers for parallel processing
	tenantID := uuid.New()
	targetType := billingtypes.ShipmentCreated
	numEvents := 1000
	workers := 10
	var wg sync.WaitGroup
	t.Logf("starting ingestion: %d workers sending %d events each", workers, numEvents)
	for w := 0; w < workers; w++ {
		wg.Add(1) //Add a goroutine to the WaitGroup
		go func(workerID int) {
			defer wg.Done() //Signal completion when goroutine finishes
			for i := 0; i < numEvents; i++ {
				event := UsageEvent{
					ID:        fmt.Sprintf("event-%d-%d", workerID, i),
					TenantID:  tenantID,
					Type:      targetType,
					Quantity:  1,
					Timestamp: time.Now().Unix(),
				}
				agg.Ingest(event)
			}
		}(w)
	}
	wg.Wait() //Wait for all goroutines to finish
	t.Log("all events ingested, waiting for flush...")

	agg.Stop()
	//Assert data integrity
	expectedTotal := int64(numEvents * workers) // expected total is 1000 * 10 = 10000 for key "farmID:SHIPMENT_CREATED"
	key := fmt.Sprintf("%s:%s", tenantID, targetType)
	mockUsageStore.mu.Lock()
	actualTotal := mockUsageStore.flushedData[key]
	mockUsageStore.mu.Unlock()
	if actualTotal != expectedTotal {
		t.Errorf("data loss! Expected: %d, got %d", expectedTotal, actualTotal)
	} else {
		t.Logf("data integrity check passed: total %d", actualTotal)
	}

}

// TestAggregator_Flush_Failure_Recovery tests that the Aggregator can recover from flush failures
func TestAggregator_Flush_Failure_Recovery(t *testing.T) {
	MockUsageStore := newMockUsageStore()
	context := context.Background()
	agg := NewAggregator(context, MockUsageStore, 1*time.Hour) //Long flush interval to avoid auto flush during test
	agg.Start(1)
	tenantID := uuid.New()
	//Ingest some events
	event1 := UsageEvent{
		ID:        "event-1",
		TenantID:  tenantID,
		Type:      "billingTypes.SHIPMENT_CREATED",
		Quantity:  10,
		Timestamp: time.Now().Unix(),
	}
	agg.Ingest(event1)
	time.Sleep(50 * time.Millisecond) //Give some time for processing
	MockUsageStore.shouldFail = true  //Simulate failure . means Flush will return error

	err := agg.Flush(context) //Manual flush. because auto flush interval is 1 hour
	if err == nil {           //err == nil means flush succeeded so it's unexpected because we set shouldFail = true
		t.Errorf("expected flush to fail, but it succeeded") // why expected to fail? because we set shouldFail = true
	}
	event2 := UsageEvent{
		ID:        "event-2",
		TenantID:  tenantID,
		Type:      "billingTypes.SHIPMENT_CREATED",
		Quantity:  5,
		Timestamp: time.Now().Unix(),
	}
	agg.Ingest(event2)
	time.Sleep(50 * time.Millisecond) //Give some time for processing
	MockUsageStore.shouldFail = false //now disable failure simulation. it means next Flush should succeed
	err = agg.Flush(context)          //Manual flush
	if err != nil {
		t.Errorf("expected flush to succeed, but it failed: %v", err)
	}
	key := fmt.Sprintf("%s:%s", tenantID, "billingTypes.SHIPMENT_CREATED")
	actualTotal := MockUsageStore.flushedData[key] //should be 15 (10 from event1 + 5 from event2)
	if actualTotal != 15 {
		t.Errorf("data loss after recovery! Expected total 15, got %d", actualTotal)
	} else {
		t.Logf("recovery successful, total %d", actualTotal)
	}
	agg.Stop()
}

// TestAggregator_IgnoreInvalidQuantity tests that the Aggregator ignores events with non-positive quantities
func TestAggregator_IgnoreInvalidQuantity(t *testing.T) {
	mockStore := newMockUsageStore()
	agg := NewAggregator(context.Background(), mockStore, 1*time.Hour)
	agg.Start(1)
	defer agg.Stop()

	tenantID := uuid.New()

	// Ingest invalid data
	agg.Ingest(UsageEvent{ID: "bad", TenantID: tenantID, Type: billingtypes.APIRequest, Quantity: -100})
	agg.Ingest(UsageEvent{ID: "zero", TenantID: tenantID, Type: billingtypes.APIRequest, Quantity: 0})

	// Ingest valid data
	agg.Ingest(UsageEvent{ID: "good", TenantID: tenantID, Type: billingtypes.APIRequest, Quantity: 1})

	// Wait & Flush
	time.Sleep(10 * time.Millisecond)
	agg.Flush(context.Background())

	// Check Store
	key := fmt.Sprintf("%s:%s", tenantID, billingtypes.APIRequest)
	actual := mockStore.flushedData[key]

	if actual != 1 {
		t.Errorf("❌ Validation Failed! Expected 1, got %d (Did it count negatives?)", actual)
	}
}
