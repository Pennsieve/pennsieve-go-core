package pgdb

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/pennsieve/pennsieve-go-core/pkg/models/pgdb"
)

type DatasetReleaseNotFoundError struct {
	ErrorMessage string
}

func (e DatasetReleaseNotFoundError) Error() string {
	return fmt.Sprintf("dataset release was not found (error: %v)", e.ErrorMessage)
}

func (q *Queries) AddDatasetRelease(ctx context.Context, release pgdb.DatasetRelease) (*pgdb.DatasetRelease, error) {
	statement := "INSERT INTO dataset_release " +
		"(dataset_id, origin, url, label, marker, release_date, release_status, publishing_status) " +
		"VALUES ($1, $2, $3, $4, $5, $6, $7, $8) returning id;"

	//result, err := q.db.ExecContext(ctx, statement,
	//	release.DatasetId,
	//	release.Origin,
	//	release.Url,
	//	release.Label,
	//	release.Marker,
	//	release.ReleaseDate,
	//)

	var id int64
	err := q.db.QueryRowContext(ctx,
		statement,
		release.DatasetId,
		release.Origin,
		release.Url,
		release.Label,
		release.Marker,
		release.ReleaseDate,
		release.ReleaseStatus,
		release.PublishingStatus,
	).Scan(&id)

	if err != nil {
		return nil, fmt.Errorf(fmt.Sprintf("database error on insert: %v", err))
	}

	return q.GetDatasetReleaseById(ctx, id)
}

func (q *Queries) UpdateDatasetRelease(ctx context.Context, release pgdb.DatasetRelease) (*pgdb.DatasetRelease, error) {
	statement := "UPDATE dataset_release SET label=$1, marker=$2, release_date=$3, release_status=$4, publishing_status=$5 WHERE id=$6;"
	_, err := q.db.ExecContext(ctx, statement,
		release.Label,
		release.Marker,
		release.ReleaseDate,
		release.ReleaseStatus,
		release.PublishingStatus,
		release.Id,
	)

	if err != nil {
		return nil, fmt.Errorf(fmt.Sprintf("database error on update: %v", err))
	}

	return q.GetDatasetReleaseById(ctx, release.Id)
}

func (q *Queries) GetDatasetReleaseById(ctx context.Context, id int64) (*pgdb.DatasetRelease, error) {
	predicate := fmt.Sprintf("id = %d", id)
	return q.getDatasetRelease(ctx, predicate)
}

func (q *Queries) GetDatasetRelease(ctx context.Context, datasetId int64, label string, marker string) (*pgdb.DatasetRelease, error) {
	predicate := fmt.Sprintf("dataset_id = %d AND label = '%s' AND marker = '%s'", datasetId, label, marker)
	return q.getDatasetRelease(ctx, predicate)
}

func (q *Queries) getDatasetRelease(ctx context.Context, predicate string) (*pgdb.DatasetRelease, error) {
	query := fmt.Sprintf(
		"SELECT id, dataset_id, origin, url, label, marker, release_date, release_status, publishing_status, created_at, updated_at "+
			"FROM dataset_release WHERE %s;", predicate)

	var datasetRelease pgdb.DatasetRelease
	row := q.db.QueryRowContext(ctx, query)
	err := row.Scan(
		&datasetRelease.Id,
		&datasetRelease.DatasetId,
		&datasetRelease.Origin,
		&datasetRelease.Url,
		&datasetRelease.Label,
		&datasetRelease.Marker,
		&datasetRelease.ReleaseDate,
		&datasetRelease.ReleaseStatus,
		&datasetRelease.PublishingStatus,
		&datasetRelease.CreatedAt,
		&datasetRelease.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, DatasetReleaseNotFoundError{fmt.Sprintf("dataset release not found where %s", predicate)}
		} else {
			return nil, fmt.Errorf(fmt.Sprintf("database error on query: %v", err))
		}
	}

	return &datasetRelease, nil
}
