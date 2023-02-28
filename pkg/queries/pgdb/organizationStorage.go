package pgdb

import (
	"context"
	log "github.com/sirupsen/logrus"
)

// IncrementOrganizationStorage increases the storage associated with the provided organization.
func (q *Queries) IncrementOrganizationStorage(ctx context.Context, organizationId int64, size int64) error {
	
	queryStr := "INSERT INTO pennsieve.organization_storage " +
		"AS organization_storage (organization_id, size) " +
		"VALUES ($1, $2) ON CONFLICT (organization_id) " +
		"DO UPDATE SET size = COALESCE(organization_storage.size, 0) + EXCLUDED.size"

	_, err := q.db.ExecContext(ctx, queryStr, organizationId, size)
	if err != nil {
		log.Println("Error incrementing package size: ", err)
	}

	return err
}
