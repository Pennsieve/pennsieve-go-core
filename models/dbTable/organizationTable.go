package dbTable

import (
	"database/sql"
	"fmt"
	"github.com/pennsieve/pennsieve-go-core/core"
	"time"
)

type Organization struct {
	Id            int64          `json:"id"`
	Name          string         `json:"name"`
	Slug          string         `json:"slug"`
	NodeId        string         `json:"node_id"`
	StorageBucket sql.NullString `json:"storage_bucket"`
	PublishBucket sql.NullString `json:"publish_bucket"`
	EmbargoBucket sql.NullString `json:"embargo_bucket"`
	CreatedAt     time.Time      `json:"created_at"`
	UpdatedAt     time.Time      `json:"updated_at"`
}

func (o *Organization) Get(db core.PostgresAPI, id int64) (*Organization, error) {

	queryStr := "SELECT id, name, slug, node_id, storage_bucket, publish_bucket, embargo_bucket, created_at, updated_at " +
		"FROM pennsieve.organizations WHERE id=$1;"

	var organization Organization
	row := db.QueryRow(queryStr, id)
	err := row.Scan(
		&organization.Id,
		&organization.Name,
		&organization.Slug,
		&organization.NodeId,
		&organization.StorageBucket,
		&organization.PublishBucket,
		&organization.EmbargoBucket,
		&organization.CreatedAt,
		&organization.UpdatedAt)

	switch err {
	case sql.ErrNoRows:
		fmt.Println("No rows were returned!")
		return nil, err
	case nil:
		return &organization, nil
	default:
		panic(err)
	}
}
