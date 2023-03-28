package pgdb

import (
	"database/sql"
	"time"
)

type Contributor struct {
	Id            int64          `json:"id"`
	FirstName     string         `json:"first_name"`
	LastName      string         `json:"last_name"`
	Email         string         `json:"email"`
	Orcid         sql.NullString `json:"orcid"`
	UserId        sql.NullInt64  `json:"user_id"`
	UpdatedAt     time.Time      `json:"updated_at"`
	CreatedAt     time.Time      `json:"created_at"`
	MiddleInitial sql.NullString `json:"middle_initial"`
	Degree        sql.NullString `json:"degree"`
}
