package dbTable

import (
	"github.com/pennsieve/pennsieve-go-core/pkg/core"
	"log"
)

type DatasetStorage struct {
	DatasetId int64 `json:"dataset_id"`
	Size      int64 `json:"size"`
}

// Increment increases the storage associated with the provided dataset.
func (d *DatasetStorage) Increment(db core.PostgresAPI, datasetId int64, size int64) error {

	queryStr := "INSERT INTO dataset_storage " +
		"AS dataset_storage (dataset_id, size) " +
		"VALUES ($1, $2) ON CONFLICT (dataset_id) " +
		"DO UPDATE SET size = COALESCE(dataset_storage.size, 0) + EXCLUDED.size"

	_, err := db.Exec(queryStr, datasetId, size)
	if err != nil {
		log.Println("Error incrementing package size: ", err)
	}

	return err
}
