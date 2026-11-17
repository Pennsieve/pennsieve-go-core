package pgdb

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/pennsieve/pennsieve-go-core/pkg/models/packageInfo/conflictStrategy"
	"github.com/pennsieve/pennsieve-go-core/pkg/models/packageInfo/packageState"
	"github.com/pennsieve/pennsieve-go-core/pkg/models/pgdb"
	"github.com/pennsieve/pennsieve-go-core/test"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// deleteJobsDDL mirrors the delete_jobs migration (pennsieve-db-migrations,
// migrations/organization). The seed database these tests run against may
// predate that migration, so the tests create the table if it is missing.
//
// TODO: delete this and ensureDeleteJobsTable once docker-compose.test.yml
// points at a pennsievedb seed image that includes the migration.
const deleteJobsDDL = `
CREATE TABLE IF NOT EXISTS "%[1]d".delete_jobs (
  id                 UUID PRIMARY KEY,
  package_id         INTEGER NOT NULL,
  package_node_id    VARCHAR(255) NOT NULL,
  dataset_id         INTEGER NOT NULL REFERENCES "%[1]d".datasets(id) ON DELETE CASCADE,
  user_node_id       VARCHAR(255),
  trace_id           VARCHAR(255),
  source             VARCHAR(50) NOT NULL DEFAULT 'UNKNOWN',
  state              VARCHAR(20) NOT NULL DEFAULT 'PENDING'
                     CHECK (state IN ('PENDING','IN_PROGRESS','SUCCEEDED','PARTIAL','FAILED')),
  attempts           INTEGER NOT NULL DEFAULT 0,
  last_error         TEXT,
  packages_total     INTEGER,
  packages_deleted   INTEGER NOT NULL DEFAULT 0,
  bytes_reclaimed    BIGINT  NOT NULL DEFAULT 0,
  s3_objects_deleted INTEGER NOT NULL DEFAULT 0,
  s3_objects_failed  INTEGER NOT NULL DEFAULT 0,
  created_at         TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  started_at         TIMESTAMP,
  finished_at        TIMESTAMP,
  updated_at         TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
)`

func ensureDeleteJobsTable(t *testing.T, db *sql.DB, orgId int) {
	t.Helper()
	_, err := db.Exec(fmt.Sprintf(deleteJobsDDL, orgId))
	require.NoError(t, err)
}

// TestDeleteJobs is the test suite for CreateDeleteJobs.
func TestDeleteJobs(t *testing.T) {
	for scenario, fn := range map[string]func(
		tt *testing.T, store *SQLStore, orgId int,
	){
		"Create PENDING rows":                  testCreateDeleteJobs,
		"Create nothing for empty input":       testCreateDeleteJobsEmpty,
		"Validation":                           testCreateDeleteJobsValidation,
		"Unknown package fails the batch":      testCreateDeleteJobsUnknownPackage,
		"Atomic with the DELETING flip":        testCreateDeleteJobsAtomicWithFlip,
		"Replace flow records predecessor job": testCreateDeleteJobsForReplace,
	} {
		t.Run(scenario, func(t *testing.T) {
			orgId := 1
			ensureDeleteJobsTable(t, testDB[orgId], orgId)
			store := NewSQLStore(testDB[orgId])
			fn(t, store, orgId)
		})
	}
}

func cleanDeleteJobsFixtures(t *testing.T, store *SQLStore, orgId int) {
	test.Truncate(t, store.db, orgId, "delete_jobs")
	test.Truncate(t, store.db, orgId, "packages")
}

func seedPackages(t *testing.T, store *SQLStore, names ...string) []pgdb.Package {
	t.Helper()
	params := make([]test.PackageParams, len(names))
	for i, n := range names {
		params[i] = test.PackageParams{Name: n, ParentId: -1}
	}
	inserted, err := store.AddPackagesWithConflict(context.Background(),
		test.GenerateTestPackages(params, 1), conflictStrategy.KeepBoth)
	require.NoError(t, err)
	require.Len(t, inserted, len(names))
	return inserted
}

func countDeleteJobs(t *testing.T, store *SQLStore, orgId int) int {
	t.Helper()
	var n int
	require.NoError(t, store.db.QueryRow(
		fmt.Sprintf(`SELECT count(*) FROM "%d".delete_jobs`, orgId)).Scan(&n))
	return n
}

