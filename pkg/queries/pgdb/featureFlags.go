package pgdb

import (
	"context"
	"fmt"
	"github.com/pennsieve/pennsieve-go-core/pkg/models/pgdb"
	log "github.com/sirupsen/logrus"
)

const featureFlagBaseQuery = "SELECT organization_id, feature, enabled,created_at, updated_at FROM pennsieve.feature_flags WHERE organization_id=$1"

// queryFeatureFlags assumes that query is a select statement with columns and order as in featureFlagBaseQuery.
func queryFeatureFlags(ctx context.Context, db DBTX, query string, params ...any) ([]pgdb.FeatureFlags, error) {
	rows, err := db.QueryContext(ctx, query, params...)
	if err != nil {
		return nil, fmt.Errorf("error getting feature flags with query %s: %w", query, err)
	}
	defer func() {
		if err := rows.Close(); err != nil {
			log.Warn("error closing feature flag rows after query:", query, "error:", err)
		}
	}()

	var featureFlags []pgdb.FeatureFlags
	for rows.Next() {
		var currentRecord pgdb.FeatureFlags
		if err := rows.Scan(
			&currentRecord.OrganizationId,
			&currentRecord.Feature,
			&currentRecord.Enabled,
			&currentRecord.CreatedAt,
			&currentRecord.UpdatedAt); err != nil {
			return nil, fmt.Errorf("error scanning feature flag rows with query %s: %w", query, err)
		}

		featureFlags = append(featureFlags, currentRecord)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error during feature flag row iteration with query %s: %w", query, err)
	}
	return featureFlags, err

}

// GetFeatureFlags returns all rows in the FeatureFlags Table
func (q *Queries) GetFeatureFlags(ctx context.Context, organizationId int64) ([]pgdb.FeatureFlags, error) {
	return queryFeatureFlags(ctx, q.db, featureFlagBaseQuery, organizationId)
}

func (q *Queries) GetEnabledFeatureFlags(ctx context.Context, organizationId int64) ([]pgdb.FeatureFlags, error) {
	query := fmt.Sprintf("%s AND enabled = true", featureFlagBaseQuery)
	return queryFeatureFlags(ctx, q.db, query, organizationId)

}
