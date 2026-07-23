package pgdb

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/pennsieve/pennsieve-go-core/pkg/models/pgdb"
	"github.com/pennsieve/pennsieve-go-core/pkg/models/teamUser"
	log "github.com/sirupsen/logrus"
)

type UserTeamMembership struct {
	UserId            int64
	UserEmail         string
	UserNodeId        string
	OrgId             int64
	OrgName           string
	OrgNodeId         string
	OrgUserPermission pgdb.DbPermission
	TeamId            int64
	TeamName          string
	TeamNodeId        string
	TeamPermission    pgdb.DbPermission
	TeamType          sql.NullString
}

// Deprecated: GetTeamMemberships resolves teams using the user's preferred_org_id, which is
// mutable server-side state and is wrong for org/dataset-scoped auth (it returns the preferred
// org's team memberships regardless of which org the request targets). Use
// GetTeamMembershipsForOrg with the organization resolved from the request instead.
func (q *Queries) GetTeamMemberships(ctx context.Context, userId int64) ([]UserTeamMembership, error) {
	query := "select " +
		"  u.id as user_id, " +
		"  u.email as user_email, " +
		"  u.node_id as user_node_id, " +
		"  o.id as org_id, " +
		"  o.name as org_name, " +
		"  o.node_id as org_node_id, " +
		"  ou.permission_bit as org_permission_bit, " +
		"  t.id as team_id, " +
		"  t.name as team_name, " +
		"  t.node_id as team_node_id, " +
		"  ot.permission_bit as team_permission_bit, " +
		"  ot.system_team_type as system_team_type " +
		"from pennsieve.users u " +
		"join pennsieve.organization_user ou on ou.user_id=u.id " +
		"join pennsieve.organizations o on ou.organization_id=o.id " +
		"left join pennsieve.organization_team ot on o.id=ot.organization_id " +
		"left join pennsieve.teams t on ot.team_id=t.id " +
		"join pennsieve.team_user tu on tu.team_id=t.id and tu.user_id=u.id " +
		"where u.id=$1 " +
		"  and o.id=u.preferred_org_id;"

	// exec query
	rows, err := q.db.QueryContext(ctx, query, userId)
	if err != nil {
		log.Error(fmt.Sprintf("error querying for user team memberships (error: %+v)", err))
		return nil, err
	}

	// iterate over rows, scan to struct
	var userTeamMemberships []UserTeamMembership
	for rows.Next() {
		var utm UserTeamMembership
		err = rows.Scan(
			&utm.UserId,
			&utm.UserEmail,
			&utm.UserNodeId,
			&utm.OrgId,
			&utm.OrgName,
			&utm.OrgNodeId,
			&utm.OrgUserPermission,
			&utm.TeamId,
			&utm.TeamName,
			&utm.TeamNodeId,
			&utm.TeamPermission,
			&utm.TeamType,
		)
		if err != nil {
			log.Error(fmt.Sprintf("error scanning user team membership row (error: %+v)", err))
		}
		userTeamMemberships = append(userTeamMemberships, utm)
	}

	if err != nil {
		log.Error(fmt.Sprintf("unable to check user team membership (error: %+v)", err))
		return nil, err
	}

	return userTeamMemberships, nil
}

