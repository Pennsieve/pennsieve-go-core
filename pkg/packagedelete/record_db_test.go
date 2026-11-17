package packagedelete

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
	"testing"

	_ "github.com/lib/pq"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// These tests run RecordPendingJobs against a real Postgres. They build
// their own scratch organization schema with a minimal packages table
// and the delete_jobs table copied from pennsieve-db-migrations
// V20261117092140__create_delete_jobs.sql (minus the updated_at trigger,
// which only fires on UPDATE), so they do not depend on the seed image
// already carrying the migration.
//
// They need POSTGRES_HOST (set by docker-compose.test.yml); without it
// they skip.

// testOrg is a schema id no seed organization uses.
const testOrg int64 = 987654

const scratchSchemaDDL = `
DROP SCHEMA IF EXISTS "987654" CASCADE;
CREATE SCHEMA "987654";
CREATE TABLE "987654".packages (
  id         SERIAL PRIMARY KEY,
  name       VARCHAR(255) NOT NULL,
  node_id    VARCHAR(255) NOT NULL,
  dataset_id INTEGER NOT NULL,
  state      VARCHAR(255) NOT NULL
);
CREATE TABLE "987654".delete_jobs (
  id                 UUID PRIMARY KEY,
  package_id         INTEGER NOT NULL,
  package_node_id    VARCHAR(255) NOT NULL,
  dataset_id         INTEGER NOT NULL,
  user_node_id       VARCHAR(255),
  trace_id           VARCHAR(255),
  source             VARCHAR(32) NOT NULL DEFAULT 'UNKNOWN'
                     CHECK (source IN ('API_DELETE','UPLOAD_REPLACE','OTHER','UNKNOWN')),
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
);`

func openTestDB(t *testing.T) *sql.DB {
	t.Helper()
	host := os.Getenv("POSTGRES_HOST")
	if host == "" {
		t.Skip("POSTGRES_HOST not set; skipping delete_jobs database tests")
	}
	dsn := fmt.Sprintf("host=%s port=%s user=%s dbname=%s sslmode=%s",
		host, envOr("POSTGRES_PORT", "5432"), envOr("POSTGRES_USER", "postgres"),
		envOr("PENNSIEVE_DB", "postgres"), envOr("POSTGRES_SSL_MODE", "disable"))
	if pw := os.Getenv("POSTGRES_PASSWORD"); pw != "" {
		dsn += " password=" + pw
	}
	db, err := sql.Open("postgres", dsn)
	require.NoError(t, err)
	require.NoError(t, db.Ping())

	_, err = db.Exec(scratchSchemaDDL)
	require.NoError(t, err)
	t.Cleanup(func() {
		_, _ = db.Exec(`DROP SCHEMA IF EXISTS "987654" CASCADE`)
		_ = db.Close()
	})
	return db
}

func envOr(k, def string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return def
}

func insertPackage(t *testing.T, db *sql.DB, name, nodeID string, datasetID int64, state string) int64 {
	t.Helper()
	var id int64
	require.NoError(t, db.QueryRow(
		`INSERT INTO "987654".packages (name, node_id, dataset_id, state) VALUES ($1,$2,$3,$4) RETURNING id`,
		name, nodeID, datasetID, state).Scan(&id))
	return id
}

type jobRow struct {
	ID, PackageNodeID, UserNodeID, TraceID, Source, State string
	PackageID, DatasetID, Attempts                        int64
}

func readJobs(t *testing.T, db *sql.DB) map[string]jobRow {
	t.Helper()
	rows, err := db.Query(`SELECT id, package_id, package_node_id, dataset_id, user_node_id,
	  trace_id, source, state, attempts FROM "987654".delete_jobs`)
	require.NoError(t, err)
	defer rows.Close()
	out := map[string]jobRow{}
	for rows.Next() {
		var r jobRow
		require.NoError(t, rows.Scan(&r.ID, &r.PackageID, &r.PackageNodeID, &r.DatasetID,
			&r.UserNodeID, &r.TraceID, &r.Source, &r.State, &r.Attempts))
		out[r.ID] = r
	}
	require.NoError(t, rows.Err())
	return out
}

