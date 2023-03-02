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

	// NOTE: When you create a new transaction (as below), the s.pgdb is NOT part of the transaction.
	// This has the following impact
	// 1. If you have set the search-path for the pgdb, the search path is no longer applied to s.pgdb
	// 2. Any funtion that is wrapped in the execTx method should ONLY use the provided queries struct that wraps the transaction.
	// 3. To enable custom Queries for a service, we wrap the pgdb.Queries in a service specific Queries struct.
	//	  This enables you to create custom queries within the service that leverage the transaction
	//    You can use the exposed db property of the Queries struct to create custom database interactions.
	//	  See the "upload-service-v2/upload lambda" for an example

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
