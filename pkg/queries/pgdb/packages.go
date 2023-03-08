package pgdb

import (
	"context"
	"database/sql"
	"errors"
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

// AddFolder adds a single folder to a dataset.
func (q *Queries) AddFolder(ctx context.Context, r pgdb.PackageParams) (*pgdb.Package, error) {

	if r.PackageType != packageType.Collection {
		return nil, errors.New("record is not of type COLLECTION")
	}

	currentTime := time.Now()
	sqlInsert := "INSERT INTO packages(name, type, state, node_id, parent_id, " +
		"dataset_id, owner_id, size, import_id, attributes, created_at, updated_at) " +
		"VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12)"

	sqlParentId := sql.NullInt64{Valid: false}
	constraintName := "(name,dataset_id,\"type\") WHERE parent_id IS NULL"
	if r.ParentId >= 0 {
		sqlParentId = sql.NullInt64{
			Int64: r.ParentId,
			Valid: true,
		}
		constraintName = "(name,dataset_id,\"type\",parent_id) WHERE parent_id IS NOT NULL"
	}

	var values []interface{}
	values = append(values, r.Name, r.PackageType.String(), r.PackageState.String(), r.NodeId, sqlParentId, r.DatasetId,
		r.OwnerId, nil, r.ImportId, r.Attributes, currentTime, currentTime)

	returnRows := "id, name, type, state, node_id, parent_id, " +
		"dataset_id, owner_id, size, import_id, created_at, updated_at"

	sqlInsert = sqlInsert +
		fmt.Sprintf("ON CONFLICT%s DO UPDATE SET updated_at=EXCLUDED.updated_at", constraintName) +
		fmt.Sprintf(" RETURNING %s;", returnRows)

	//prepare the statement
	stmt, err := q.db.PrepareContext(ctx, sqlInsert)
	if err != nil {
		log.Fatalln("ERROR: ", err)
	}
	//goland:noinspection ALL
	defer stmt.Close()

	// format all values at once
	row := stmt.QueryRowContext(ctx, values...)
	var currentRecord pgdb.Package
	err = row.Scan(
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

	switch err {
	case sql.ErrNoRows:
		fmt.Println("Error creating or getting a folder")
		return nil, err
	case nil:
		return &currentRecord, nil
	default:
		return nil, err
	}

}

// AddPackages adds packages to a dataset.
// 	* This call should typically be wrapped in a Transaction as it will run multiple queries.
// 	* Packages can be in different folders, but it is assumed that the folders already exist.
func (q *Queries) AddPackages(ctx context.Context, records []pgdb.PackageParams) ([]pgdb.Package, error) {
	var allInsertedPackages []pgdb.Package

	for _, r := range records {
		if r.PackageType == packageType.Collection {
			return nil, errors.New("cannot create COLLECTION package with AddPackages method, use AddFolder instead")
		}
	}

	// Group files by parentID, so we can combine SQL queries for children of the parent.
	parentIdMap := map[int64][]pgdb.PackageParams{}
	for _, r := range records {
		parentIdMap[r.ParentId] = append(parentIdMap[r.ParentId], r)
	}

	for parentId, pRecords := range parentIdMap {
		insertedPackages, failedPackages, err := q.addPackageByParent(ctx, parentId, pRecords)
		if err != nil {
			return nil, err
		}

		allInsertedPackages = append(allInsertedPackages, insertedPackages...)

		// Update name of failed packages and re-insert.
		failedNameMap := map[string]string{}
		for _, p := range failedPackages {
			failedNameMap[p.Name] = p.Name
		}

		index := 1
		for len(failedPackages) > 0 {
			for i := range failedPackages {

				originalName := failedNameMap[failedPackages[i].Name]
				expandName(&failedPackages[i], originalName, index)
				failedNameMap[failedPackages[i].Name] = originalName
			}

			insertedPackages, failedPackages, err = q.addPackageByParent(ctx, parentId, failedPackages)
			if err != nil {
				return nil, err
			}

			allInsertedPackages = append(allInsertedPackages, insertedPackages...)

			index++
		}
	}
	return allInsertedPackages, nil

}

// GetPackageChildren Get the children in a package
func (q *Queries) GetPackageChildren(ctx context.Context, parent *pgdb.Package, datasetId int, onlyFolders bool) ([]pgdb.Package, error) {

	folderFilter := ""
	if onlyFolders {
		folderFilter = fmt.Sprintf("AND type = '%s'", packageType.Collection.String())
	}

	// Return children for specific dataset in specific org with specific parent.
	// Do NOT return any packages that are in DELETE State
	queryRows := "id, name, type, state, node_id, parent_id, " +
		"dataset_id, owner_id, size, created_at, updated_at"

	// If parent is empty => return children of root of dataset.
	var queryStr string
	if parent == nil {
		queryStr = fmt.Sprintf("SELECT %s FROM packages WHERE dataset_id = %d AND parent_id IS NULL AND state != '%s' %s;",
			queryRows, datasetId, packageState.Deleting.String(), folderFilter)
	} else {
		queryStr = fmt.Sprintf("SELECT %s FROM packages WHERE dataset_id = %d AND parent_id = %d AND state != '%s' %s;",
			queryRows, datasetId, parent.Id, packageState.Deleting.String(), folderFilter)
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

// PRIVATE
// AddPackages runs the query to insert a set of packages that belong to the same parent folder.
// It returns two arrays:
// 		1) successfully created packages,
//		2) packages that failed to be inserted due to a name constraint.
func (q *Queries) addPackageByParent(ctx context.Context, parentId int64, records []pgdb.PackageParams) ([]pgdb.Package, []pgdb.PackageParams, error) {

	log.Debug(fmt.Sprintf("ADD PACKAGES: %v", records))

	var validRecords []pgdb.PackageParams
	var validNames []string
	var expectedNodeIds []string

	for _, r := range records {
		// Check all packages share provided parent
		if r.ParentId != parentId {
			return nil, nil, errors.New("mismatch provided parentID and parentId in packageParams")
		}

		// Check for name duplication and return duplicates as failed inserts
		// Create a list of expected Node IDS to compare with actual results from query.
		if !contains(validNames, r.Name) {
			validRecords = append(validRecords, r)
			validNames = append(validNames, r.Name)
			expectedNodeIds = append(expectedNodeIds, r.NodeId)
		}
	}

	currentTime := time.Now()
	var values []interface{}
	var inserts []string

	// Convert parentId to sql.null int64 (root folder is -1 in params, but nil in table)
	sqlParentId := sql.NullInt64{Valid: false}
	constraintStr := "(name,dataset_id,\"type\") WHERE parent_id IS NULL"
	if parentId >= 0 {
		sqlParentId = sql.NullInt64{
			Int64: parentId,
			Valid: true,
		}
		constraintStr = "(name,dataset_id,\"type\",parent_id) WHERE parent_id IS NOT NULL"
	}

	sqlInsert := "INSERT INTO packages(name, type, state, node_id, parent_id, " +
		"dataset_id, owner_id, size, import_id, attributes, created_at, updated_at) VALUES "

	for index, row := range validRecords {
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

		values = append(values, row.Name, row.PackageType.String(), row.PackageState.String(), row.NodeId, sqlParentId, row.DatasetId,
			row.OwnerId, row.Size, row.ImportId, row.Attributes, currentTime, currentTime)
	}

	returnRows := "id, name, type, state, node_id, parent_id, " +
		"dataset_id, owner_id, size, import_id, created_at, updated_at"

	// Returning packages. If we run into a constraint, return the existing package.
	// We will check in the addPackages function for the ID to see if there was a conflict.
	sqlInsert = sqlInsert + strings.Join(inserts, ",") +
		fmt.Sprintf("ON CONFLICT%s DO UPDATE SET updated_at=EXCLUDED.updated_at", constraintStr) +
		fmt.Sprintf(" RETURNING %s;", returnRows)

	// prepare the statement
	stmt, err := q.db.PrepareContext(ctx, sqlInsert)
	if err != nil {
		log.Fatalln("ERROR: ", err)
	}
	//goland:noinspection GoUnhandledErrorResult
	defer stmt.Close()

	log.Debug(fmt.Sprintf("Insert statement: %v", stmt))

	// format all values at once
	var allInsertedPackages []pgdb.Package
	rows, err := stmt.Query(values...)
	if err != nil {
		if pqErr, ok := err.(*pq.Error); ok {
			log.Println(pqErr)
		}
		return nil, nil, err
	}

	var resultNodeIds []string
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
			return nil, nil, err
		}

		// Only return newly inserted objects
		if contains(expectedNodeIds, currentRecord.NodeId) {
			allInsertedPackages = append(allInsertedPackages, currentRecord)
			resultNodeIds = append(resultNodeIds, currentRecord.NodeId)
		}

	}

	// Check the nodeIds of the package
	var failedInserts []pgdb.PackageParams
	for _, r := range records {
		if !contains(resultNodeIds, r.NodeId) {
			log.Debug(fmt.Sprintf("MISMATCH NODEID: %s, %v", r.NodeId, resultNodeIds))

			failedInserts = append(failedInserts, r)
		}
	}

	return allInsertedPackages, failedInserts, nil

}

// HELPER FUNCTIONS

// contains checks if a string is present in a slice
func contains(s []string, str string) bool {
	for _, v := range s {
		if v == str {
			return true
		}
	}

	return false
}

// expandName appends an index to a package name.
func expandName(item *pgdb.PackageParams, originalName string, index int) *pgdb.PackageParams {
	r := regexp.MustCompile(`(?P<FileName>[^.]*)\.?(?P<Extension>.*)`)
	pathParts := r.FindStringSubmatch(originalName)

	filePart := pathParts[r.SubexpIndex("FileName")]
	extension := pathParts[r.SubexpIndex("Extension")]

	updatedName := ""
	if extension != "" {
		updatedName = fmt.Sprintf("%s (%d).%s", filePart, index, extension)
		log.Debug(fmt.Sprintf("CHECKUPDATE -Add index: %d - %s", index, updatedName))

	} else {
		updatedName = fmt.Sprintf("%s (%d)", filePart, index)
		log.Debug(fmt.Sprintf("CHECKUPDATE -Add index: %d - %s", index, updatedName))

	}

	item.Name = updatedName

	return item
}
