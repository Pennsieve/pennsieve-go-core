package pgdb

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/pennsieve/pennsieve-go-core/pkg/models/pgdb"
	log "github.com/sirupsen/logrus"
)

// GetDefaultDatasetStatus will return the default dataset status for the organization.
// This is assumed to be the dataset status row with the lowest id number.
func (q *Queries) GetDefaultDatasetStatus(ctx context.Context, organizationId int) (*pgdb.DatasetStatus, error) {
	query := fmt.Sprintf("SELECT id, name, display_name, original_name, color, created_at, updated_at"+
		" FROM \"%d\".dataset_status order by id limit 1;", organizationId)

	row := q.db.QueryRowContext(ctx, query)
	datasetStatus := pgdb.DatasetStatus{}
	err := row.Scan(
		&datasetStatus.Id,
		&datasetStatus.Name,
		&datasetStatus.DisplayName,
		&datasetStatus.OriginalName,
		&datasetStatus.Color,
		&datasetStatus.CreatedAt,
		&datasetStatus.UpdatedAt)

	if err != nil {
		switch err {
		case sql.ErrNoRows:
			log.Error("No rows were returned!")
			return nil, err
		default:
			log.Error("Unknown Error while scanning dataset_status table: ", err)
			panic(err)
		}
	}

	return &datasetStatus, nil
}
