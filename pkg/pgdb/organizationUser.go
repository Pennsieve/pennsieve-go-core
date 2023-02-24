package pgdb

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/pennsieve/pennsieve-go-core/pkg/pgdb/models"
)

func (q *Queries) GetOrganizationUserById(ctx context.Context, id int64) (*models.OrganizationUser, error) {

	queryStr := "SELECT organization_id, user_id, permission_bit, created_at, updated_at " +
		"FROM pennsieve.organization_user WHERE user_id=$1;"

	var orgUser models.OrganizationUser
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
