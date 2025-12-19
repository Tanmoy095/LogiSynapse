package usage

import (
	"sync"

	"github.com/docker/distribution/uuid"
)

type Bucket struct {
	mu       sync.Mutex
	TenantID uuid.UUID
	Type     UsagesType
	Count    int64
}

// Increment safely adds to the count
func (b *Bucket) Increment(amount int64) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.Count += amount
}

// GetCount safely reads the count
func (b *Bucket) GetCount() int64 {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.Count
}
