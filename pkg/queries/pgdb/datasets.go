package pgdb

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/pennsieve/pennsieve-go-core/pkg/models/dataset"
	"github.com/pennsieve/pennsieve-go-core/pkg/models/dataset/role"
	"github.com/pennsieve/pennsieve-go-core/pkg/models/dataset/state"
	"github.com/pennsieve/pennsieve-go-core/pkg/models/nodeId"
	"github.com/pennsieve/pennsieve-go-core/pkg/models/pgdb"
	log "github.com/sirupsen/logrus"
	"sort"
	"strings"
	"time"
)

type DatasetNotFoundError struct {
	ErrorMessage string
}

func (e DatasetNotFoundError) Error() string {
	return fmt.Sprintf("dataset was not found (error: %v)", e.ErrorMessage)
}

type DatasetUserNotFoundError struct {
	ErrorMessage string
}

func (e DatasetUserNotFoundError) Error() string {
	return fmt.Sprintf("dataset user was not found (error: %v)", e.ErrorMessage)
}

type CreateDatasetParams struct {
	Name                         string
	Description                  string
	Status                       *pgdb.DatasetStatus
	AutomaticallyProcessPackages bool
	License                      string
	Tags                         []string
	DataUseAgreement             *pgdb.DataUseAgreement
}

func (q *Queries) CreateDataset(ctx context.Context, p CreateDatasetParams) (*pgdb.Dataset, error) {
	var err error

	if p.Name == "" {
		return nil, fmt.Errorf("dataset name cannot be empty or null")
	}

	if len(p.Name) > 255 {
		return nil, fmt.Errorf("dataset name cannot exceed 255 characters")
	}

	_, err = q.GetDatasetByName(ctx, p.Name)
	if err != nil {
		switch err.(type) {
		case DatasetNotFoundError:
			// do nothing
		default:
			return nil, fmt.Errorf("a dataset with the name \"%s\" already exists (error: %v)", p.Name, err)
		}
	}

	statement := fmt.Sprintf("INSERT INTO datasets (name, node_id, state, description, automatically_process_packages," +
		" status_id, license, tags, data_use_agreement_id)" +
		" VALUES($1, $2, $3, $4, $5, $6, NULLIF($7, ''), $8, $9);")

	_, err = q.db.ExecContext(ctx, statement,
		p.Name,
		nodeId.NodeId(nodeId.DataSetCode),
		state.READY,
		p.Description,
		p.AutomaticallyProcessPackages,
		p.Status.Id,
		p.License,
		fmt.Sprintf("{%s}", strings.Join(p.Tags, ",")),
		p.DataUseAgreement.Id)

	if err != nil {
		return nil, fmt.Errorf(fmt.Sprintf("database error on insert: %v", err))
	}

	dataset, err := q.GetDatasetByName(ctx, p.Name)
	if err != nil {
		return nil, fmt.Errorf(fmt.Sprintf("database error on query: %v", err))
	}

	return dataset, nil
}

// GetDatasetByName will query workspace datasets by name and return one if found.
func (q *Queries) GetDatasetByName(ctx context.Context, name string) (*pgdb.Dataset, error) {
	query := fmt.Sprintf("SELECT id, name, state, description, updated_at, created_at, node_id,"+
		" permission_bit, type, role, status, automatically_process_packages, license, tags, contributors,"+
		" banner_id, readme_id, status_id, publication_status_id, size, etag, data_use_agreement_id, changelog_id"+
		" FROM datasets WHERE name='%s';", name)
	row := q.db.QueryRowContext(ctx, query)
	return scanDataset(row)
}

// GetDatasets returns all rows in the Upload Record Table
func (q *Queries) GetDatasets(ctx context.Context, organizationId int) ([]pgdb.Dataset, error) {
	queryStr := "SELECT (name, state) FROM datasets"

	rows, err := q.db.QueryContext(ctx, queryStr)
	var allDatasets []pgdb.Dataset
	if err == nil {
		for rows.Next() {
			var currentRecord pgdb.Dataset
			err = rows.Scan(
				&currentRecord.Name,
				&currentRecord.State)

			if err != nil {
				log.Println("ERROR: ", err)
			}

			allDatasets = append(allDatasets, currentRecord)
		}
		return allDatasets, err
	}
	return allDatasets, err
}

