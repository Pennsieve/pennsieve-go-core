package test

import (
	"database/sql"
	"fmt"
	"github.com/google/uuid"
	"github.com/pennsieve/pennsieve-go-core/pkg/models/packageInfo"
	"github.com/pennsieve/pennsieve-go-core/pkg/models/packageInfo/packageState"
	"github.com/pennsieve/pennsieve-go-core/pkg/models/packageInfo/packageType"
	"github.com/pennsieve/pennsieve-go-core/pkg/models/pgdb"
	"github.com/stretchr/testify/assert"
	"testing"
)

type TestPackageParams struct {
	Name     string
	ParentId int64
	NodeId   string
}

func GenerateTestPackages(params []TestPackageParams, datasetId int) []pgdb.PackageParams {

	var result []pgdb.PackageParams

	attr := []packageInfo.PackageAttribute{
		{
			Key:      "subtype",
			Fixed:    false,
			Value:    "Image",
			Hidden:   true,
			Category: "Pennsieve",
			DataType: "string",
		}, {
			Key:      "icon",
			Fixed:    false,
			Value:    "Microscope",
			Hidden:   true,
			Category: "Pennsieve",
			DataType: "string",
		},
	}

	for _, p := range params {
		var uploadId string
		var nodeId string
		if p.NodeId == "" {
			u, _ := uuid.NewUUID()
			uploadId = u.String()
			nodeId = fmt.Sprintf("N:Package:%s", u.String())
		} else {
			u, _ := uuid.NewUUID()
			uploadId = u.String()
			nodeId = p.NodeId
		}

		insertPackage := pgdb.PackageParams{
			Name:         p.Name,
			PackageType:  packageType.Image,
			PackageState: packageState.Unavailable,
			NodeId:       nodeId,
			ParentId:     p.ParentId,
			DatasetId:    datasetId,
			OwnerId:      1,
			Size:         1000,
			ImportId: sql.NullString{
				String: uploadId,
				Valid:  true,
			},
			Attributes: attr,
		}

		result = append(result, insertPackage)
	}

	return result
}

func Truncate(t *testing.T, db *sql.DB, orgID int, table string) {

	var query string

	switch table {
	case "organization_storage":
		query = fmt.Sprintf("TRUNCATE TABLE pennsieve.%s CASCADE", table)
	default:
		query = fmt.Sprintf("TRUNCATE TABLE \"%d\".%s CASCADE", orgID, table)
	}

	_, err := db.Exec(query)
	assert.NoError(t, err)
}
