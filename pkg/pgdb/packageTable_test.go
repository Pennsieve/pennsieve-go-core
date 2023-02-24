package pgdb

import (
	"context"
	"database/sql"
	"github.com/pennsieve/pennsieve-go-core/pkg/core"
	"github.com/pennsieve/pennsieve-go-core/pkg/models/packageInfo"
	"github.com/pennsieve/pennsieve-go-core/pkg/models/packageInfo/packageState"
	"github.com/pennsieve/pennsieve-go-core/pkg/models/packageInfo/packageType"
	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestPackageTable(t *testing.T) {
	for scenario, fn := range map[string]func(
		tt *testing.T, store *SQLStore,
	){
		"Add package": testAddPackage,
	} {
		t.Run(scenario, func(t *testing.T) {
			orgId := 2
			testDB, err := core.ConnectENVWithOrg(orgId)
			if err != nil {
				log.Fatal("cannot connect to db:", err)
			}

			store := NewSQLStore(testDB)
			fn(t, store)
		})
	}
}

func testAddPackage(t *testing.T, store *SQLStore) {

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

	records := []PackageParams{
		{
			Name:         "TestAddPackage.jpg",
			PackageType:  packageType.Image,
			PackageState: packageState.Ready,
			NodeId:       "N:package:12312314",
			ParentId:     -1,
			DatasetId:    1,
			OwnerId:      1,
			Size:         1000,
			ImportId:     sql.NullString{String: "12323243243245678"},
			Attributes:   attr,
		},
	}
	results, err := store.AddPackages(context.Background(), records)
	assert.NoError(t, err)
	assert.Equal(t, records[0].Name, results[0].Name)

}

//
//func truncate(t *testing.T, db *sql.DB, orgID int, table string) {
//	query := fmt.Sprintf("TRUNCATE TABLE \"%d\".%s CASCADE", orgID, table)
//	_, err := db.Exec(query)
//	assert.NoError(t, err)
//}
//
//func TestPackageAttributeValueAndScan(t *testing.T) {
//	tests := map[string]packageInfo.PackageAttributes{
//		"non-empty": {
//			{Key: "subtype",
//				Fixed:    false,
//				Value:    "Image",
//				Hidden:   true,
//				Category: "Pennsieve",
//				DataType: "string"},
//			{Key: "icon",
//				Fixed:    false,
//				Value:    "Microscope",
//				Hidden:   true,
//				Category: "Pennsieve",
//				DataType: "string"}},
//		"nil":   nil,
//		"empty": {},
//	}
//
//	orgId := 2
//	db, err := core.ConnectENVWithOrg(orgId)
//
//	assert.NoError(t, err)
//	defer db.Close()
//	for name, expectedAttributes := range tests {
//		t.Run(name, func(t *testing.T) {
//			p := Package{
//				Name:         "image.jpg",
//				PackageType:  packageType.Image,
//				PackageState: packageState.Ready,
//				NodeId:       "N:package:1234",
//				DatasetId:    1,
//				OwnerId:      1,
//				Attributes:   expectedAttributes}
//			insert := fmt.Sprintf(
//				"INSERT INTO \"%d\".packages (name, type, state, node_id, dataset_id, owner_id, attributes) VALUES ($1, $2, $3, $4, $5, $6, $7)",
//				orgId)
//			_, err = db.Exec(insert, p.Name, p.PackageType, p.PackageState, p.NodeId, p.DatasetId, p.OwnerId, p.Attributes)
//			assert.NoError(t, err)
//			defer truncate(t, db, orgId, "packages")
//
//			countStmt := fmt.Sprintf("SELECT COUNT(*) FROM \"%d\".packages", orgId)
//			var count int
//			assert.NoError(t, db.QueryRow(countStmt).Scan(&count))
//			assert.Equal(t, 1, count)
//
//			selectStmt := fmt.Sprintf(
//				"SELECT name, type, state, node_id, dataset_id, owner_id, attributes FROM \"%d\".packages",
//				orgId)
//
//			var actual Package
//			assert.NoError(t, db.QueryRow(selectStmt).Scan(
//				&actual.Name,
//				&actual.PackageType,
//				&actual.PackageState,
//				&actual.NodeId,
//				&actual.DatasetId,
//				&actual.OwnerId,
//				&actual.Attributes))
//
//			assert.Equal(t, p.Name, actual.Name)
//			assert.Equal(t, p.PackageType, actual.PackageType)
//			assert.Equal(t, p.PackageState, actual.PackageState)
//			assert.Equal(t, p.NodeId, actual.NodeId)
//			assert.Equal(t, p.DatasetId, actual.DatasetId)
//			assert.Equal(t, p.OwnerId, actual.OwnerId)
//			assert.Equal(t, p.Attributes, actual.Attributes)
//		})
//	}
//
//}
