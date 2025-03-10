package pgdb

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/pennsieve/pennsieve-go-core/pkg/models/pgdb"
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
			return nil, err
		}
	}
	return &contributor, nil
}

func getContributor(ctx context.Context, db DBTX, query string) (*pgdb.Contributor, error) {
	return scanContributor(db.QueryRowContext(ctx, query))
}

// AddContributor will add a Contributor to the Organization's Contributors table.
func (q *Queries) AddContributor(ctx context.Context, newContributor NewContributor) (*pgdb.Contributor, error) {
	var err error
	var contributor *pgdb.Contributor

	// try to find the contributor; if the contributor is found, then return it
	contributor, err = q.FindContributor(ctx, newContributor)
	if contributor != nil {
		return contributor, nil
	}

	// the contributor does not exist, so create it
	if newContributor.UserId > 0 {
		_, err = q.db.ExecContext(ctx,
			"INSERT INTO contributors (first_name, middle_initial, last_name, degree, email, orcid, user_id) VALUES($1, $2, $3, NULLIF($4, ''), $5, $6, $7)",
			newContributor.FirstName,
			newContributor.MiddleInitial,
			newContributor.LastName,
			newContributor.Degree,
			newContributor.EmailAddress,
			newContributor.Orcid,
			newContributor.UserId)
	} else {
		_, err = q.db.ExecContext(ctx,
			"INSERT INTO contributors (first_name, middle_initial, last_name, degree, email, orcid) VALUES($1, $2, $3, NULLIF($4, ''), $5, $6)",
			newContributor.FirstName,
			newContributor.MiddleInitial,
			newContributor.LastName,
			newContributor.Degree,
			newContributor.EmailAddress,
			newContributor.Orcid)
	}

	// ExecContext returned an error
	if err != nil {
		return nil, err
	}

	return q.GetContributorByEmail(ctx, newContributor.EmailAddress)
}

// FindContributor will search for a contributor by several User Id, Email Address, and ORCID
func (q *Queries) FindContributor(ctx context.Context, search NewContributor) (*pgdb.Contributor, error) {
	var err error
	var contributor *pgdb.Contributor
	contributor = nil

	if contributor == nil && search.UserId > 0 {
		contributor, err = q.GetContributorByUserId(ctx, search.UserId)
	}

	if contributor == nil && search.EmailAddress != "" {
		contributor, err = q.GetContributorByEmail(ctx, search.EmailAddress)
	}

	if contributor == nil && search.Orcid != "" {
		contributor, err = q.GetContributorByOrcid(ctx, search.Orcid)
	}

	return contributor, err
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