func testCreateDeleteJobs(t *testing.T, store *SQLStore, orgId int) {
	defer cleanDeleteJobsFixtures(t, store, orgId)
	ctx := context.Background()
	pkgs := seedPackages(t, store, "a.csv", "b.csv")

	jobs, err := store.Queries.CreateDeleteJobs(ctx, []DeleteJobParams{
		{PackageID: pkgs[0].Id, UserNodeID: "N:user:1", TraceID: "trace-1", Source: DeleteJobSourceAPIDelete},
		{PackageID: pkgs[1].Id, UserNodeID: "N:user:1", TraceID: "trace-1", Source: DeleteJobSourceUploadReplace},
	})
	require.NoError(t, err)
	require.Len(t, jobs, 2)

	// IDs are distinct UUIDs, returned in input order.
	for i, j := range jobs {
		_, perr := uuid.Parse(j.ID)
		assert.NoError(t, perr)
		assert.Equal(t, pkgs[i].Id, j.PackageID)
	}
	assert.NotEqual(t, jobs[0].ID, jobs[1].ID)

	// Rows carry what the consumer and the monitor need.
	type row struct {
		packageId             int64
		nodeId, state, source string
		userNodeId, traceId   string
		datasetId, attempts   int
	}
	for i, j := range jobs {
		var r row
		require.NoError(t, store.db.QueryRow(fmt.Sprintf(
			`SELECT package_id, package_node_id, state, source, user_node_id, trace_id, dataset_id, attempts
			   FROM "%d".delete_jobs WHERE id = $1`, orgId), j.ID).
			Scan(&r.packageId, &r.nodeId, &r.state, &r.source, &r.userNodeId, &r.traceId, &r.datasetId, &r.attempts))
		assert.Equal(t, pkgs[i].Id, r.packageId)
		assert.Equal(t, pkgs[i].NodeId, r.nodeId, "package_node_id comes from packages")
		assert.Equal(t, int(pkgs[i].DatasetId), r.datasetId, "dataset_id comes from packages")
		assert.Equal(t, "PENDING", r.state)
		assert.Equal(t, 0, r.attempts)
		assert.Equal(t, "N:user:1", r.userNodeId)
		assert.Equal(t, "trace-1", r.traceId)
	}
	var s0, s1 string
	require.NoError(t, store.db.QueryRow(fmt.Sprintf(`SELECT source FROM "%d".delete_jobs WHERE id=$1`, orgId), jobs[0].ID).Scan(&s0))
	require.NoError(t, store.db.QueryRow(fmt.Sprintf(`SELECT source FROM "%d".delete_jobs WHERE id=$1`, orgId), jobs[1].ID).Scan(&s1))
	assert.Equal(t, DeleteJobSourceAPIDelete, s0)
	assert.Equal(t, DeleteJobSourceUploadReplace, s1)
}

func testCreateDeleteJobsEmpty(t *testing.T, store *SQLStore, orgId int) {
	defer cleanDeleteJobsFixtures(t, store, orgId)
	jobs, err := store.Queries.CreateDeleteJobs(context.Background(), nil)
	assert.NoError(t, err)
	assert.Empty(t, jobs)
	assert.Equal(t, 0, countDeleteJobs(t, store, orgId))
}

func testCreateDeleteJobsValidation(t *testing.T, store *SQLStore, orgId int) {
	defer cleanDeleteJobsFixtures(t, store, orgId)
	ctx := context.Background()
	pkgs := seedPackages(t, store, "v.csv")
	ok := DeleteJobParams{PackageID: pkgs[0].Id, UserNodeID: "N:user:1", TraceID: "t", Source: DeleteJobSourceAPIDelete}

	bad := map[string]DeleteJobParams{
		"zero package":  {PackageID: 0, UserNodeID: ok.UserNodeID, TraceID: ok.TraceID, Source: ok.Source},
		"no user":       {PackageID: ok.PackageID, TraceID: ok.TraceID, Source: ok.Source},
		"no trace":      {PackageID: ok.PackageID, UserNodeID: ok.UserNodeID, Source: ok.Source},
		"no source":     {PackageID: ok.PackageID, UserNodeID: ok.UserNodeID, TraceID: ok.TraceID},
		"source length": {PackageID: ok.PackageID, UserNodeID: ok.UserNodeID, TraceID: ok.TraceID, Source: strings.Repeat("x", 51)},
	}
	for name, p := range bad {
		_, err := store.Queries.CreateDeleteJobs(ctx, []DeleteJobParams{p})
		assert.Error(t, err, name)
	}
	// A bad entry anywhere rejects the whole batch before anything is written.
	_, err := store.Queries.CreateDeleteJobs(ctx, []DeleteJobParams{ok, bad["no trace"]})
	assert.Error(t, err)
	assert.Equal(t, 0, countDeleteJobs(t, store, orgId))
}

