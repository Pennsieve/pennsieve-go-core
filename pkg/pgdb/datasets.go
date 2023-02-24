package pgdb

import (
	"context"
	"github.com/pennsieve/pennsieve-go-core/pkg/pgdb/models"
	log "github.com/sirupsen/logrus"
)

// GetDatasets returns all rows in the Upload Record Table
func (q *Queries) GetDatasets(ctx context.Context, organizationId int) ([]models.Dataset, error) {
	queryStr := "SELECT (name, state) FROM datasets"

	rows, err := q.db.QueryContext(ctx, queryStr)
	var allDatasets []models.Dataset
	if err == nil {
		for rows.Next() {
			var currentRecord models.Dataset
			err = rows.Scan(
				&currentRecord.Name,
				&currentRecord.State)

			if err != nil {
				log.Println("ERROR: ", err)
			}

			allDatasets = append(allDatasets, currentRecord)
		}
		return allDatasets, err
	}
	return allDatasets, err
}
