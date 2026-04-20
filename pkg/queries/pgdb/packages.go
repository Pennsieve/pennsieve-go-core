package pgdb

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"github.com/lib/pq"
	"github.com/pennsieve/pennsieve-go-core/pkg/models/packageInfo/conflictStrategy"
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
	constraintName := "(name,dataset_id) WHERE parent_id IS NULL"
	if r.ParentId >= 0 {
		sqlParentId = sql.NullInt64{
			Int64: r.ParentId,
			Valid: true,
		}
		constraintName = "(name,dataset_id,parent_id) WHERE parent_id IS NOT NULL"
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
		return nil, fmt.Errorf("error preparing addFolder statement: %w", err)
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
		log.Error("Error creating or getting a folder")
		return nil, err
	case nil:
		return &currentRecord, nil
	default:
		return nil, err
	}

}

// AddPackages adds packages to a dataset with the legacy "keep both" conflict
// behavior: name collisions result in the new package being renamed with a
// " (N)" suffix until it lands. Equivalent to AddPackagesWithConflict using
// conflictStrategy.KeepBoth.
//   - This call should typically be wrapped in a Transaction as it will run multiple queries.
//   - Packages can be in different folders, but it is assumed that the folders already exist.
func (q *Queries) AddPackages(ctx context.Context, records []pgdb.PackageParams) ([]pgdb.Package, error) {
	return q.AddPackagesWithConflict(ctx, records, conflictStrategy.KeepBoth)
}

// AddPackagesWithConflict adds packages, using the given strategy to resolve
// name collisions with existing non-deleted packages under the same
// (dataset_id, parent_id, name) tuple.
//   - This call should typically be wrapped in a Transaction as it will run multiple queries.
//   - Collection-type records are rejected; use AddFolder instead.
//   - Replace soft-deletes the conflicting predecessor (state=DELETING, name
//     prefixed with __DELETED__<nodeId>_) and links the new package via
//     replaces_package_id / replaced_by_package_id.
//   - Fail returns an error listing conflicting names without inserting anything.
func (q *Queries) AddPackagesWithConflict(ctx context.Context, records []pgdb.PackageParams, strategy conflictStrategy.Strategy) ([]pgdb.Package, error) {
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

	var allInsertedPackages []pgdb.Package
	for parentId, pRecords := range parentIdMap {
		var inserted []pgdb.Package
		var err error
		switch strategy {
		case conflictStrategy.KeepBoth:
			inserted, err = q.addPackagesKeepBoth(ctx, parentId, pRecords)
		case conflictStrategy.Replace:
			inserted, err = q.addPackagesReplace(ctx, parentId, pRecords)
		case conflictStrategy.Fail:
			inserted, err = q.addPackagesFail(ctx, parentId, pRecords)
		default:
			return nil, fmt.Errorf("unknown conflict strategy: %s", strategy)
		}
		if err != nil {
			return nil, err
		}
		allInsertedPackages = append(allInsertedPackages, inserted...)
	}
	return allInsertedPackages, nil
}

// addPackagesKeepBoth retries with " (N)"-suffixed names until all records
// are inserted.
func (q *Queries) addPackagesKeepBoth(ctx context.Context, parentId int64, records []pgdb.PackageParams) ([]pgdb.Package, error) {
	var allInsertedPackages []pgdb.Package

	insertedPackages, failedPackages, err := q.addPackageByParent(ctx, parentId, records, nil)
	if err != nil {
		return nil, err
	}
	allInsertedPackages = append(allInsertedPackages, insertedPackages...)

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

		insertedPackages, failedPackages, err = q.addPackageByParent(ctx, parentId, failedPackages, nil)
		if err != nil {
			return nil, err
		}

		allInsertedPackages = append(allInsertedPackages, insertedPackages...)

		index++
	}
	return allInsertedPackages, nil
}