// GetDatasetClaim returns the highest role that the user has for a given dataset.
// This method checks the roles of the dataset, the teams, and the specific user roles.
func (q *Queries) GetDatasetClaim(ctx context.Context, user *pgdb.User, datasetNodeId string, organizationId int64) (*dataset.Claim, error) {

	// if user is super-admin
	if user.IsSuperAdmin {
		// USER IS A SUPER-ADMIN

		//TODO: HANDLE SPECIAL CASE
		log.Warn("Not handling super-user authorization at this point.")

	}

	// 1. Get Dataset Role and integer ID
	datasetQuery := fmt.Sprintf("SELECT id, role FROM \"%d\".datasets WHERE node_id='%s';", organizationId, datasetNodeId)

	var datasetId int64
	var maybeDatasetRole sql.NullString

	row := q.db.QueryRowContext(ctx, datasetQuery)
	err := row.Scan(
		&datasetId,
		&maybeDatasetRole)

	if err != nil {
		switch err {
		case sql.ErrNoRows:
			log.Error("No rows were returned!")
			return nil, err
		default:
			log.Error("Uknown Error while scanning dataset table: ", err)
			panic(err)
		}
	}

	// If maybeDatasetRole is set, include the role, otherwise use none-role
	datasetRole := role.None
	if maybeDatasetRole.Valid {
		var ok bool
		datasetRole, ok = role.RoleFromString(maybeDatasetRole.String)
		if !ok {
			log.Fatalln("Could not map Dataset Role from database string: ", maybeDatasetRole.String)
		}
	}

	// 2. Get Team Role
	teamPermission := fmt.Sprintf("\"%d\".dataset_team.role", organizationId)
	datasetTeam := fmt.Sprintf("\"%d\".dataset_team", organizationId)
	teamQueryStr := fmt.Sprintf("SELECT %s FROM pennsieve.team_user JOIN %s "+
		"ON pennsieve.team_user.team_id = %s.team_id "+
		"WHERE user_id=%d AND dataset_id=%d", teamPermission, datasetTeam, datasetTeam, user.Id, datasetId)

	// Get User Role
	userPermission := fmt.Sprintf("\"%d\".dataset_user.role", organizationId)
	datasetUser := fmt.Sprintf("\"%d\".dataset_user", organizationId)
	userQueryStr := fmt.Sprintf("SELECT %s FROM %s WHERE user_id=%d AND dataset_id=%d",
		userPermission, datasetUser, user.Id, datasetId)

	// Combine all queries in a single Union.
	fullQuery := teamQueryStr + " UNION " + userQueryStr + ";"

	rows, err := q.db.QueryContext(ctx, fullQuery)
	if err != nil {
		return nil, err
	}

	roles := []role.Role{
		datasetRole,
	}
	for rows.Next() {
		var roleString string
		err = rows.Scan(
			&roleString)

		if err != nil {
			log.Error("ERROR: ", err)
		}

		role, ok := role.RoleFromString(roleString)
		if !ok {
			log.Fatalln("Could not map Dataset Role from database string.")
		}
		roles = append(roles, role)
	}

	// Sort roles by enum value --> first entry is the highest level of permission.
	sort.Slice(roles, func(i, j int) bool {
		return roles[i] > roles[j]
	})

	// return the maximum role that the user has.
	claim := dataset.Claim{
		Role:   roles[0],
		NodeId: datasetNodeId,
		IntId:  datasetId,
	}

	return &claim, nil

}

