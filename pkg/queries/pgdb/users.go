package pgdb

import (
	"context"
	"github.com/pennsieve/pennsieve-go-core/pkg/models/pgdb"
)

// GetByCognitoId returns a user from Postgress based on his/her cognito-id
// This function also returns the preferred org and whether the user is a super-admin.
// Returns (nil, sql.ErrNoRows) if no user with the given cognito id exists
func (q *Queries) GetByCognitoId(ctx context.Context, id string) (*pgdb.User, error) {

	queryStr := "SELECT id, node_id, email, first_name, last_name, is_super_admin, COALESCE(preferred_org_id, -1) as preferred_org_id " +
		"FROM pennsieve.users WHERE cognito_id=$1;"

	var user pgdb.User
	row := q.db.QueryRowContext(ctx, queryStr, id)
	err := row.Scan(
		&user.Id,
		&user.NodeId,
		&user.Email,
		&user.FirstName,
		&user.LastName,
		&user.IsSuperAdmin,
		&user.PreferredOrg)

	if err != nil {
		return nil, err
	}
	return &user, nil
}

// GetUserById returns a user from Postgres based on the user's int id
// This function also returns the preferred org and whether the user is a super-admin.
// Returns (nil, sql.ErrNoRows) if no user with the given id exists.
func (q *Queries) GetUserById(ctx context.Context, id int64) (*pgdb.User, error) {

	queryStr := "SELECT id, node_id, email, first_name, last_name, is_super_admin, COALESCE(preferred_org_id, -1) as preferred_org_id " +
		"FROM pennsieve.users WHERE id=$1;"

	var user pgdb.User
	row := q.db.QueryRowContext(ctx, queryStr, id)
	err := row.Scan(
		&user.Id,
		&user.NodeId,
		&user.Email,
		&user.FirstName,
		&user.LastName,
		&user.IsSuperAdmin,
		&user.PreferredOrg)

	if err != nil {
		return nil, err
	}
	return &user, nil
}
