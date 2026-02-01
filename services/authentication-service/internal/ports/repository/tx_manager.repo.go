//services/authentication-service/internal/ports/repository/tx_manager.repo.go

package repository

import "context"

// TransactionManager interface abstracts the database transaction.
// DSA: Required for atomicity across two tables (requests + tenants).
type TransactionManager interface {
	RunInTx(ctx context.Context, fn func(ctx context.Context) error) error
}
