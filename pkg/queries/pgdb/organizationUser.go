package pgdb

import (
	"context"
	"database/sql"
	"fmt"
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
		//log.Error("No rows were returned!")
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

	currentOrgUser, err := q.GetOrganizationUser(ctx, organizationId, userId)
	if err != nil {
		log.Error("Unable to check Org User: ", err)
		return nil, err
	}

	org, err := q.GetOrganization(ctx, organizationId)
	if err != nil {
		log.Error("Unable to check Organization: ", err)
		return nil, err
	}

	enabledFeatures, err := q.GetEnabledFeatureFlags(ctx, organizationId)
	if err != nil {
		log.Error("Unable to check Feature Flags: ", err)
		return nil, err
	}

	orgRole := organization.Claim{
		Role:            currentOrgUser.DbPermission,
		IntId:           organizationId,
		NodeId:          org.NodeId,
		EnabledFeatures: enabledFeatures,
	}

	return &orgRole, nil

}

// nullableFeatureFlag is a temp struct used to hold results from GetOrganizationClaimByNodeId which
// will return rows with null values for featureFlag columns if the org has no flags set
type nullableFeatureFlag struct {
	feature   sql.NullString
	enabled   sql.NullBool
	createdAt sql.NullTime
	updatedAt sql.NullTime
}

func (n nullableFeatureFlag) Valid() bool {
	return n.feature.Valid && n.enabled.Valid && n.createdAt.Valid && n.updatedAt.Valid
}

func (n nullableFeatureFlag) toFeatureFlag(organizationId int64) pgdb.FeatureFlags {
	return pgdb.FeatureFlags{
		OrganizationId: organizationId,
		Feature:        n.feature.String,
		Enabled:        n.enabled.Bool,
		CreatedAt:      n.createdAt.Time,
		UpdatedAt:      n.updatedAt.Time,
	}
}

func (q *Queries) GetOrganizationClaimByNodeId(ctx context.Context, userId int64, organizationNodeId string) (*organization.Claim, error) {
	query := `SELECT o.id, ou.permission_bit, f.feature, f.enabled, f.created_at, f.updated_at 
			  FROM pennsieve.users u JOIN pennsieve.organization_user ou ON u.id = ou.user_id
         			                 JOIN pennsieve.organizations o ON ou.organization_id = o.id
         			            LEFT JOIN pennsieve.feature_flags f ON o.id = f.organization_id and f.enabled = true
			  WHERE u.id = $1 AND o.node_id = $2`

	rows, err := q.db.QueryContext(ctx, query, userId, organizationNodeId)
	if err != nil {
		return nil, fmt.Errorf("error getting organization claim by node id: %w", err)
	}
	defer func() {
		if err := rows.Close(); err != nil {
			log.Warn("error closing rows for get organization claim by node id", err)
		}
	}()

	// orgId and orgPerms may get multiple scans below (if the org has >1 feature flag) but should always be the same value. Not
	// sure how else to handle this except with the redundant scans.
	var orgId int64
	var orgPerms pgdb.DbPermission
	var flags []pgdb.FeatureFlags
	for rows.Next() {
		var nullableFlag nullableFeatureFlag
		if err := rows.Scan(
			&orgId,
			&orgPerms,
			&nullableFlag.feature,
			&nullableFlag.enabled,
			&nullableFlag.createdAt,
			&nullableFlag.updatedAt); err != nil {
			return nil, fmt.Errorf("error reading row for get organization claim by node id: %w", err)
		}
		// an org may have no feature flags. don't add a bunch of zero-ed structs to claim
		if nullableFlag.Valid() {
			featureFlag := nullableFlag.toFeatureFlag(orgId)
			flags = append(flags, featureFlag)
		}
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("row error on get organization claim by node id: %w", err)
	}
	// zero orgId means no rows returned.
	if orgId == 0 {
		return nil, OrganizationUserNotFoundError{fmt.Sprintf("user id: %d, organization node id: %s", userId, organizationNodeId)}
	}

	return &organization.Claim{
		Role:            orgPerms,
		IntId:           orgId,
		NodeId:          organizationNodeId,
		EnabledFeatures: flags,
	}, nil
}
