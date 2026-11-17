package pgdb

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/lib/pq"
)

// Values for delete_jobs.source: which producer path created the job. The
// column is a plain VARCHAR (no CHECK) so a new producer does not need a schema
// migration; use a constant here so values stay consistent.
const (
	// DeleteJobSourceAPIDelete is a user delete through pennsieve-api (/data/delete).
	DeleteJobSourceAPIDelete = "API_DELETE"
	// DeleteJobSourceUploadReplace is the predecessor of an upload that replaced
	// it (conflictStrategy.Replace).
	DeleteJobSourceUploadReplace = "UPLOAD_REPLACE"

	deleteJobSourceMaxLen = 50
)

// DeleteJobParams describes one package-delete job to record.
//
// All of PackageID, UserNodeID, TraceID and Source are required. The consumer
// (process-jobs-service) does not use UserNodeID yet, but the job record is the
// durable audit row and the completion changelog event will need it; TraceID is
// used in the consumer's logs and delete records.
type DeleteJobParams struct {
	PackageID  int64
	UserNodeID string
	TraceID    string
	Source     string
}

// DeleteJob is a recorded job: the ID to put in the DeletePackageJob message
// (packagedelete.DeleteRequest.JobID) and the package it is for.
type DeleteJob struct {
	ID        string
	PackageID int64
}

// CreateDeleteJobs records one PENDING delete_jobs row per params entry, and
// returns the minted job IDs in the same order.
//
// This is the producer half of the delete-job contract (see
// pkg/packagedelete/README.md): the producer writes the PENDING row in the same
// transaction that flips the package to DELETING, THEN enqueues the
// DeletePackageJob carrying the returned ID. If the enqueue is lost, the row
// stays PENDING and is detectable and re-drivable; without the row, the lost
// delete is invisible.
//
// Call it on a Queries created WithTx for the caller's own transaction, after
// the DELETING flip and before commit. It uses the organization schema on the
// connection's search_path, like the rest of this package. It is one INSERT
// statement, but if it reports that some packages were not found the rows for
// the others are already written, so the caller must roll back its transaction
// on any error.
//
// The delete_jobs table must exist in the organization schema (migration in
// pennsieve-db-migrations). Inside a Postgres transaction a failed statement
// aborts the whole transaction, so a missing table fails the caller's
// transaction rather than being skippable: apply the migration before
// deploying a producer that calls this.
func (q *Queries) CreateDeleteJobs(ctx context.Context, jobs []DeleteJobParams) ([]DeleteJob, error) {
	if len(jobs) == 0 {
		return nil, nil
	}

	ids := make([]string, len(jobs))
	packageIDs := make([]int64, len(jobs))
	userNodeIDs := make([]string, len(jobs))
	traceIDs := make([]string, len(jobs))
	sources := make([]string, len(jobs))
	out := make([]DeleteJob, len(jobs))

	for i, j := range jobs {
		if err := validateDeleteJobParams(j); err != nil {
			return nil, fmt.Errorf("CreateDeleteJobs: job %d: %w", i, err)
		}
		id := uuid.NewString()
		ids[i] = id
		packageIDs[i] = j.PackageID
		userNodeIDs[i] = j.UserNodeID
		traceIDs[i] = j.TraceID
		sources[i] = j.Source
		out[i] = DeleteJob{ID: id, PackageID: j.PackageID}
	}

	// package_node_id and dataset_id come from packages in the same statement,
	// so the producer cannot record a job for a package that does not exist.
	const query = `
		INSERT INTO delete_jobs
		       (id, package_id, package_node_id, dataset_id, user_node_id, trace_id, source, state)
		SELECT j.id, p.id, p.node_id, p.dataset_id, j.user_node_id, j.trace_id, j.source, 'PENDING'
		  FROM unnest($1::uuid[], $2::bigint[], $3::text[], $4::text[], $5::text[])
		       AS j(id, package_id, user_node_id, trace_id, source)
		  JOIN packages p ON p.id = j.package_id`

	res, err := q.db.ExecContext(ctx, query,
		pq.Array(ids), pq.Array(packageIDs), pq.Array(userNodeIDs), pq.Array(traceIDs), pq.Array(sources))
	if err != nil {
		return nil, fmt.Errorf("CreateDeleteJobs: inserting %d job(s): %w", len(jobs), err)
	}
	n, err := res.RowsAffected()
	if err != nil {
		return nil, fmt.Errorf("CreateDeleteJobs: checking rows affected: %w", err)
	}
	if int(n) != len(jobs) {
		return nil, fmt.Errorf("CreateDeleteJobs: recorded %d of %d job(s): some packages were not found; roll back", n, len(jobs))
	}
	return out, nil
}

func validateDeleteJobParams(j DeleteJobParams) error {
	if j.PackageID <= 0 {
		return errors.New("PackageID must be > 0")
	}
	if j.UserNodeID == "" {
		return errors.New("UserNodeID is required")
	}
	if j.TraceID == "" {
		return errors.New("TraceID is required")
	}
	if j.Source == "" {
		return errors.New("Source is required")
	}
	if len(j.Source) > deleteJobSourceMaxLen {
		return fmt.Errorf("Source must be at most %d characters", deleteJobSourceMaxLen)
	}
	return nil
}