// addPackagesReplace soft-deletes each conflicting predecessor, decrements
// its storage counts, inserts the new packages with replaces_package_id set,
// then writes the back-reference. Async S3 asset cleanup is the caller's
// responsibility (publish a DeletePackageJob to the jobs queue for each
// returned package with ReplacesPackageId set).
// Caller must ensure the call runs in a transaction for atomicity.
func (q *Queries) addPackagesReplace(ctx context.Context, parentId int64, records []pgdb.PackageParams) ([]pgdb.Package, error) {
	conflicts, err := q.findConflictingPackages(ctx, parentId, records)
	if err != nil {
		return nil, err
	}

	replacementByNodeId := map[string]int64{}
	for _, r := range records {
		if old, ok := conflicts[r.Name]; ok {
			replacementByNodeId[r.NodeId] = old.Id
		}
	}

	datasetId := int64(records[0].DatasetId)

	// Rename + soft-delete each predecessor so the new insert doesn't trip
	// the unique (name, dataset_id, parent_id) partial indexes, and decrement
	// its storage so dataset/ancestor counts reflect the removal. Mirrors
	// pennsieve-api's PackageManager.delete behavior on the DB side.
	for _, old := range conflicts {
		newName := fmt.Sprintf("__DELETED__%s_%s", old.NodeId, old.Name)
		_, err := q.db.ExecContext(ctx,
			"UPDATE packages SET state=$1, name=$2 WHERE id=$3",
			packageState.Deleting.String(), newName, old.Id)
		if err != nil {
			return nil, fmt.Errorf("soft-deleting predecessor %d: %w", old.Id, err)
		}

		size, err := q.GetPackageStorageById(ctx, old.Id)
		if err != nil {
			// No storage row just means no decrement needed.
			if errors.Is(err, sql.ErrNoRows) {
				continue
			}
			return nil, fmt.Errorf("reading storage for predecessor %d: %w", old.Id, err)
		}
		if size <= 0 {
			continue
		}
		if err := q.IncrementPackageStorageAncestors(ctx, old.Id, -size); err != nil {
			return nil, fmt.Errorf("decrementing package/ancestor storage for predecessor %d: %w", old.Id, err)
		}
		if err := q.IncrementDatasetStorage(ctx, datasetId, -size); err != nil {
			return nil, fmt.Errorf("decrementing dataset storage for predecessor %d: %w", old.Id, err)
		}
	}

	inserted, failed, err := q.addPackageByParent(ctx, parentId, records, replacementByNodeId)
	if err != nil {
		return nil, err
	}
	if len(failed) > 0 {
		return nil, fmt.Errorf("replace: %d unexpected insert failures after predecessor rename", len(failed))
	}

	// Link back: old.replaced_by_package_id = new.id
	for _, p := range inserted {
		if !p.ReplacesPackageId.Valid {
			continue
		}
		_, err := q.db.ExecContext(ctx,
			"UPDATE packages SET replaced_by_package_id=$1 WHERE id=$2",
			p.Id, p.ReplacesPackageId.Int64)
		if err != nil {
			return nil, fmt.Errorf("setting back-ref on predecessor %d: %w", p.ReplacesPackageId.Int64, err)
		}
	}

	return inserted, nil
}

// addPackagesFail returns an error if any record name collides with an
// existing non-deleted package under the same parent; otherwise inserts
// straight through without the rename-retry loop.
func (q *Queries) addPackagesFail(ctx context.Context, parentId int64, records []pgdb.PackageParams) ([]pgdb.Package, error) {
	conflicts, err := q.findConflictingPackages(ctx, parentId, records)
	if err != nil {
		return nil, err
	}
	if len(conflicts) > 0 {
		names := make([]string, 0, len(conflicts))
		for n := range conflicts {
			names = append(names, n)
		}
		return nil, fmt.Errorf("conflict strategy FAIL: %d conflicting name(s): %v", len(names), names)
	}

	inserted, failed, err := q.addPackageByParent(ctx, parentId, records, nil)
	if err != nil {
		return nil, err
	}
	if len(failed) > 0 {
		// Post-check failed (race with another insert). Surface the names.
		failedNames := make([]string, 0, len(failed))
		for _, f := range failed {
			failedNames = append(failedNames, f.Name)
		}
		return nil, fmt.Errorf("conflict strategy FAIL: %d insert failure(s) after conflict check: %v", len(failed), failedNames)
	}
	return inserted, nil
}

