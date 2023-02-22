package dbTable

import (
	"github.com/pennsieve/pennsieve-go-core/pkg/core"
	"log"
)

type OrganizationStorage struct {
	OrganizationId int64 `json:"organization_id"`
	Size           int64 `json:"size"`
}

// Increment increases the storage associated with the provided organization.
func (d *OrganizationStorage) Increment(db core.PostgresAPI, organizationId int64, size int64) error {

	queryStr := "INSERT INTO pennsieve.organization_storage " +
		"AS organization_storage (organization_id, size) " +
		"VALUES ($1, $2) ON CONFLICT (organization_id) " +
		"DO UPDATE SET size = COALESCE(organization_storage.size, 0) + EXCLUDED.size"

	_, err := db.Exec(queryStr, organizationId, size)
	if err != nil {
		log.Println("Error incrementing package size: ", err)
	}

	return err
}
