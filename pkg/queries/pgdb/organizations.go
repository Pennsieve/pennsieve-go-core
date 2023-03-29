package pgdb

import (
	"context"
	"database/sql"
	"github.com/pennsieve/pennsieve-go-core/pkg/models/pgdb"
	log "github.com/sirupsen/logrus"
)

// GetOrganization returns a single organization
func (q *Queries) GetOrganization(ctx context.Context, id int64) (*pgdb.Organization, error) {

	queryStr := "SELECT id, name, slug, node_id, storage_bucket, publish_bucket, embargo_bucket, created_at, updated_at " +
		"FROM pennsieve.organizations WHERE id=$1;"

	var organization pgdb.Organization
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
		log.Error("No rows were returned!")
		return nil, err
	case nil:
		return &organization, nil
	default:
		panic(err)
	}
}
