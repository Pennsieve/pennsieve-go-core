package dbTable

import (
	"database/sql"
	"github.com/google/uuid"
	"github.com/pennsieve/pennsieve-go-core/pkg/core"
	"github.com/pennsieve/pennsieve-go-core/pkg/models/dataset"
	"log"
	"time"
)

type Dataset struct {
	Id                           int64          `json:"id"`
	Name                         string         `json:"name"`
	State                        string         `json:"state"`
	Description                  sql.NullString `json:"description"`
	UpdatedAt                    time.Time      `json:"updated_at"`
	CreatedAt                    time.Time      `json:"created_at"`
	NodeId                       sql.NullString `json:"node_id"`
	PermissionBit                sql.NullInt32  `json:"permission_bit"`
	Type                         string         `json:"type"`
	Role                         sql.NullString `json:"role"`
	Status                       string         `json:"status"`
	AutomaticallyProcessPackages bool           `json:"automatically_process_packages"`
	Licence                      sql.NullString `json:"licence"`
	Tags                         []string       `json:"tags"`
	Contributors                 []string       `json:"contributors"`
	BannerId                     uuid.UUID      `json:"banner_id"`
	ReadmeId                     uuid.UUID      `json:"readme_id"`
	StatusId                     int32          `json:"status_id"`
	PublicationStatusId          sql.NullInt32  `json:"publication_status_id"`
	Size                         sql.NullInt64  `json:"size"`
	ETag                         time.Time      `json:"etag"`
	DataUseAgreementId           sql.NullInt32  `json:"data_use_agreement_id"`
	ChangelogId                  uuid.NullUUID  `json:"changelog_id"`
}

type DatasetUser struct {
	DatasetId int64        `json:"dataset_id"`
	UserId    int64        `json:"user_id"`
	Role      dataset.Role `json:"role"`
	CreatedAt time.Time    `json:"created_at"`
	UpdatedAt time.Time    `json:"updated_at"`
}

type DatasetTeam struct {
	DatasetId int64        `json:"dataset_id"`
	TeamId    int64        `json:"team_id"`
	Role      dataset.Role `json:"role"`
	CreatedAt time.Time    `json:"created_at"`
	UpdatedAt time.Time    `json:"updated_at"`
}

// GetAll returns all rows in the Upload Record Table
func (d *Dataset) GetAll(db core.PostgresAPI, organizationId int) ([]Dataset, error) {
	queryStr := "SELECT (name, state) FROM datasets"

	rows, err := db.Query(queryStr)
	var allDatasets []Dataset
	if err == nil {
		for rows.Next() {
			var currentRecord Dataset
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
