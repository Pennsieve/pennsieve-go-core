package models

import (
	"database/sql"
	"fmt"
	"github.com/pennsieve/pennsieve-go-core/pkg/pgdb"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestDatasetsInsertSelect(t *testing.T) {

	orgId := 3
	db, err := pgdb.ConnectENVWithOrg(3)
	defer func() {
		if db != nil {
			assert.NoError(t, db.Close())
		}
	}()
	assert.NoErrorf(t, err, "could not open DB")

	input := Dataset{
		Id:           1,
		Name:         "Test Dataset",
		State:        "READY",
		Description:  sql.NullString{},
		NodeId:       sql.NullString{String: "N:dataset:1234", Valid: true},
		Role:         sql.NullString{String: "editor", Valid: true},
		Tags:         Tags{"test", "sql"},
		Contributors: Contributors{},
		StatusId:     int32(1),
	}
	_, err = db.Exec("INSERT INTO datasets (id, name, state, description, node_id, role, tags, contributors, status_id) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)", input.Id, input.Name, input.State, input.Description, input.NodeId, input.Role, input.Tags, input.Contributors, input.StatusId)
	defer truncate(t, db, orgId, "datasets")
	if assert.NoError(t, err) {

		countStmt := fmt.Sprintf("SELECT COUNT(*) FROM datasets")
		var count int
		assert.NoError(t, db.QueryRow(countStmt).Scan(&count))
		assert.Equal(t, 1, count)

		var actual Dataset
		err = db.QueryRow("SELECT id, name, state, description, node_id, role, tags, contributors, status_id FROM datasets").Scan(
			&actual.Id,
			&actual.Name,
			&actual.State,
			&actual.Description,
			&actual.NodeId,
			&actual.Role,
			&actual.Tags,
			&actual.Contributors,
			&actual.StatusId)
		if assert.NoError(t, err) {
			assert.Equal(t, input.Name, actual.Name)
			assert.Equal(t, input.State, actual.State)
			assert.Equal(t, input.NodeId, actual.NodeId)
			assert.Equal(t, input.Role, actual.Role)
			assert.Equal(t, input.StatusId, actual.StatusId)

			assert.Equal(t, input.Tags, actual.Tags)
			assert.Equal(t, input.Contributors, actual.Contributors)
			assert.False(t, actual.Description.Valid)
		}
	}

}

func truncate(t *testing.T, db *sql.DB, orgID int, table string) {
	query := fmt.Sprintf("TRUNCATE TABLE \"%d\".%s CASCADE", orgID, table)
	_, err := db.Exec(query)
	assert.NoError(t, err)
}