func (q *Queries) GetDatasetUser(ctx context.Context, dataset *pgdb.Dataset, user *pgdb.User) (*pgdb.DatasetUser, error) {
	query := "SELECT dataset_id, user_id, role, permission_bit, created_at, updated_at FROM dataset_user WHERE dataset_id=$1 AND user_id=$2"

	var datasetUser pgdb.DatasetUser

	err := q.db.QueryRowContext(ctx, query, dataset.Id, user.Id).Scan(
		&datasetUser.DatasetId,
		&datasetUser.UserId,
		&datasetUser.Role,
		&datasetUser.PermissionBit,
		&datasetUser.CreatedAt,
		&datasetUser.UpdatedAt,
	)

	if err != nil {
		switch err {
		case sql.ErrNoRows:
			return nil, DatasetUserNotFoundError{fmt.Sprintf("%+v", err)}
		default:
			log.Error("Unknown Error while query/scan dataset user table: ", err)
			panic(err)
		}
	}

	return &datasetUser, nil
}

func (q *Queries) AddDatasetUser(ctx context.Context, dataset *pgdb.Dataset, user *pgdb.User, role role.Role) (*pgdb.DatasetUser, error) {
	existing, err := q.GetDatasetUser(ctx, dataset, user)
	if err != nil {
		switch err.(type) {
		case DatasetUserNotFoundError:
			// do nothing
		default:
			return nil, err
		}
	}

	if existing != nil {
		return existing, nil
	}

	statement := "INSERT INTO dataset_user (dataset_id, user_id, role, permission_bit) VALUES ($1, $2, $3, $4)"
	_, err = q.db.ExecContext(ctx, statement, dataset.Id, user.Id, strings.ToLower(role.String()), datasetRoleToPermission(role))
	if err != nil {
		return nil, err
	}

	return q.GetDatasetUser(ctx, dataset, user)
}

func (q *Queries) SetUpdatedAt(ctx context.Context, dataset *pgdb.Dataset, t time.Time) error {
	queryStr := fmt.Sprintf("UPDATE datasets SET updated_at=$1 WHERE id=$2;")
	result, err := q.db.ExecContext(ctx, queryStr, t, dataset.Id)

	msg := ""
	if err != nil {
		msg = fmt.Sprintf("Error updating the updated_at column: %v", err)
		log.Println(msg)
		return err
	}

	affectedRows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if affectedRows != 1 {
		if affectedRows == 0 {
			nofFoundError := &pgdb.ErrFileNotFound{}
			log.Println(nofFoundError.Error())
			return nofFoundError
		}

		multipleRowError := &pgdb.ErrMultipleRowsAffected{}
		log.Println(multipleRowError.Error())
		return multipleRowError
	}

	return nil

}

func datasetRoleToPermission(r role.Role) pgdb.DbPermission {
	switch r {
	case role.None:
		return pgdb.NoPermission
	case role.Viewer:
		return pgdb.Read
	case role.Editor:
		return pgdb.Delete
	case role.Manager:
		return pgdb.Administer
	case role.Owner:
		return pgdb.Owner
	default:
		return pgdb.NoPermission
	}
}

func scanDataset(row *sql.Row) (*pgdb.Dataset, error) {
	var dataset pgdb.Dataset

	err := row.Scan(
		&dataset.Id,
		&dataset.Name,
		&dataset.State,
		&dataset.Description,
		&dataset.UpdatedAt,
		&dataset.CreatedAt,
		&dataset.NodeId,
		&dataset.PermissionBit,
		&dataset.Type,
		&dataset.Role,
		&dataset.Status,
		&dataset.AutomaticallyProcessPackages,
		&dataset.License,
		&dataset.Tags,
		&dataset.Contributors,
		&dataset.BannerId,
		&dataset.ReadmeId,
		&dataset.StatusId,
		&dataset.PublicationStatusId,
		&dataset.Size,
		&dataset.ETag,
		&dataset.DataUseAgreementId,
		&dataset.ChangelogId,
	)

	if err != nil {
		switch err {
		case sql.ErrNoRows:
			return nil, DatasetNotFoundError{"No rows were returned!"}
		default:
			log.Error("Unknown Error while scanning dataset row: ", err)
			panic(err)
		}
	}

	return &dataset, nil
}
