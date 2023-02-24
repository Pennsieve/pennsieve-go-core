package pgdb

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/pennsieve/pennsieve-go-core/pkg/pgdb/models"
)

//GetByCognitoId returns a user from Postgress based on his/her cognito-id
//This function also returns the preferred org and whether the user is a super-admin.
func (q *Queries) GetByCognitoId(ctx context.Context, id string) (*models.User, error) {

	queryStr := "SELECT id, node_id, email, first_name, last_name, is_super_admin, preferred_org_id " +
		"FROM pennsieve.users WHERE cognito_id=$1;"

	var user models.User
	row := q.db.QueryRowContext(ctx, queryStr, id)
	err := row.Scan(
		&user.Id,
		&user.NodeId,
		&user.Email,
		&user.FirstName,
		&user.LastName,
		&user.IsSuperAdmin,
		&user.PreferredOrg)

	switch err {
	case sql.ErrNoRows:
		fmt.Println("No rows were returned!")
		return nil, err
	case nil:
		return &user, nil
	default:
		panic(err)
	}
}
