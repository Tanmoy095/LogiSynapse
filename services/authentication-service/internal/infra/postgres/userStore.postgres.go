package postgres

import (
	"database/sql"
	//"github.com/Tanmoy095/LogiSynapse/services/authentication-service/internal/ports/repository"
)

// Ensure PostgresUserStore implements the interface at compile time
//var _ repository.UserStore = (*PostgresUserStore)(nil)

type PostgresUserStore struct {
	db *sql.DB
}

func NewPostgresUserStore(db *sql.DB) *PostgresUserStore {
	return &PostgresUserStore{
		db: db,
	}
}
