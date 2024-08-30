package pgdb

import (
	"database/sql"
	"database/sql/driver"
	"github.com/google/uuid"
	"github.com/lib/pq"
	"github.com/pennsieve/pennsieve-go-core/pkg/models/dataset/role"
	"time"
)

type Tags []string
type Properties map[string]interface{}

func (t Tags) Value() (driver.Value, error) {
	return (*pq.StringArray)(&t).Value()
}

func (t *Tags) Scan(src any) error {
	return (*pq.StringArray)(t).Scan(src)
}

type Contributors []string

func (c Contributors) Value() (driver.Value, error) {
	return (*pq.StringArray)(&c).Value()
}

func (c *Contributors) Scan(src any) error {
	return (*pq.StringArray)(c).Scan(src)
}

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
	License                      sql.NullString `json:"license"`
	Tags                         Tags           `json:"tags"`
	Contributors                 Contributors   `json:"contributors"`
	BannerId                     uuid.UUID      `json:"banner_id"`
	ReadmeId                     uuid.UUID      `json:"readme_id"`
	StatusId                     int32          `json:"status_id"`
	PublicationStatusId          sql.NullInt32  `json:"publication_status_id"`
	Size                         sql.NullInt64  `json:"size"`
	ETag                         time.Time      `json:"etag"`
	DataUseAgreementId           sql.NullInt32  `json:"data_use_agreement_id"`
	ChangelogId                  uuid.NullUUID  `json:"changelog_id"`
}

type DatasetStatus struct {
	Id           int64     `json:"id"`
	Name         string    `json:"name"`
	DisplayName  string    `json:"display_name"`
	OriginalName string    `json:"original_name"`
	Color        string    `json:"color"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

type DatasetUser struct {
	DatasetId     int64     `json:"dataset_id"`
	UserId        int64     `json:"user_id"`
	Role          string    `json:"role"`
	PermissionBit int64     `json:"permission_bit"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

type DatasetTeam struct {
	DatasetId int64     `json:"dataset_id"`
	TeamId    int64     `json:"team_id"`
	Role      role.Role `json:"role"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type DatasetContributor struct {
	DatasetId        int64     `json:"dataset_id"`
	ContributorId    int64     `json:"contributor_id"`
	CreatedAt        time.Time `json:"created_at"`
	UpdatedAt        time.Time `json:"updated_at"`
	ContributorOrder int64     `json:"contributor_order"`
}

type DatasetRelease struct {
	Id               int64          `json:"id"`
	DatasetId        int64          `json:"dataset_id"`
	Origin           string         `json:"origin"`
	Url              string         `json:"url"`
	Label            sql.NullString `json:"label"`
	Marker           sql.NullString `json:"marker"`
	Properties       Properties     `json:"properties"`
	Tags             Tags           `json:"tags"`
	ReleaseDate      sql.NullTime   `json:"release_date"`
	ReleaseStatus    string         `json:"release_status"`
	PublishingStatus string         `json:"publishing_status"`
	CreatedAt        time.Time      `json:"created_at"`
	UpdatedAt        time.Time      `json:"updated_at"`
}

type DatasetReleaseDTO struct {
	Dataset Dataset        `json:"dataset"`
	Release DatasetRelease `json:"release"`
}
