//services/billing-service/internal/usage/aggregator.go

package usage

import (
	"fmt"
	"sync"
)

//The Engine Managing Workers and Events

type Aggregator struct {
	mu      sync.Mutex         //Read-write mutex—protects the buckets map (briefly, to avoid slowing everything).
	buckets map[string]*Bucket //Map of usage buckets, keyed by tenant/account ID.
	//key is "TenantID:Type" (e.g., "ABC:SHIPMENT_CREATED"), value is a Bucket (counter).
	eventChan chan UsagesEvent //Channel for incoming usage events.
	//Buffered channel (holds up to 1000 events)—like a queue for pending work.

	quitChan chan struct{}  //Channel to signal shutdown of the aggregator.
	wg       sync.WaitGroup //WaitGroup to track active worker goroutines.

}

func NewAggregator() *Aggregator {
	return &Aggregator{
		buckets:   make(map[string]*Bucket),
		eventChan: make(chan UsagesEvent, 1000), //Buffered channel for incoming events.
		quitChan:  make(chan struct{}),          //Channel to signal shutdown.
	}
}

// Start Launching the backgerround woorkers to process usage events
func (agg *Aggregator) Start(workers int) {
	for i := 0; i < workers; i++ {
		agg.wg.Add(1)
		go agg.Worker(i)
	}
}

// Ingest is the public method to add events (Thread-Safe)
func (agg *Aggregator) Ingest(event UsagesEvent) {
	//case a.eventChan <- e: Tries to send (<-) the event e to the channel eventChan. If successful (channel has space), it adds the event and continues (comment: "Event sent successfully").
	//default: If the send fails (channel full), run this instead—print a warning with the event's ID.
	select {
	case agg.eventChan <- event:
		//Event sent successfully
	default:
		// Channel full: In production, log error or push to Dead Letter Queue
		fmt.Println("⚠️ Aggregator channel full, dropping event:", event.id)
	}

}

func (agg *Aggregator) Worker(id int) {
	defer agg.wg.Done()
	for {
		select {
		case event := <-agg.eventChan:
			agg.Process(event)
		case <-agg.quitChan:
			// Process remaining events in channel before quitting?
			// For simplicity, we quit immediately, but in Prod we drain.
			return
		}

	}

}
func (agg *Aggregator) Process(event UsagesEvent) {
	key := fmt.Sprintf("%s:%s", event.TenantID, event.Type)
	agg.mu.Lock()
	bucket, exist := agg.buckets[key]
	if !exist {
		bucket = &Bucket{
			TenantID: event.TenantID,
			Type:     event.Type,
			Count:    0,
		}
		agg.buckets[key] = bucket //Add new bucket to map
	}
	agg.mu.Unlock()
	//Now increnent the bucket count (outside lock to minimize contention)
	bucket.Increment(event.Quantity)

}
