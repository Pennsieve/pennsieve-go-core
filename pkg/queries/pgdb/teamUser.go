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
			// TODO: handle error on scan
		}
		userTeamMemberships = append(userTeamMemberships, utm)
	}

	if err != nil {
		log.Error("Unable to check User Team Membership (Publishers): ", err)
		return nil, err
	}

	return userTeamMemberships, nil
}

func (q *Queries) GetTeamClaims(ctx context.Context, orgId int64, userId int64) ([]teamUser.Claim, error) {
	userTeamMemberships, err := q.GetTeamMemberships(ctx, userId)
	if err != nil {
		return nil, err
	}

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

	return teamClaims, nil
}

// GetPublishersClaim will return a Claim if the user is on the Publishers team in the organization
// TODO: remove in place of generalized form with "check" method in Authorizer
func (q *Queries) GetPublishersClaim(ctx context.Context, orgId int64, userId int64) (*teamUser.Claim, error) {
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
		"  and o.id=$2 " +
		"  and ot.system_team_type='publishers';"

	var utm UserTeamMembership
	row := q.db.QueryRowContext(ctx, query, userId, orgId)
	err := row.Scan(
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
		log.Error("Unable to check User Team Membership (Publishers): ", err)
		return nil, err
	}

	var teamType string
	if utm.TeamType.Valid {
		teamType = utm.TeamType.String
	} else {
		teamType = "<none>"
	}

	claim := teamUser.Claim{
		IntId:      utm.TeamId,
		Name:       utm.TeamName,
		NodeId:     utm.TeamNodeId,
		Permission: utm.TeamPermission,
		TeamType:   teamType,
	}

	return &claim, nil
}