// GetTeamMembershipsForOrg returns the user's team memberships within the given organization.
// Unlike the deprecated GetTeamMemberships, the organization is supplied explicitly (resolved
// from the request) rather than read from the user's mutable preferred_org_id.
func (q *Queries) GetTeamMembershipsForOrg(ctx context.Context, userId int64, organizationId int64) ([]UserTeamMembership, error) {
	query := "select " +
		"  u.id as user_id, " +
		"  u.email as user_email, " +
		"  u.node_id as user_node_id, " +
		"  o.id as org_id, " +
		"  o.name as org_name, " +
		"  o.node_id as org_node_id, " +
		"  ou.permission_bit as org_permission_bit, " +
		"  t.id as team_id, " +
		"  t.name as team_name, " +
		"  t.node_id as team_node_id, " +
		"  ot.permission_bit as team_permission_bit, " +
		"  ot.system_team_type as system_team_type " +
		"from pennsieve.users u " +
		"join pennsieve.organization_user ou on ou.user_id=u.id " +
		"join pennsieve.organizations o on ou.organization_id=o.id " +
		"left join pennsieve.organization_team ot on o.id=ot.organization_id " +
		"left join pennsieve.teams t on ot.team_id=t.id " +
		"join pennsieve.team_user tu on tu.team_id=t.id and tu.user_id=u.id " +
		"where u.id=$1 " +
		"  and o.id=$2;"

	// exec query
	rows, err := q.db.QueryContext(ctx, query, userId, organizationId)
	if err != nil {
		log.Error(fmt.Sprintf("error querying for user team memberships (error: %+v)", err))
		return nil, err
	}

	// iterate over rows, scan to struct
	var userTeamMemberships []UserTeamMembership
	for rows.Next() {
		var utm UserTeamMembership
		err = rows.Scan(
			&utm.UserId,
			&utm.UserEmail,
			&utm.UserNodeId,
			&utm.OrgId,
			&utm.OrgName,
			&utm.OrgNodeId,
			&utm.OrgUserPermission,
			&utm.TeamId,
			&utm.TeamName,
			&utm.TeamNodeId,
			&utm.TeamPermission,
			&utm.TeamType,
		)
		if err != nil {
			log.Error(fmt.Sprintf("error scanning user team membership row (error: %+v)", err))
		}
		userTeamMemberships = append(userTeamMemberships, utm)
	}

	if err != nil {
		log.Error(fmt.Sprintf("unable to check user team membership (error: %+v)", err))
		return nil, err
	}

	return userTeamMemberships, nil
}

// Deprecated: GetTeamClaims resolves teams using the user's preferred_org_id (via
// GetTeamMemberships), which is mutable server-side state and wrong for org/dataset-scoped auth.
// Use GetTeamClaimsForOrg with the organization resolved from the request instead.
func (q *Queries) GetTeamClaims(ctx context.Context, userId int64) ([]teamUser.Claim, error) {
	userTeamMemberships, err := q.GetTeamMemberships(ctx, userId)
	if err != nil {
		log.Error(fmt.Sprintf("unable to get user team memberships (error: %+v)", err))
		return nil, err
	}

	return teamClaimsFromMemberships(userTeamMemberships), nil
}

// GetTeamClaimsForOrg returns the user's team claims within the given organization. Unlike the
// deprecated GetTeamClaims, the organization is supplied explicitly (resolved from the request)
// rather than read from the user's mutable preferred_org_id.
func (q *Queries) GetTeamClaimsForOrg(ctx context.Context, userId int64, organizationId int64) ([]teamUser.Claim, error) {
	userTeamMemberships, err := q.GetTeamMembershipsForOrg(ctx, userId, organizationId)
	if err != nil {
		log.Error(fmt.Sprintf("unable to get user team memberships (error: %+v)", err))
		return nil, err
	}

	return teamClaimsFromMemberships(userTeamMemberships), nil
}

func teamClaimsFromMemberships(userTeamMemberships []UserTeamMembership) []teamUser.Claim {
	var teamClaims []teamUser.Claim
	for _, membership := range userTeamMemberships {
		var teamType string
		if membership.TeamType.Valid {
			teamType = membership.TeamType.String
		} else {
			teamType = "<none>"
		}
		claim := teamUser.Claim{
			IntId:      membership.TeamId,
			Name:       membership.TeamName,
			NodeId:     membership.TeamNodeId,
			Permission: membership.TeamPermission,
			TeamType:   teamType,
		}

		teamClaims = append(teamClaims, claim)
	}

	return teamClaims
}
