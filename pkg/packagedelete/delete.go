package packagedelete

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"strconv"

	"github.com/google/uuid"
	"github.com/lib/pq"
)

// QueueSender sends batch messages to the delete queue. Caller supplies
// the implementation for loading messages on the queue.
// EnqueueJobs calls SendToQueue once per call.
type QueueSender interface {
	SendToQueue(ctx context.Context, bodies [][]byte) error
}

// DBTX is the slice of *sql.Tx (or *sql.DB) RecordPendingJobs needs.
// Pass the transaction that flips the packages to DELETING.
type DBTX interface {
	QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error)
}

// Source names the producer of a delete job. It is stored in
// delete_jobs.source so a stuck row can be traced back to its caller.
// The allowed values are fixed by the CHECK constraint in
// pennsieve-db-migrations V20261117092140__create_delete_jobs.sql.
// UNKNOWN is reserved for rows the consumer creates itself, so it is
// not offered here.
type Source string

const (
	// SourceAPIDelete is a user-initiated delete (pennsieve-api).
	SourceAPIDelete Source = "API_DELETE"
	// SourceUploadReplace is a new upload replacing an existing package
	// (upload-service-v2).
	SourceUploadReplace Source = "UPLOAD_REPLACE"
	// SourceOther is any other Go producer.
	SourceOther Source = "OTHER"
)

func (s Source) valid() bool {
	switch s {
	case SourceAPIDelete, SourceUploadReplace, SourceOther:
		return true
	}
	return false
}

// DeleteRequest is a single top-level package to delete. Only the
// requested package gets a job; its descendants are handled by the
// consumer as part of that job.
//
// UserNodeID is the user's node id, e.g. "N:user:abc". (The wire field
// is named "userId" to match pennsieve-api — both sides carry a node id
// here.)
//
// TraceID is any string that helps trace this job in the consumer's
// logs. The upload service uses the manifest id; an API call would use
// its request trace id. Required.
//
// Source is written to delete_jobs.source and is not part of the queue
// message. Required.
type DeleteRequest struct {
	PackageID      int64  `json:"packageId"`
	OrganizationID int64  `json:"organizationId"`
	UserNodeID     string `json:"userId"`
	TraceID        string `json:"traceId"`
	Source         Source `json:"-"`
}

// Job is a delete request with the job id minted for it. ID is both the
// delete_jobs primary key and the "id" field of the queue message, which
// is how the consumer finds its row.
type Job struct {
	ID string
	DeleteRequest
}

// RecordPendingJobs mints a job id for each request and inserts a
// delete_jobs row in state PENDING (attempts 0) in the request's
// organization schema. package_node_id and dataset_id are read from the
// packages row in the same statement.
//
// Call it inside the same transaction that sets the packages to
// DELETING, before commit, so the flip and the job row land together.
// Each package must already be DELETING in that transaction; a request
// for a package that is missing or not DELETING is an error.
//
// It is all-or-nothing: any invalid request or failed insert returns an
// error and no jobs, and the caller should roll back. After commit,
// pass the returned jobs to EnqueueJobs.
func RecordPendingJobs(ctx context.Context, db DBTX, reqs []DeleteRequest) ([]Job, error) {
	if db == nil {
		return nil, errors.New("packagedelete: DBTX is nil")
	}
	if len(reqs) == 0 {
		return nil, nil
	}

	var errs []error
	type key struct{ org, pkg int64 }
	seen := make(map[key]int, len(reqs))
	for i, r := range reqs {
		if err := validate(r); err != nil {
			errs = append(errs, fmt.Errorf("request %d: %w", i, err))
			continue
		}
		k := key{r.OrganizationID, r.PackageID}
		if j, dup := seen[k]; dup {
			errs = append(errs, fmt.Errorf("request %d: duplicate of request %d (package %d)", i, j, r.PackageID))
			continue
		}
		seen[k] = i
	}
	if len(errs) > 0 {
		return nil, errors.Join(errs...)
	}

	jobs := make([]Job, len(reqs))
	byOrg := map[int64][]int{}
	for i, r := range reqs {
		jobs[i] = Job{ID: uuid.NewString(), DeleteRequest: r}
		byOrg[r.OrganizationID] = append(byOrg[r.OrganizationID], i)
	}

	// Deterministic order, so a multi-org call takes locks the same way
	// every time.
	orgs := make([]int64, 0, len(byOrg))
	for org := range byOrg {
		orgs = append(orgs, org)
	}
	sort.Slice(orgs, func(a, b int) bool { return orgs[a] < orgs[b] })

	for _, org := range orgs {
		if err := insertPending(ctx, db, org, jobs, byOrg[org]); err != nil {
			return nil, err
		}
	}
	return jobs, nil
}

