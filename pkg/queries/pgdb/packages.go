package pgdb

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/lib/pq"
	"github.com/pennsieve/pennsieve-go-core/pkg/models/packageInfo/packageState"
	"github.com/pennsieve/pennsieve-go-core/pkg/models/packageInfo/packageType"
	"github.com/pennsieve/pennsieve-go-core/pkg/models/pgdb"
	log "github.com/sirupsen/logrus"
	"regexp"
	"strings"
	"time"
)

func (q *Queries) AddPackages(ctx context.Context, records []pgdb.PackageParams) ([]pgdb.Package, error) {
	// Steps:
	// 1. Checks current packages in folder and check if they have sugested name.
	// 2. If package with name already exists, append (#) and check if that name exists recursively.
	// 3. Insert packages
	// 4. If package with provided packageId exists, don't create new one but return existing one.
	// 5. return Package objects for returned objects.

	// Handling folders...
	// 1. If folder already exist --> return existing folder
	// 2. Update parentId for all records to reflect existing record.

	currentTime := time.Now()
	var vals []interface{}
	var inserts []string

	// All records have the same datasetID
	datasetId := records[0].DatasetId

	// CHECK EXISTING FILES IN FOLDER AND UPDATE NAME IF NECESSARY

	// Group files by parentID so we can combine SQL queries for children of the parent.
	parentIdMap := map[int64][]pgdb.PackageParams{}
	for _, r := range records {
		parentIdMap[r.ParentId] = append(parentIdMap[r.ParentId], r)
	}

	// Iterate over map of parentIDs and get children that have names like the ones uploaded.
	for key, value := range parentIdMap {
		var names []string
		for _, v := range value {
			r := regexp.MustCompile(`(?P<FileName>[^\.]*)?\.?(?P<Extension>.*)`)
			pathParts := r.FindStringSubmatch(v.Name)
			fName := pathParts[r.SubexpIndex("FileName")]
			names = append(names, fmt.Sprintf("'%s%%'", fName))
		}
		arrayString := strings.Join(names, ",")

		var sqlString string
		switch key {
		case -1:
			// Check for files in root folder.
			sqlString = fmt.Sprintf("SELECT name "+
				"FROM packages "+
				"WHERE dataset_id=%d "+
				"AND parent_id IS NULL "+
				"AND name LIKE ANY (ARRAY[%s]) "+
				"AND state != '%s';", datasetId, arrayString, packageState.Deleting.String())
		default:
			sqlString = fmt.Sprintf("SELECT name "+
				"FROM packages "+
				"WHERE dataset_id=%d "+
				"AND parent_id=%d "+
				"AND name LIKE ANY (ARRAY[%s]) "+
				"AND state != '%s';", datasetId, key, arrayString, packageState.Deleting.String())
		}

		stmt, err := q.db.PrepareContext(ctx, sqlString)
		if err != nil {
			return nil, err
		}

		// format all vals at once
		var allNames []string
		rows, _ := stmt.Query(vals...)
		for rows.Next() {
			var currentFile string
			err = rows.Scan(
				&currentFile,
			)
			allNames = append(allNames, currentFile)
		}

		// Update names if suggested name exists for files
		// Don't do anything for folders as conflict will return the existing folder.
		for i, _ := range records {
			if records[i].PackageType != packageType.Collection {

				// TODO: Check Package Merge
				// If we need to merge --> set flag in record so we don't add package and only add file to the existing package.

				// Check Name Collision
				checkUpdateName(&records[i], 1, "", allNames)
			}
		}

	}

	sqlInsert := "INSERT INTO packages(name, type, state, node_id, parent_id, " +
		"dataset_id, owner_id, size, import_id, attributes, created_at, updated_at) VALUES "

	for index, row := range records {
		inserts = append(inserts, fmt.Sprintf("($%d,$%d,$%d,$%d,$%d,$%d,$%d,$%d,$%d,$%d,$%d,$%d)",
			index*12+1,
			index*12+2,
			index*12+3,
			index*12+4,
			index*12+5,
			index*12+6,
			index*12+7,
			index*12+8,
			index*12+9,
			index*12+10,
			index*12+11,
			index*12+12,
		))

		sqlParentId := sql.NullInt64{Valid: false}
		if row.ParentId >= 0 {
			sqlParentId = sql.NullInt64{
				Int64: row.ParentId,
				Valid: true,
			}
		}

		vals = append(vals, row.Name, row.PackageType.String(), row.PackageState.String(), row.NodeId, sqlParentId, row.DatasetId,
			row.OwnerId, row.Size, row.ImportId, row.Attributes, currentTime, currentTime)
	}

	returnRows := "id, name, type, state, node_id, parent_id, " +
		"dataset_id, owner_id, size, import_id, created_at, updated_at"

	sqlInsert = sqlInsert + strings.Join(inserts, ",") +
		fmt.Sprintf("ON CONFLICT(node_id) DO UPDATE SET updated_at=EXCLUDED.updated_at") +
		fmt.Sprintf(" RETURNING %s;", returnRows)

	//prepare the statement
	stmt, err := q.db.PrepareContext(ctx, sqlInsert)
	if err != nil {
		log.Fatalln("ERROR: ", err)
	}
	defer stmt.Close()

	// format all vals at once
	var allInsertedPackages []pgdb.Package
	rows, err := stmt.Query(vals...)
	if err != nil {
		if pqErr, ok := err.(*pq.Error); ok {
			log.Println(pqErr)
		}
		return nil, err
	}

	for rows.Next() {
		var currentRecord pgdb.Package
		err = rows.Scan(
			&currentRecord.Id,
			&currentRecord.Name,
			&currentRecord.PackageType,
			&currentRecord.PackageState,
			&currentRecord.NodeId,
			&currentRecord.ParentId,
			&currentRecord.DatasetId,
			&currentRecord.OwnerId,
			&currentRecord.Size,
			&currentRecord.ImportId,
			&currentRecord.CreatedAt,
			&currentRecord.UpdatedAt,
		)

		if err != nil {
			log.Println("ERROR: ", err)
		}

		allInsertedPackages = append(allInsertedPackages, currentRecord)

	}

	if err != nil {
		log.Println(err)
	}

	return allInsertedPackages, err
}

