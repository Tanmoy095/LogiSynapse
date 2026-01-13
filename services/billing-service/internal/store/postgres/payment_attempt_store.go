package PostgresStore

import "database/sql"

type PaymentAttemptStore struct {
	db *sql.DB
}

func NewPaymentAttemptStore(db *sql.DB) *PaymentAttemptStore {
	return &PaymentAttemptStore{db: db}
}
