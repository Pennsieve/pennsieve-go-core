package pgdb

import (
	"context"
	log "github.com/sirupsen/logrus"
)

// IncrementDatasetStorage increases the storage associated with the provided dataset.
func (q *Queries) IncrementDatasetStorage(ctx context.Context, datasetId int64, size int64) error {

	queryStr := "INSERT INTO dataset_storage " +
		"AS dataset_storage (dataset_id, size) " +
		"VALUES ($1, $2) ON CONFLICT (dataset_id) " +
		"DO UPDATE SET size = COALESCE(dataset_storage.size, 0) + EXCLUDED.size"

	_, err := q.db.ExecContext(ctx, queryStr, datasetId, size)
	if err != nil {
		log.Println("Error incrementing package size: ", err)
	}

	return err
}
