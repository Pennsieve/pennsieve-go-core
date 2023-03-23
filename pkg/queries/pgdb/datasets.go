package pgdb

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/pennsieve/pennsieve-go-core/pkg/models/dataset"
	"github.com/pennsieve/pennsieve-go-core/pkg/models/pgdb"
	log "github.com/sirupsen/logrus"
	"sort"
)

type CreateDatasetParams struct {
	name                         string
	description                  string
	datasetState                 string
	status                       pgdb.DatasetStatus
	automaticallyProcessPackages bool
	license                      string
	tags                         []string
	dataUseAgreement             pgdb.DataUseAgreement
}

func (q *Queries) CreateDataset(p CreateDatasetParams) (*pgdb.Dataset, error) {
	if p.name == "" {
		return nil, fmt.Errorf("dataset name cannot be empty or null")
	}

	if len(p.name) > 255 {
		return nil, fmt.Errorf("dataset name cannot exceed 255 characters")
	}

	_, err := q.GetDatasetByName(context.TODO(), 0, p.name)
	if err != nil {
		return nil, fmt.Errorf("a dataset with the name already exists")
	}

	return nil, nil
}

// GetDatasetByName will query workspace datasets by name and return one if found.
func (q *Queries) GetDatasetByName(ctx context.Context, organizationId int, name string) (*pgdb.Dataset, error) {
	query := fmt.Sprintf("SELECT id, name, state, description, updated_at, created_at, node_id,"+
		" permission_bit, type, role, status, automatically_process_packages, license, tags, contributors,"+
		" banner_id, readme_id, status_id, publication_status_id, size, etag, data_use_agreement_id, changelog_id"+
		" FROM \"%d\".datasets WHERE name='%s';", organizationId, name)
	row := q.db.QueryRowContext(ctx, query)
	return rowToDataset(row)
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
	datasetRole := dataset.None
	if maybeDatasetRole.Valid {
		var ok bool
		datasetRole, ok = dataset.RoleFromString(maybeDatasetRole.String)
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

	roles := []dataset.Role{
		datasetRole,
	}
	for rows.Next() {
		var roleString string
		err = rows.Scan(
			&roleString)

		if err != nil {
			log.Error("ERROR: ", err)
		}

		role, ok := dataset.RoleFromString(roleString)
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

func rowToDataset(row *sql.Row) (*pgdb.Dataset, error) {
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
			log.Error("No rows were returned!")
			return nil, err
		default:
			log.Error("Unknown Error while scanning dataset row: ", err)
			panic(err)
		}
	}

	return &dataset, nil
}
