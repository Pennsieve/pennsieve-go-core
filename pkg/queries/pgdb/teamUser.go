package pgdb

import (
	"context"
	"database/sql"
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
