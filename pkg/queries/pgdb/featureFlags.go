package pgdb

import (
	"context"
	"github.com/pennsieve/pennsieve-go-core/pkg/models/pgdb"
	log "github.com/sirupsen/logrus"
)

// GetFeatureFlags returns all rows in the FeatureFlags Table
func (q *Queries) GetFeatureFlags(ctx context.Context, organizationId int64) ([]pgdb.FeatureFlags, error) {
	queryStr := "SELECT organization_id, feature, enabled,created_at, updated_at FROM pennsieve.feature_flags WHERE organization_id=$1; "

	rows, err := q.db.QueryContext(ctx, queryStr, organizationId)
	var allFeatureFlags []pgdb.FeatureFlags
	if err == nil {
		for rows.Next() {
			var currentRecord pgdb.FeatureFlags
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
