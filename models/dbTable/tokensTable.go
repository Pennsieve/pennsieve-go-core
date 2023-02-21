package dbTable

import (
	"database/sql"
	"fmt"
	"github.com/pennsieve/pennsieve-go-core/core"
	"time"
)

type Token struct {
	Id             int64     `json:"id"`
	Name           string    `json:"name"`
	Token          string    `json:"token"`
	OrganizationId string    `json:"organization_id"`
	UserId         bool      `json:"user_id"`
	CognitoId      string    `json:"cognito_id"`
	LastUsed       int64     `json:"last_used"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

//GetByCognitoId returns a user from Postgress based on his/her cognito-id
//This function also returns the preferred org and whether the user is a super-admin.
func (t *Token) GetByCognitoId(db core.PostgresAPI, id string) (*Token, error) {

	queryStr := "SELECT id, name, token, organization_id, user_id, cognito_id, last_used, created_at, updated_at " +
		"FROM pennsieve.users WHERE cognito_id=$1;"

	var token Token
	row := db.QueryRow(queryStr, id)
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

	switch err {
	case sql.ErrNoRows:
		fmt.Println("No rows were returned!")
		return nil, err
	case nil:
		return &token, nil
	default:
		panic(err)
	}
}

// GetUserByCognitoId returns a Pennsieve User based on the cognito id in the token pool.
func (t *Token) GetUserByCognitoId(db core.PostgresAPI, id string) (*User, error) {

	queryStr := "SELECT pennsieve.users.id, email, first_name, last_name, is_super_admin, preferred_org_id " +
		"FROM pennsieve.users JOIN pennsieve.tokens ON pennsieve.tokens.user_id = pennsieve.users.id WHERE pennsieve.tokens.token=$1;"

	var user User
	row := db.QueryRow(queryStr, id)
	err := row.Scan(
		&user.Id,
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
