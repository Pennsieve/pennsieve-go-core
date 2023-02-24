package pgdb

import (
	"context"
	log "github.com/sirupsen/logrus"
)

// IncrementPackageStorage increases the storage associated with the provided package.
func (q *Queries) IncrementPackageStorage(ctx context.Context, packageId int64, size int64) error {

	queryStr := "INSERT INTO package_storage AS package_storage (package_id, size) " +
		"VALUES ($1, $2) ON CONFLICT (package_id) " +
		"DO UPDATE SET size = COALESCE(package_storage.size, 0) + EXCLUDED.size;"

	_, err := q.db.ExecContext(ctx, queryStr, packageId, size)
	if err != nil {
		log.Println("Error incrementing package size: ", err)
	}

	return err
}

// IncrementPackageStorageAncestors increases the storage associated with the parents of the provided package.
func (q *Queries) IncrementPackageStorageAncestors(ctx context.Context, parentId int64, size int64) error {

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
		"INSERT INTO package_storage " +
		"AS package_storage (package_id, size) " +
		"SELECT id, $2 FROM ancestors " +
		"ON CONFLICT (package_id) " +
		"DO UPDATE SET size = COALESCE(package_storage.size, 0) + EXCLUDED.size;"

	_, err := q.db.ExecContext(ctx, queryStr, parentId, size)
	if err != nil {
		log.Println("Error incrementing package size: ", err)
	}

	return err
}
