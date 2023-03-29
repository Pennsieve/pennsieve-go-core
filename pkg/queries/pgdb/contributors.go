package pgdb

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/pennsieve/pennsieve-go-core/pkg/models/pgdb"
	log "github.com/sirupsen/logrus"
	"strings"
)

type ContributorNotFoundError struct {
	err error
}

func (e ContributorNotFoundError) Error() string {
	return fmt.Sprintf("contributor was not found (error: %v)", e.err)
}

type NewContributor struct {
	FirstName     string
	MiddleInitial string
	LastName      string
	Degree        string
	EmailAddress  string
	Orcid         string
	UserId        int64
}

func ContributorColumns() []string {
	columns := []string{
		"id",
		"first_name",
		"last_name",
		"email",
		"orcid",
		"user_id",
		"updated_at",
		"created_at",
		"middle_initial",
		"degree"}
	return columns
}

func ReadContributorColumns() string {
	return strings.Join(ContributorColumns(), ",")
}

func WriteContributorColumns() string {
	// TODO: filter out "id", "updated_at", "created_at"
	return strings.Join(ContributorColumns(), ",")
}

func scanContributor(row *sql.Row) (*pgdb.Contributor, error) {
	var contributor pgdb.Contributor

	err := row.Scan(
		&contributor.Id,
		&contributor.FirstName,
		&contributor.LastName,
		&contributor.Email,
		&contributor.Orcid,
		&contributor.UserId,
		&contributor.UpdatedAt,
		&contributor.CreatedAt,
		&contributor.MiddleInitial,
		&contributor.Degree)

	if err != nil {
		switch err {
		case sql.ErrNoRows:
			return nil, ContributorNotFoundError{err}
		default:
			log.Error("Unknown Error while scanning dataset row: ", err)
			panic(err)
		}
	}
	return &contributor, nil
}

func getContributor(ctx context.Context, db DBTX, query string) (*pgdb.Contributor, error) {
	return scanContributor(db.QueryRowContext(ctx, query))
}

// AddContributor will add a Contributor to the Organization's Contributors table.
func (q *Queries) AddContributor(ctx context.Context, contributor NewContributor) (*pgdb.Contributor, error) {
	var err error
	if contributor.UserId > 0 {
		_, err = q.db.ExecContext(ctx,
			"INSERT INTO contributors (first_name, middle_initial, last_name, degree, email, orcid, user_id) VALUES($1, $2, $3, $4, $5, $6, $7)",
			contributor.FirstName,
			contributor.MiddleInitial,
			contributor.LastName,
			contributor.Degree,
			contributor.EmailAddress,
			contributor.Orcid,
			contributor.UserId)
	} else {
		_, err = q.db.ExecContext(ctx,
			"INSERT INTO contributors (first_name, middle_initial, last_name, degree, email, orcid) VALUES($1, $2, $3, $4, $5, $6)",
			contributor.FirstName,
			contributor.MiddleInitial,
			contributor.LastName,
			contributor.Degree,
			contributor.EmailAddress,
			contributor.Orcid)
	}

	// ExecContext returned an error
	if err != nil {
		return nil, err
	}

	return q.GetContributorByEmail(ctx, contributor.EmailAddress)
}

// GetContributor will get a Contributor by the Contributor Id (not the User Id).
func (q *Queries) GetContributor(ctx context.Context, id int64) (*pgdb.Contributor, error) {
	query := fmt.Sprintf("SELECT %s FROM contributors WHERE id=%d", ReadContributorColumns(), id)
	return getContributor(ctx, q.db, query)
}

// GetContributorByUserId will get a Contributor by User Id.
func (q *Queries) GetContributorByUserId(ctx context.Context, userId int64) (*pgdb.Contributor, error) {
	query := fmt.Sprintf("SELECT %s FROM contributors WHERE user_id=%d", ReadContributorColumns(), userId)
	return getContributor(ctx, q.db, query)
}

// GetContributorByEmail will get a Contributor by Email Address.
func (q *Queries) GetContributorByEmail(ctx context.Context, email string) (*pgdb.Contributor, error) {
	query := fmt.Sprintf("SELECT %s FROM contributors WHERE email='%s'", ReadContributorColumns(), email)
	return getContributor(ctx, q.db, query)
}

// GetContributorByOrcid will get a Contributor by ORCID iD.
func (q *Queries) GetContributorByOrcid(ctx context.Context, orcid string) (*pgdb.Contributor, error) {
	query := fmt.Sprintf("SELECT %s FROM contributors WHERE orcid='%s'", ReadContributorColumns(), orcid)
	return getContributor(ctx, q.db, query)
}
