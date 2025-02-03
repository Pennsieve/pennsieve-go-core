package pgdb

import (
	"context"
	"github.com/pennsieve/pennsieve-go-core/pkg/models/pgdb"
)

// GetTokenByCognitoId returns a user from Postgress based on his/her cognito-id
// This function also returns the preferred org and whether the user is a super-admin.
// Returns (nil, sql.ErrNoRows) if no user with the given cognito id exists.
func (q *Queries) GetTokenByCognitoId(ctx context.Context, id string) (*pgdb.Token, error) {

	queryStr := "SELECT id, name, token, organization_id, user_id, cognito_id, last_used, created_at, updated_at " +
		"FROM pennsieve.users WHERE cognito_id=$1;"

	var token pgdb.Token
	row := q.db.QueryRowContext(ctx, queryStr, id)
	err := row.Scan(
		&token.Id,
		&token.Name,
		&token.Token,
		&token.OrganizationId,
		&token.UserId,
		&token.CognitoId,
		&token.LastUsed,
		&token.CreatedAt,
		&token.UpdatedAt)

	if err != nil {
		return nil, err
	}
	return &token, nil
}

// GetUserByCognitoId returns a Pennsieve User based on the cognito id in the token pool.
// Returns (nil, sql.ErrNoRows) if no user with the given token exists
func (q *Queries) GetUserByCognitoId(ctx context.Context, id string) (*pgdb.User, error) {

	queryStr := "SELECT pennsieve.users.id, pennsieve.users.node_id, email, first_name, last_name, is_super_admin, pennsieve.tokens.organization_id as preferred_org_id " +
		"FROM pennsieve.users JOIN pennsieve.tokens ON pennsieve.tokens.user_id = pennsieve.users.id WHERE pennsieve.tokens.token=$1;"

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
