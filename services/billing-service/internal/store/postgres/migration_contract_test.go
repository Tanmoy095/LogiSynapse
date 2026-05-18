package PostgresStore

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestPaymentAttemptMigrationContract(t *testing.T) {
	path := filepath.Join("..", "..", "..", "db", "migrations", "008_create_payment_attempts.sql")
	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("failed to read migration file: %v", err)
	}
	sql := string(b)

	if !strings.Contains(sql, "attempt_id UUID PRIMARY KEY") {
		t.Fatalf("migration contract mismatch: expected attempt_id primary key")
	}
	if !strings.Contains(sql, "REFERENCES invoices(invoice_id)") {
		t.Fatalf("migration contract mismatch: expected FK to invoices(invoice_id)")
	}
}