// findConflictingPackages returns existing, non-deleted packages under the
// given parent whose name matches any of the incoming records.
func (q *Queries) findConflictingPackages(ctx context.Context, parentId int64, records []pgdb.PackageParams) (map[string]*pgdb.Package, error) {
	if len(records) == 0 {
		return nil, nil
	}

	names := make([]string, 0, len(records))
	for _, r := range records {
		names = append(names, r.Name)
	}
	datasetId := records[0].DatasetId

	var queryStr string
	var args []interface{}
	if parentId < 0 {
		queryStr = "SELECT id, name, node_id FROM packages " +
			"WHERE dataset_id=$1 AND parent_id IS NULL " +
			"AND state NOT IN ($2, $3) AND name = ANY($4)"
		args = []interface{}{datasetId, packageState.Deleting.String(), packageState.Deleted.String(), pq.Array(names)}
	} else {
		queryStr = "SELECT id, name, node_id FROM packages " +
			"WHERE dataset_id=$1 AND parent_id=$2 " +
			"AND state NOT IN ($3, $4) AND name = ANY($5)"
		args = []interface{}{datasetId, parentId, packageState.Deleting.String(), packageState.Deleted.String(), pq.Array(names)}
	}

	rows, err := q.db.QueryContext(ctx, queryStr, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	conflicts := map[string]*pgdb.Package{}
	for rows.Next() {
		var p pgdb.Package
		if err := rows.Scan(&p.Id, &p.Name, &p.NodeId); err != nil {
			return nil, err
		}
		pkg := p
		conflicts[pkg.Name] = &pkg
	}
	return conflicts, nil
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

func (q *Queries) GetPackageByNodeId(ctx context.Context, nodeId string) (*pgdb.Package, error) {

	queryRows := "id, name, type, state, node_id, parent_id, " +
		"dataset_id, owner_id, size, created_at, updated_at"

	queryStr := fmt.Sprintf("SELECT %s FROM packages WHERE node_id = %s", queryRows, nodeId)
	result := q.db.QueryRowContext(ctx, queryStr)

	currentRecord := pgdb.Package{}
	err := result.Scan(&currentRecord.Id,
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

	return &currentRecord, nil

}

// GetPackageAncestorIds returns an array of Package Ids corresponding with the ancestor Package Ids for the provided package.
//   - resulting array includes requested package Id as first entry
//   - resulting array includes first folder in dataset as last entry if package is in nested folder
func (q *Queries) GetPackageAncestorIds(ctx context.Context, packageId int64) ([]int64, error) {

	queryStr := "" +
		"WITH RECURSIVE ancestors(id, parent_id) AS (" +
		"SELECT " +
		"packages.id, " +
		"packages.parent_id " +
		"FROM packages packages " +
		"WHERE packages.id = $1 " +
		"UNION " +
		"SELECT parents.id, parents.parent_id " +
		"FROM packages parents " +
		"JOIN ancestors ON ancestors.parent_id = parents.id" +
		") " +
		"SELECT id FROM ancestors "

	rows, err := q.db.QueryContext(ctx, queryStr, packageId)
	if err != nil {
		log.Error("Error fetching package ancestors: ", err)
	}

	var ancestorIds []int64
	if err == nil {
		for rows.Next() {
			var currentRow *int64
			currentRow = new(int64)
			err = rows.Scan(&currentRow)

			if err != nil {
				log.Error("Error scanning package ids from results: ", err)
				return nil, err
			}
			ancestorIds = append(ancestorIds, *currentRow)

		}
		return ancestorIds, nil
	}
	return ancestorIds, err
}

// PRIVATE
// addPackageByParent runs the query to insert a set of packages that belong
// to the same parent folder.
//   - If adding packages to root-folder, the parentId should be set to -1.
//   - replacementByNodeId, if non-nil, maps each incoming record's NodeId to
//     the id of the existing package it replaces. The caller is responsible
//     for soft-deleting the predecessor before this call to avoid tripping
//     the (name, dataset_id, parent_id) unique indexes.
//
// It returns two arrays:
//  1. successfully created packages,
//  2. packages that failed to be inserted due to a name constraint.
func (q *Queries) addPackageByParent(ctx context.Context, parentId int64, records []pgdb.PackageParams, replacementByNodeId map[string]int64) ([]pgdb.Package, []pgdb.PackageParams, error) {

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
	constraintStr := "(name,dataset_id) WHERE parent_id IS NULL"
	if parentId >= 0 {
		sqlParentId = sql.NullInt64{
			Int64: parentId,
			Valid: true,
		}
		constraintStr = "(name,dataset_id,parent_id) WHERE parent_id IS NOT NULL"
	}

	sqlInsert := "INSERT INTO packages(name, type, state, node_id, parent_id, " +
		"dataset_id, owner_id, size, import_id, attributes, created_at, updated_at, replaces_package_id) VALUES "

	const columnsPerRow = 13
	for index, row := range validRecords {
		placeholders := make([]string, columnsPerRow)
		for c := 0; c < columnsPerRow; c++ {
			placeholders[c] = fmt.Sprintf("$%d", index*columnsPerRow+c+1)
		}
		inserts = append(inserts, "("+strings.Join(placeholders, ",")+")")

		var replacesId sql.NullInt64
		if replacementByNodeId != nil {
			if oldId, ok := replacementByNodeId[row.NodeId]; ok {
				replacesId = sql.NullInt64{Int64: oldId, Valid: true}
			}
		}

		values = append(values, row.Name, row.PackageType.String(), row.PackageState.String(), row.NodeId, sqlParentId, row.DatasetId,
			row.OwnerId, row.Size, row.ImportId, row.Attributes, currentTime, currentTime, replacesId)
	}

	returnRows := "id, name, type, state, node_id, parent_id, " +
		"dataset_id, owner_id, size, import_id, created_at, updated_at, replaces_package_id, replaced_by_package_id"

	// Returning packages. If we run into a constraint, return the existing package.
	// We will check in the addPackages function for the ID to see if there was a conflict.
	sqlInsert = sqlInsert + strings.Join(inserts, ",") +
		fmt.Sprintf("ON CONFLICT%s DO UPDATE SET updated_at=EXCLUDED.updated_at", constraintStr) +
		fmt.Sprintf(" RETURNING %s;", returnRows)

	// prepare the statement
	stmt, err := q.db.PrepareContext(ctx, sqlInsert)
	if err != nil {
		return nil, nil, fmt.Errorf("error preparing addPackageByParent statement: %w", err)
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
			&currentRecord.ReplacesPackageId,
			&currentRecord.ReplacedByPackageId,
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