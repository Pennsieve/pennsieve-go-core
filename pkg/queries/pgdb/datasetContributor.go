package pgdb

import (
	"context"
	"github.com/pennsieve/pennsieve-go-core/pkg/models/pgdb"
)

func (q *Queries) GetDatasetContributor(ctx context.Context, datasetId int64, contributorId int64) (*pgdb.DatasetContributor, error) {
	query := "SELECT dataset_id, contributor_id, created_at, updated_at, contributor_order FROM dataset_contributor WHERE dataset_id=$1 and contributor_id=$2"
	var datasetContributor pgdb.DatasetContributor
	err := q.db.QueryRowContext(ctx, query, datasetId, contributorId).Scan(
		&datasetContributor.DatasetId,
		&datasetContributor.ContributorId,
		&datasetContributor.CreatedAt,
		&datasetContributor.UpdatedAt,
		&datasetContributor.ContributorOrder)

	if err != nil {
		return nil, err
	}
	return &datasetContributor, nil
}

func (q *Queries) AddDatasetContributor(ctx context.Context, dataset *pgdb.Dataset, contributor *pgdb.Contributor) (*pgdb.DatasetContributor, error) {
	position := int64(1)
	statement := "INSERT INTO dataset_contributor(dataset_id, contributor_id, contributor_order) VALUES($1, $2, $3)"
	_, err := q.db.ExecContext(ctx, statement, dataset.Id, contributor.Id, position)
	if err != nil {
		return nil, err
	}
	return q.GetDatasetContributor(ctx, dataset.Id, contributor.Id)
}