func (q *Queries) GetPackageChildren(ctx context.Context, parent *pgdb.Package, datasetId int, onlyFolders bool) ([]pgdb.Package, error) {
	folderFilter := ""
	if onlyFolders {
		folderFilter = fmt.Sprintf("AND type = '%s'", packageType.Collection.String())
	}

	// Return children for specific dataset in specific org with specific parent.
	// Do NOT return any packages that are in DELETE State
	queryRows := "id, name, type, state, node_id, parent_id, " +
		"dataset_id, owner_id, size, created_at, updated_at"

	queryStr := fmt.Sprintf("SELECT %s FROM packages WHERE dataset_id = %d AND parent_id = %d AND state != '%s' %s;",
		queryRows, datasetId, parent.Id, packageState.Deleting.String(), folderFilter)

	// If parent is empty => return children of root of dataset.
	if parent.NodeId == "" {
		queryStr = fmt.Sprintf("SELECT %s FROM packages WHERE dataset_id = %d AND parent_id IS NULL AND state != '%s' %s;",
			queryRows, datasetId, packageState.Deleting.String(), folderFilter)
	}

	rows, err := q.db.QueryContext(ctx, queryStr)
	var allPackages []pgdb.Package
	if err != nil {
		return nil, err
	}

	for rows.Next() {
		var currentRecord pgdb.Package
		err = rows.Scan(
			&currentRecord.Id,
			&currentRecord.Name,
			&currentRecord.PackageType,
			&currentRecord.PackageState,
			&currentRecord.NodeId,
			&currentRecord.ParentId,
			&currentRecord.DatasetId,
			&currentRecord.OwnerId,
			&currentRecord.Size,
			&currentRecord.CreatedAt,
			&currentRecord.UpdatedAt,
		)

		if err != nil {
			log.Println("ERROR: ", err)
			return nil, err
		}

		allPackages = append(allPackages, currentRecord)
	}
	return allPackages, err
}

// checkUpdateName Recursively checks name and append integer if name exists.
func checkUpdateName(item *pgdb.PackageParams, index int, newName string, namesInFolder []string) {

	if newName == "" {
		newName = item.Name
	}

	for _, n := range namesInFolder {
		if newName == n {
			r := regexp.MustCompile(`(?P<FileName>[^\.]*)?\.?(?P<Extension>.*)`)
			pathParts := r.FindStringSubmatch(item.Name)

			name := pathParts[r.SubexpIndex("FileName")]
			extension := pathParts[r.SubexpIndex("Extension")]

			index++

			updatedName := ""
			if extension != "" {
				updatedName = fmt.Sprintf("%s (%d).%s", name, index, extension)
			} else {
				updatedName = fmt.Sprintf("%s (%d)", name, index)
			}

			// Recursively call this function to check if updated name also exists.
			checkUpdateName(item, index, updatedName, namesInFolder)
			return
		}
	}

	// Update name to new name
	item.Name = newName
}
