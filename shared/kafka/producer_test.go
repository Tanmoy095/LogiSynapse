package kafka

import (
	"context"
	"testing"

	skafka "github.com/segmentio/kafka-go"
)

// fakeWriter is a test writer that records messages written.
type fakeWriter struct {
	msgs []skafka.Message
}

func (f *fakeWriter) WriteMessages(ctx context.Context, msgs ...skafka.Message) error {
	f.msgs = append(f.msgs, msgs...)
	return nil
}

func (f *fakeWriter) Close() error { return nil }

func TestPublish(t *testing.T) {
	fw := &fakeWriter{}
	p := NewKafkaProducerWithWriter(fw)
	err := p.Publish(context.Background(), "key1", map[string]string{"a": "b"})
	if err != nil {
		t.Fatalf("publish failed: %v", err)
	}
	if len(fw.msgs) != 1 {
		t.Fatalf("expected 1 message, got %d", len(fw.msgs))
	}
}
