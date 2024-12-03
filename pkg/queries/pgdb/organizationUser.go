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

// GetOrganizationClaim returns an organization claim for a specific user given a workspace id
func (q *Queries) GetOrganizationClaim(ctx context.Context, userId int64, organizationId int64) (*organization.Claim, error) {
	return queryOrganizationClaim(ctx, q.db, &orgClaimByOrgId, userId, organizationId)
}

// GetOrganizationClaimByNodeId returns an organization claim for a specific user given a workspace node id
func (q *Queries) GetOrganizationClaimByNodeId(ctx context.Context, userId int64, organizationNodeId string) (*organization.Claim, error) {
	return queryOrganizationClaim(ctx, q.db, &orgClaimByOrgNodeId, userId, organizationNodeId)
}

// nullableFeatureFlag is a temp struct used to hold results from the query used by queryOrganizationClaim which
// will return rows with null values for featureFlag columns if the org has no flags set
type nullableFeatureFlag struct {
	feature   sql.NullString
	enabled   sql.NullBool
	createdAt sql.NullTime
	updatedAt sql.NullTime
}

func (n nullableFeatureFlag) valid() bool {
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

type orgClaimQuery struct {
	name            string
	query           string
	paramsMsgFormat string
}

func (q *orgClaimQuery) String() string {
	return q.name
}

func (q *orgClaimQuery) paramsMsg(userId int64, orgIdentifier any) string {
	return fmt.Sprintf(q.paramsMsgFormat, userId, orgIdentifier)
}

const orgClaimQueryFormat = `SELECT o.id, o.node_id, ou.permission_bit, f.feature, f.enabled, f.created_at, f.updated_at 
			  		         FROM pennsieve.users u JOIN pennsieve.organization_user ou ON u.id = ou.user_id
         			             			        JOIN pennsieve.organizations o ON ou.organization_id = o.id
         			                           LEFT JOIN pennsieve.feature_flags f ON o.id = f.organization_id and f.enabled = true
			                 WHERE u.id = $1 AND %s = $2`

// orgClaimByOrgNodeId is the query to get an org claim given a user id and workspace node id
var orgClaimByOrgNodeId = orgClaimQuery{
	name:            "GetOrganizationClaimByNodeId",
	query:           fmt.Sprintf(orgClaimQueryFormat, "o.node_id"),
	paramsMsgFormat: "user id: %d, workspace node id: %s",
}

// orgClaimByOrgId is the query to get an org claim given a user id and workspace (int) id
var orgClaimByOrgId = orgClaimQuery{
	name:            "GetOrganizationClaim",
	query:           fmt.Sprintf(orgClaimQueryFormat, "o.id"),
	paramsMsgFormat: "user id: %d, workspace id: %d",
}

// queryOrganizationClaim constructs an org claim from the given query and params. It assumes that the given query
// is like orgClaimQueryFormat: it has those columns in that order, and has two positional params. $1 is the user id
// and $2 is the org identifier, either the int or node id of the workspace.
func queryOrganizationClaim(ctx context.Context, db DBTX, orgClaimQuery *orgClaimQuery, userId int64, orgIdentifier any) (*organization.Claim, error) {
	rows, err := db.QueryContext(ctx, orgClaimQuery.query, userId, orgIdentifier)
	if err != nil {
		return nil, fmt.Errorf("error getting organization claim with query %s: %w", orgClaimQuery, err)
	}
	defer func() {
		if err := rows.Close(); err != nil {
			log.Warn("error closing rows for get organization claim with query", orgClaimQuery, "error:", err)
		}
	}()

	// orgId, orgNodeId, and orgPerms may get multiple scans below (if the org has >1 feature flag) but should always be the same value. Not
	// sure how else to handle this except with the redundant scans.
	var orgId int64
	var orgNodeId string
	var orgPerms pgdb.DbPermission
	var flags []pgdb.FeatureFlags
	for rows.Next() {
		var nullableFlag nullableFeatureFlag
		if err := rows.Scan(
			&orgId,
			&orgNodeId,
			&orgPerms,
			&nullableFlag.feature,
			&nullableFlag.enabled,
			&nullableFlag.createdAt,
			&nullableFlag.updatedAt); err != nil {
			return nil, fmt.Errorf("error reading row for get organization claim with query %s: %w", orgClaimQuery, err)
		}
		// an org may have no feature flags. don't add a bunch of zero-ed structs to claim
		if nullableFlag.valid() {
			featureFlag := nullableFlag.toFeatureFlag(orgId)
			flags = append(flags, featureFlag)
		}
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("row error on get organization claim with query %s: %w", orgClaimQuery, err)
	}
	// zero orgId means no rows returned.
	if orgId == 0 {
		return nil, OrganizationUserNotFoundError{orgClaimQuery.paramsMsg(userId, orgIdentifier)}
	}

	return &organization.Claim{
		Role:            orgPerms,
		IntId:           orgId,
		NodeId:          orgNodeId,
		EnabledFeatures: flags,
	}, nil
}
