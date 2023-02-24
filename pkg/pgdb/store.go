package pgdb

import (
	"context"
	"database/sql"
	"fmt"
)

// SQLStore provides the Queries interface and a db instance.
type SQLStore struct {
	*Queries
	db *sql.DB
}

// NewSQLStore returns a SQLStore object which implements the Queires
func NewSQLStore(db *sql.DB) *SQLStore {
	return &SQLStore{
		db:      db,
		Queries: New(db),
	}
}

func (store *SQLStore) execTx(ctx context.Context, fn func(*Queries) error) error {
	tx, err := store.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}

	q := New(tx)
	err = fn(q)
	if err != nil {
		if rbErr := tx.Rollback(); rbErr != nil {
			return fmt.Errorf("tx err: %v, rb err: %v", err, rbErr)
		}
		return err
	}

	return tx.Commit()
}

// ImportFiles creates rows for uploaded files in Packages and Files tables as a transaction
func (store *SQLStore) ImportFiles(ctx context.Context, records []PackageParams) ([]Package, error) {
	var result []Package

	err := store.execTx(ctx, func(q *Queries) error {
		// TODO: add packages
		// TODO: add files
		return nil
	})

	return result, err
}
