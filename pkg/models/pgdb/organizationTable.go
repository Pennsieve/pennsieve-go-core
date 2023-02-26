package pgdb

import (
	"database/sql"
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
