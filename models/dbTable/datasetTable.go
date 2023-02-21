package dbTable

import (
	"github.com/pennsieve/pennsieve-go-core/core"
	"github.com/pennsieve/pennsieve-go-core/models/dataset"
	"log"
	"time"
)

type Dataset struct {
	Name  string `json:"name"`
	State string `json:"state"`
	Role  string `json:"role"`
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