func testCreateDeleteJobsUnknownPackage(t *testing.T, store *SQLStore, orgId int) {
	defer cleanDeleteJobsFixtures(t, store, orgId)
	ctx := context.Background()
	pkgs := seedPackages(t, store, "real.csv")

	tx, err := store.db.BeginTx(ctx, nil)
	require.NoError(t, err)
	_, err = store.Queries.WithTx(tx).CreateDeleteJobs(ctx, []DeleteJobParams{
		{PackageID: pkgs[0].Id, UserNodeID: "N:user:1", TraceID: "t", Source: DeleteJobSourceAPIDelete},
		{PackageID: 999999, UserNodeID: "N:user:1", TraceID: "t", Source: DeleteJobSourceAPIDelete},
	})
	assert.Error(t, err, "a package that does not exist must fail the batch")
	require.NoError(t, tx.Rollback())
	assert.Equal(t, 0, countDeleteJobs(t, store, orgId), "caller's rollback leaves no rows")
}

// The point of the contract: the PENDING row and the DELETING flip commit
// together or not at all.
func testCreateDeleteJobsAtomicWithFlip(t *testing.T, store *SQLStore, orgId int) {
	defer cleanDeleteJobsFixtures(t, store, orgId)
	ctx := context.Background()
	pkgs := seedPackages(t, store, "atomic.csv")
	target := pkgs[0]

	stateOf := func() string {
		var s string
		require.NoError(t, store.db.QueryRow(
			fmt.Sprintf(`SELECT state FROM "%d".packages WHERE id=$1`, orgId), target.Id).Scan(&s))
		return s
	}
	flipAndRecord := func(q *Queries) error {
		if err := q.PrepareReplace(ctx, []*pgdb.Package{&target}); err != nil {
			return err
		}
		_, err := q.CreateDeleteJobs(ctx, []DeleteJobParams{
			{PackageID: target.Id, UserNodeID: "N:user:1", TraceID: "t", Source: DeleteJobSourceUploadReplace},
		})
		return err
	}

	// Rolled back: neither the flip nor the row survives.
	tx, err := store.db.BeginTx(ctx, nil)
	require.NoError(t, err)
	require.NoError(t, flipAndRecord(store.Queries.WithTx(tx)))
	require.NoError(t, tx.Rollback())
	assert.NotEqual(t, packageState.Deleting.String(), stateOf())
	assert.Equal(t, 0, countDeleteJobs(t, store, orgId))

	// Committed: both are there.
	tx, err = store.db.BeginTx(ctx, nil)
	require.NoError(t, err)
	require.NoError(t, flipAndRecord(store.Queries.WithTx(tx)))
	require.NoError(t, tx.Commit())
	assert.Equal(t, packageState.Deleting.String(), stateOf())
	assert.Equal(t, 1, countDeleteJobs(t, store, orgId))
}

// The upload replace flow: AddPackagesWithConflict(Replace) flips the
// predecessor; the caller then records a job for each replaced predecessor in
// the same transaction.
func testCreateDeleteJobsForReplace(t *testing.T, store *SQLStore, orgId int) {
	defer cleanDeleteJobsFixtures(t, store, orgId)
	ctx := context.Background()
	old := seedPackages(t, store, "report.csv")[0]

	tx, err := store.db.BeginTx(ctx, nil)
	require.NoError(t, err)
	q := store.Queries.WithTx(tx)

	replacement := test.GenerateTestPackages([]test.PackageParams{{Name: "report.csv", ParentId: -1}}, 1)
	inserted, err := q.AddPackagesWithConflict(ctx, replacement, conflictStrategy.Replace)
	require.NoError(t, err)
	require.Len(t, inserted, 1)
	require.True(t, inserted[0].ReplacesPackageId.Valid)
	require.Equal(t, old.Id, inserted[0].ReplacesPackageId.Int64)

	var params []DeleteJobParams
	for _, p := range inserted {
		if p.ReplacesPackageId.Valid {
			params = append(params, DeleteJobParams{
				PackageID: p.ReplacesPackageId.Int64, UserNodeID: "N:user:1",
				TraceID: "manifest-1", Source: DeleteJobSourceUploadReplace,
			})
		}
	}
	jobs, err := q.CreateDeleteJobs(ctx, params)
	require.NoError(t, err)
	require.Len(t, jobs, 1)
	require.NoError(t, tx.Commit())

	// The job is for the predecessor (now DELETING), not the new package.
	var pkgId int64
	var state string
	require.NoError(t, store.db.QueryRow(fmt.Sprintf(
		`SELECT d.package_id, p.state FROM "%[1]d".delete_jobs d JOIN "%[1]d".packages p ON p.id = d.package_id WHERE d.id=$1`, orgId),
		jobs[0].ID).Scan(&pkgId, &state))
	assert.Equal(t, old.Id, pkgId)
	assert.Equal(t, packageState.Deleting.String(), state)
}
