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
		log.Println("Error incrementing dataset size: ", err)
	}

	return err
}

func (q *Queries) GetDatasetStorageById(ctx context.Context, datasetId int64) (int64, error) {

	datasetSize := int64(0)
	err := q.db.QueryRowContext(ctx,
		"select p.size from dataset_storage as p where p.dataset_id = $1;",
		datasetId).Scan(&datasetSize)

	if err != nil {
		log.Error("unable to get dataset size", err)
		return int64(0), err
	}

	return datasetSize, nil
}
