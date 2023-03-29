package pgdb

import (
	"context"
	"database/sql"
	"github.com/pennsieve/pennsieve-go-core/pkg/models/organization"
	"github.com/pennsieve/pennsieve-go-core/pkg/models/pgdb"
	log "github.com/sirupsen/logrus"
)

type OrganizationUserNotFoundError struct {
	ErrorMessage string
}

func (e OrganizationUserNotFoundError) Error() string {
	return fmt.Sprintf("organization user was not found (error: %v)", e.ErrorMessage)
}

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
		log.Error("No rows were returned!")
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
		return nil, OrganizationUserNotFoundError{fmt.Sprintf("%+v", err)}
	case nil:
		return &orgUser, nil
	default:
		panic(err)
	}
}

func (q *Queries) AddOrganizationUser(ctx context.Context, orgId int64, userId int64, permBit pgdb.DbPermission) (*pgdb.OrganizationUser, error) {
	var err error

	// check for existing user membership in the organization
	existing, err := q.GetOrganizationUser(ctx, orgId, userId)
	if err != nil {
		switch err.(type) {
		case OrganizationUserNotFoundError:
			// do nothing
		default:
			return nil, err
		}
	}

	// the user is already in the organization, return existing membership (do not update)
	if existing != nil {
		return existing, nil
	}

	statement := "INSERT INTO pennsieve.organization_user (organization_id, user_id, permission_bit) VALUES ($1, $2, $3)"

	_, err = q.db.ExecContext(ctx, statement, orgId, userId, permBit)
	if err != nil {
		return nil, fmt.Errorf(fmt.Sprintf("database error on insert: %v", err))
	}

	orgUser, err := q.GetOrganizationUser(ctx, orgId, userId)
	if err != nil {
		return nil, fmt.Errorf(fmt.Sprintf("database error on query: %v", err))
	}

	return orgUser, nil
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
