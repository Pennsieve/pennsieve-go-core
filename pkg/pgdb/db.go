package pgdb

import (
	"context"
	"database/sql"
)

// From https://dev.to/techschoolguru/a-clean-way-to-implement-database-transaction-in-golang-2ba

// DBTX Default interface with methods that are available for both DB adn TX sessions.
type DBTX interface {
	ExecContext(context.Context, string, ...interface{}) (sql.Result, error)
	PrepareContext(context.Context, string) (*sql.Stmt, error)
	QueryContext(context.Context, string, ...interface{}) (*sql.Rows, error)
	QueryRowContext(context.Context, string, ...interface{}) *sql.Row
}

// New returns a Queries object backed by a DBTX interface (either DB or TX)
func New(db DBTX) *Queries {
	return &Queries{db: db}
}

// Queries is a struct with a db object that implements the DBTX interface.
// This means that db can either be a direct DB connection or a TX transaction.
type Queries struct {
	db DBTX
}

// WithTx Returns a new Queries object wrapped by a transactions.
func (q *Queries) WithTx(tx *sql.Tx) *Queries {
	return &Queries{
		db: tx,
	}
}
