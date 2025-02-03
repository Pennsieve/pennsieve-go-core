package pgdb

import (
	"context"
	"github.com/pennsieve/pennsieve-go-core/pkg/models/pgdb"
)

// getOrganization returns (nil, sql.ErrNoRows) if no Organization is found
func (q *Queries) getOrganization(ctx context.Context, query string, value any) (*pgdb.Organization, error) {
	var organization pgdb.Organization
	row := q.db.QueryRowContext(ctx, query, value)
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

	if err != nil {
		return nil, err
	}
	return &organization, nil
}

// GetOrganization returns a single organization
func (q *Queries) GetOrganization(ctx context.Context, id int64) (*pgdb.Organization, error) {

	queryStr := "SELECT id, name, slug, node_id, storage_bucket, publish_bucket, embargo_bucket, created_at, updated_at " +
		"FROM pennsieve.organizations WHERE id=$1;"

	return q.getOrganization(ctx, queryStr, id)
}

// GetOrganizationByNodeId returns a single organization
func (q *Queries) GetOrganizationByNodeId(ctx context.Context, nodeId string) (*pgdb.Organization, error) {

	queryStr := "SELECT id, name, slug, node_id, storage_bucket, publish_bucket, embargo_bucket, created_at, updated_at " +
		"FROM pennsieve.organizations WHERE node_id=$1;"

	return q.getOrganization(ctx, queryStr, nodeId)
}

// GetOrganizationByName returns a single organization
func (q *Queries) GetOrganizationByName(ctx context.Context, name string) (*pgdb.Organization, error) {

	queryStr := "SELECT id, name, slug, node_id, storage_bucket, publish_bucket, embargo_bucket, created_at, updated_at " +
		"FROM pennsieve.organizations WHERE name=$1;"

	return q.getOrganization(ctx, queryStr, name)
}

// GetOrganizationBySlug returns a single organization
func (q *Queries) GetOrganizationBySlug(ctx context.Context, slug string) (*pgdb.Organization, error) {

	queryStr := "SELECT id, name, slug, node_id, storage_bucket, publish_bucket, embargo_bucket, created_at, updated_at " +
		"FROM pennsieve.organizations WHERE slug=$1;"

	return q.getOrganization(ctx, queryStr, slug)
}
