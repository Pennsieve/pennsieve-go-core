package pgdb

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/pennsieve/pennsieve-go-core/pkg/pgdb/models"
)

// GetOrganization returns a single organization
func (q *Queries) GetOrganization(ctx context.Context, id int64) (*models.Organization, error) {

	queryStr := "SELECT id, name, slug, node_id, storage_bucket, publish_bucket, embargo_bucket, created_at, updated_at " +
		"FROM pennsieve.organizations WHERE id=$1;"

	var organization models.Organization
	row := q.db.QueryRowContext(ctx, queryStr, id)
	err := row.Scan(
		&organization.Id,
		&organization.Name,
		&organization.Slug,
		&organization.NodeId,
		&organization.StorageBucket,
		&organization.PublishBucket,
		&organization.EmbargoBucket,
		&organization.CreatedAt,
		&organization.UpdatedAt)

	switch err {
	case sql.ErrNoRows:
		fmt.Println("No rows were returned!")
		return nil, err
	case nil:
		return &organization, nil
	default:
		panic(err)
	}
}