// insertPending writes the rows for one organization in one statement.
// The schema is named explicitly rather than relying on search_path, so
// the row always lands next to the package it refers to.
func insertPending(ctx context.Context, db DBTX, org int64, jobs []Job, idx []int) error {
	schema := pq.QuoteIdentifier(strconv.FormatInt(org, 10))
	query := fmt.Sprintf(`
INSERT INTO %[1]s.delete_jobs
  (id, package_id, package_node_id, dataset_id, user_node_id, trace_id, source, state, attempts)
SELECT j.id, p.id, p.node_id, p.dataset_id, j.user_node_id, j.trace_id, j.source, 'PENDING', 0
FROM unnest($1::uuid[], $2::int[], $3::text[], $4::text[], $5::text[])
       AS j(id, package_id, user_node_id, trace_id, source)
JOIN %[1]s.packages p ON p.id = j.package_id
WHERE p.state = 'DELETING'
RETURNING package_id`, schema)

	ids := make([]string, len(idx))
	pkgs := make([]int64, len(idx))
	users := make([]string, len(idx))
	traces := make([]string, len(idx))
	sources := make([]string, len(idx))
	for n, i := range idx {
		ids[n] = jobs[i].ID
		pkgs[n] = jobs[i].PackageID
		users[n] = jobs[i].UserNodeID
		traces[n] = jobs[i].TraceID
		sources[n] = string(jobs[i].Source)
	}

	rows, err := db.QueryContext(ctx, query,
		pq.Array(ids), pq.Array(pkgs), pq.Array(users), pq.Array(traces), pq.Array(sources))
	if err != nil {
		return fmt.Errorf("packagedelete: inserting delete_jobs for organization %d: %w", org, err)
	}
	defer rows.Close()

	inserted := make(map[int64]struct{}, len(idx))
	for rows.Next() {
		var id int64
		if err := rows.Scan(&id); err != nil {
			return fmt.Errorf("packagedelete: reading inserted delete_jobs for organization %d: %w", org, err)
		}
		inserted[id] = struct{}{}
	}
	if err := rows.Err(); err != nil {
		return fmt.Errorf("packagedelete: inserting delete_jobs for organization %d: %w", org, err)
	}

	var missing []int64
	for _, p := range pkgs {
		if _, ok := inserted[p]; !ok {
			missing = append(missing, p)
		}
	}
	if len(missing) > 0 {
		return fmt.Errorf("packagedelete: organization %d: packages %v not found or not DELETING; "+
			"set them to DELETING in this transaction before recording jobs", org, missing)
	}
	return nil
}

// EnqueueJobs sends one DeletePackageJob message per job to the sender,
// in a single SendToQueue call. Call it after the transaction that ran
// RecordPendingJobs has committed. Each message's "id" is the job's ID,
// so the consumer picks up the row RecordPendingJobs wrote.
//
// If the send fails, the committed PENDING rows remain. That is by
// design: an aged PENDING row is the signal the monitor and re-drive
// use to find a lost enqueue.
func EnqueueJobs(ctx context.Context, sender QueueSender, jobs []Job) error {
	if sender == nil {
		return errors.New("packagedelete: QueueSender is nil")
	}
	if len(jobs) == 0 {
		return nil
	}

	bodies := make([][]byte, 0, len(jobs))
	for i, j := range jobs {
		if _, err := uuid.Parse(j.ID); err != nil {
			return fmt.Errorf("packagedelete: job %d: ID %q is not a UUID; use RecordPendingJobs to create jobs", i, j.ID)
		}
		body, err := json.Marshal(map[string]any{
			"DeletePackageJob": struct {
				DeleteRequest
				Id string `json:"id"`
			}{j.DeleteRequest, j.ID},
		})
		if err != nil {
			return fmt.Errorf("packagedelete: job %d: marshal: %w", i, err)
		}
		bodies = append(bodies, body)
	}

	if err := sender.SendToQueue(ctx, bodies); err != nil {
		return fmt.Errorf("packagedelete: send: %w", err)
	}
	return nil
}

func validate(r DeleteRequest) error {
	if r.PackageID <= 0 {
		return errors.New("PackageID must be > 0")
	}
	if r.OrganizationID <= 0 {
		return errors.New("OrganizationID must be > 0")
	}
	if r.UserNodeID == "" {
		return errors.New("UserNodeID is required")
	}
	if r.TraceID == "" {
		return errors.New("TraceID is required")
	}
	if !r.Source.valid() {
		return fmt.Errorf("Source %q is not one of API_DELETE, UPLOAD_REPLACE, OTHER", r.Source)
	}
	return nil
}
