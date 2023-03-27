package pgdb

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/pennsieve/pennsieve-go-core/pkg/models/organization"
	"github.com/pennsieve/pennsieve-go-core/pkg/models/pgdb"
	log "github.com/sirupsen/logrus"
)

func (q *Queries) GetOrganizationUserById(ctx context.Context, id int64) (*pgdb.OrganizationUser, error) {

	queryStr := "SELECT organization_id, user_id, permission_bit, created_at, updated_at " +
		"FROM pennsieve.organization_user WHERE user_id=$1;"

	var orgUser pgdb.OrganizationUser
	row := q.db.QueryRowContext(ctx, queryStr, id)
	err := row.Scan(
		&orgUser.OrganizationId,
		&orgUser.UserId,
		&orgUser.DbPermission,
		&orgUser.CreatedAt,
		&orgUser.UpdatedAt)

	switch err {
	case sql.ErrNoRows:
		fmt.Println("No rows were returned!")
		return nil, err
	case nil:
		return &orgUser, nil
	default:
		panic(err)
	}
}

func (q *Queries) GetOrganizationUser(ctx context.Context, orgId int64, userId int64) (*pgdb.OrganizationUser, error) {
	queryStr := "SELECT organization_id, user_id, permission_bit, created_at, updated_at " +
		"FROM pennsieve.organization_user WHERE organization_id=$1 AND user_id=$2;"

	var orgUser pgdb.OrganizationUser
	row := q.db.QueryRowContext(ctx, queryStr, orgId, userId)
	err := row.Scan(
		&orgUser.OrganizationId,
		&orgUser.UserId,
		&orgUser.DbPermission,
		&orgUser.CreatedAt,
		&orgUser.UpdatedAt)

	switch err {
	case sql.ErrNoRows:
		fmt.Println("No rows were returned!")
		return nil, err
	case nil:
		return &orgUser, nil
	default:
		panic(err)
	}
}

// GetOrganizationClaim returns an organization claim for a specific user.
func (q *Queries) GetOrganizationClaim(ctx context.Context, userId int64, organizationId int64) (*organization.Claim, error) {

	currentOrgUser, err := q.GetOrganizationUserById(ctx, userId)
	if err != nil {
		log.Error("Unable to check Org User: ", err)
		return nil, err
	}

	allFeatures, err := q.GetFeatureFlags(ctx, organizationId)
	if err != nil {
		log.Error("Unable to check Feature Flags: ", err)
		return nil, err
	}

	orgRole := organization.Claim{
		Role:            currentOrgUser.DbPermission,
		IntId:           organizationId,
		EnabledFeatures: allFeatures,
	}

	return &orgRole, nil

}
