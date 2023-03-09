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

func (q *Queries) GetOrganizationStorageById(ctx context.Context, organizationId int64) (int64, error) {

	orgSize := int64(0)
	err := q.db.QueryRowContext(ctx,
		"select p.size from pennsieve.organization_storage as p where p.organization_id = $1;",
		organizationId).Scan(&orgSize)

	if err != nil {
		log.Error("unable to get organization size", err)
		return int64(0), err
	}

	return orgSize, nil
}
