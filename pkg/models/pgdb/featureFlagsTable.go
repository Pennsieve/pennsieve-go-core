package pgdb

import (
	"time"
)

type FeatureFlags struct {
	OrganizationId int64     `json:"organization_id"`
	Feature        string    `json:"feature"`
	Enabled        bool      `json:"enabled"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}