// TestRecordPendingJobs_DB_WritesPendingRowsMatchingMessages is the
// contract end to end: flip to DELETING and record jobs in one tx,
// commit, enqueue. Each row is PENDING/0 with node and dataset ids read
// from packages, and each message id is its row's primary key.
func TestRecordPendingJobs_DB_WritesPendingRowsMatchingMessages(t *testing.T) {
	db := openTestDB(t)
	ctx := context.Background()

	a := insertPackage(t, db, "a.txt", "N:package:a", 11, "READY")
	b := insertPackage(t, db, "b.txt", "N:package:b", 12, "READY")

	tx, err := db.BeginTx(ctx, nil)
	require.NoError(t, err)
	_, err = tx.ExecContext(ctx, `UPDATE "987654".packages SET state='DELETING' WHERE id = ANY($1)`,
		fmt.Sprintf("{%d,%d}", a, b))
	require.NoError(t, err)

	jobs, err := RecordPendingJobs(ctx, tx, []DeleteRequest{
		{PackageID: a, OrganizationID: testOrg, UserNodeID: "N:user:u", TraceID: "t-1", Source: SourceUploadReplace},
		{PackageID: b, OrganizationID: testOrg, UserNodeID: "N:user:u", TraceID: "t-1", Source: SourceUploadReplace},
	})
	require.NoError(t, err)
	require.Len(t, jobs, 2)
	require.NoError(t, tx.Commit())

	rows := readJobs(t, db)
	require.Len(t, rows, 2)
	want := map[int64]struct {
		node string
		ds   int64
	}{a: {"N:package:a", 11}, b: {"N:package:b", 12}}
	for _, j := range jobs {
		r, ok := rows[j.ID]
		require.True(t, ok, "no row for job %s", j.ID)
		assert.Equal(t, j.PackageID, r.PackageID)
		assert.Equal(t, want[r.PackageID].node, r.PackageNodeID)
		assert.Equal(t, want[r.PackageID].ds, r.DatasetID)
		assert.Equal(t, "N:user:u", r.UserNodeID)
		assert.Equal(t, "t-1", r.TraceID)
		assert.Equal(t, "UPLOAD_REPLACE", r.Source)
		assert.Equal(t, "PENDING", r.State)
		assert.EqualValues(t, 0, r.Attempts)
	}

	sender := newFakeQueueSender()
	require.NoError(t, EnqueueJobs(ctx, sender, jobs))
	require.Len(t, sender.bodies, 2)
	for _, body := range sender.bodies {
		var msg map[string]wireMessage
		require.NoError(t, json.Unmarshal(body, &msg))
		r, ok := rows[msg["DeletePackageJob"].Id]
		require.True(t, ok, "message id %s has no delete_jobs row", msg["DeletePackageJob"].Id)
		assert.Equal(t, r.PackageID, msg["DeletePackageJob"].PackageID)
	}
}

// TestRecordPendingJobs_DB_RequiresDeleting: a package that was not
// flipped to DELETING in the transaction gets no row, and the whole
// call fails so the caller rolls back.
func TestRecordPendingJobs_DB_RequiresDeleting(t *testing.T) {
	db := openTestDB(t)
	ctx := context.Background()

	flipped := insertPackage(t, db, "a.txt", "N:package:a", 11, "DELETING")
	notFlipped := insertPackage(t, db, "b.txt", "N:package:b", 11, "READY")

	tx, err := db.BeginTx(ctx, nil)
	require.NoError(t, err)
	defer func() { _ = tx.Rollback() }()

	_, err = RecordPendingJobs(ctx, tx, []DeleteRequest{
		{PackageID: flipped, OrganizationID: testOrg, UserNodeID: "N:user:u", TraceID: "t", Source: SourceAPIDelete},
		{PackageID: notFlipped, OrganizationID: testOrg, UserNodeID: "N:user:u", TraceID: "t", Source: SourceAPIDelete},
	})
	require.Error(t, err)
	assert.Contains(t, err.Error(), fmt.Sprintf("[%d]", notFlipped))
}

func TestRecordPendingJobs_DB_MissingPackage(t *testing.T) {
	db := openTestDB(t)
	ctx := context.Background()

	tx, err := db.BeginTx(ctx, nil)
	require.NoError(t, err)
	defer func() { _ = tx.Rollback() }()

	_, err = RecordPendingJobs(ctx, tx, []DeleteRequest{
		{PackageID: 424242, OrganizationID: testOrg, UserNodeID: "N:user:u", TraceID: "t", Source: SourceOther},
	})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "not found or not DELETING")
}

// TestRecordPendingJobs_DB_RollsBackWithTx: the row belongs to the
// caller's transaction, so rolling back the DELETING flip also drops it.
func TestRecordPendingJobs_DB_RollsBackWithTx(t *testing.T) {
	db := openTestDB(t)
	ctx := context.Background()

	id := insertPackage(t, db, "a.txt", "N:package:a", 11, "READY")

	tx, err := db.BeginTx(ctx, nil)
	require.NoError(t, err)
	_, err = tx.ExecContext(ctx, `UPDATE "987654".packages SET state='DELETING' WHERE id=$1`, id)
	require.NoError(t, err)
	_, err = RecordPendingJobs(ctx, tx, []DeleteRequest{
		{PackageID: id, OrganizationID: testOrg, UserNodeID: "N:user:u", TraceID: "t", Source: SourceOther},
	})
	require.NoError(t, err)
	require.NoError(t, tx.Rollback())

	assert.Empty(t, readJobs(t, db))
}

// TestRecordPendingJobs_DB_UnknownOrgSchema: the schema is named from
// OrganizationID, so a wrong org fails instead of writing elsewhere.
func TestRecordPendingJobs_DB_UnknownOrgSchema(t *testing.T) {
	db := openTestDB(t)
	ctx := context.Background()

	tx, err := db.BeginTx(ctx, nil)
	require.NoError(t, err)
	defer func() { _ = tx.Rollback() }()

	_, err = RecordPendingJobs(ctx, tx, []DeleteRequest{
		{PackageID: 1, OrganizationID: 987653, UserNodeID: "N:user:u", TraceID: "t", Source: SourceOther},
	})
	require.Error(t, err)
}
