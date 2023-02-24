package models

import (
	"database/sql"
	"github.com/pennsieve/pennsieve-go-core/pkg/models/packageInfo"
	"github.com/pennsieve/pennsieve-go-core/pkg/models/packageInfo/packageState"
	"github.com/pennsieve/pennsieve-go-core/pkg/models/packageInfo/packageType"
	"time"
)

// Package is a representation of a container on Pennsieve that contains one or more sourceFiles
type Package struct {
	Id           int64                         `json:"id"`
	Name         string                        `json:"name"`
	PackageType  packageType.Type              `json:"type"`
	PackageState packageState.State            `json:"state"`
	NodeId       string                        `json:"node_id"`
	ParentId     sql.NullInt64                 `json:"parent_id"`
	DatasetId    int                           `json:"dataset_id"`
	OwnerId      int                           `json:"owner_id"`
	Size         sql.NullInt64                 `json:"size"`
	ImportId     sql.NullString                `json:"import_id"`
	Attributes   packageInfo.PackageAttributes `json:"attributes"`
	CreatedAt    time.Time                     `json:"created_at"`
	UpdatedAt    time.Time                     `json:"updated_at"`
}

// PackageParams is used as the input to create a package
// ParentID is not an optional and -1 refers to the root folder.
type PackageParams struct {
	Name         string                        `json:"name"`
	PackageType  packageType.Type              `json:"type"`
	PackageState packageState.State            `json:"state"`
	NodeId       string                        `json:"node_id"`
	ParentId     int64                         `json:"parent_id"`
	DatasetId    int                           `json:"dataset_id"`
	OwnerId      int                           `json:"owner_id"`
	Size         int64                         `json:"size"`
	ImportId     sql.NullString                `json:"import_id"`
	Attributes   packageInfo.PackageAttributes `json:"attributes"`
}

// PackageMap maps path to models.Package
type PackageMap = map[string]Package
