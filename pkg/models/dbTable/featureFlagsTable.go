package dbTable

import (
	"github.com/pennsieve/pennsieve-go-core/pkg/core"
	"log"
	"time"
)

type FeatureFlags struct {
	OrganizationId int64     `json:"organization_id"`
	Feature        string    `json:"feature"`
	Enabled        bool      `json:"enabled"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

// GetAll returns all rows in the FeatureFlags Table
func (d *FeatureFlags) GetAll(db core.PostgresAPI, organizationId int64) ([]FeatureFlags, error) {
	queryStr := "SELECT organization_id, feature, enabled,created_at, updated_at FROM pennsieve.feature_flags WHERE organization_id=$1; "

	rows, err := db.Query(queryStr, organizationId)
	var allFeatureFlags []FeatureFlags
	if err == nil {
		for rows.Next() {
			var currentRecord FeatureFlags
			err = rows.Scan(
				&currentRecord.OrganizationId,
				&currentRecord.Feature,
				&currentRecord.Enabled,
				&currentRecord.CreatedAt,
				&currentRecord.UpdatedAt)

			if err != nil {
				log.Println("ERROR: ", err)
			}

			allFeatureFlags = append(allFeatureFlags, currentRecord)
		}
		return allFeatureFlags, err
	}
	return allFeatureFlags, err
}
