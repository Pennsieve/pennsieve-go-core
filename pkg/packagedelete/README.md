# packagedelete

How a package gets deleted on Pennsieve, and how Go producers hand a delete to
the consumer. Read this before writing a new producer or changing the message,
the `DELETING` flip, or the `delete_jobs` record.

## The pieces

```
 producer                                   queue                 consumer
 ───────────────────────────────           ─────────────         ───────────────────────────────
 pennsieve-api  /data/delete        ┐                            process-jobs-service
   (PackageManager.delete)          ├─ DeletePackageJob ───────►  (package-delete Lambda,
 upload-service-v2  (replace flow)  ┘   jobs_queue_v2              Fargate for huge deletes)
   via this package + pgdb
```

- **Producers** mark the package `DELETING` (and rename it), record the job,
  then enqueue one `DeletePackageJob` message per package.
- **The consumer**, `Pennsieve/process-jobs-service`, does the real work:
  gathers the package's subtree and files, deletes the S3 objects of unpublished
  files, records the deleted objects in DynamoDB (so restore can find them),
  flips the packages to `DELETED`, and decrements the storage counters.
- This package (`DeletePackages`) is the Go producer's half of the *message*.
  `pgdb.Queries.CreateDeleteJobs` is the Go producer's half of the *job record*.
  `pennsieve-api` (Scala) does the same two things itself.

## The message

```json
{ "DeletePackageJob": {
    "packageId": 123, "organizationId": 5,
    "userId": "N:user:...", "traceId": "...", "id": "<uuid>" } }
```

| field | meaning |
|---|---|
| `packageId`, `organizationId` | the package to delete and its organization (schema) |
| `userId` | the acting user's **node id**. The consumer does not use it yet; the job record keeps it and the planned completion changelog event needs it |
| `traceId` | any string that helps follow the job in the consumer's logs (upload uses the manifest id, the API its request trace id). The consumer logs it and writes it into its DynamoDB delete records. **Required** |
| `id` | the job's UUID. It is the primary key of the job's `delete_jobs` row, so a redelivered message finds its own record |

`DeletePackages` validates every request and collects errors per request, so one
bad request does not drop the others. If `DeleteRequest.JobID` is empty it mints
the `id`; producers that record a `delete_jobs` row must pass the row's id.

## What the producer must do, in order

1. In **one transaction**: flip the package to `DELETING`, rename it
   (see below), and record the job with `CreateDeleteJobs` (a `PENDING` row).
2. **Commit.**
3. Send the messages with `DeletePackages`, passing each job's id as `JobID`.

```go
tx, _ := db.BeginTx(ctx, nil)
q := pgdb.New(tx) // org schema on the connection's search_path

// replace flow: this flips and renames the predecessors, inside tx
inserted, err := q.AddPackagesWithConflict(ctx, params, conflictStrategy.Replace)

var jobs []pgdb.DeleteJobParams
for _, p := range inserted {
    if p.ReplacesPackageId.Valid {
        jobs = append(jobs, pgdb.DeleteJobParams{
            PackageID: p.ReplacesPackageId.Int64, UserNodeID: user.NodeId,
            TraceID: manifestID, Source: pgdb.DeleteJobSourceUploadReplace})
    }
}
created, err := q.CreateDeleteJobs(ctx, jobs) // PENDING rows, same tx

err = tx.Commit()
// ...then, after commit:
reqs := make([]packagedelete.DeleteRequest, len(created))
for i, j := range created {
    reqs[i] = packagedelete.DeleteRequest{PackageID: j.PackageID, OrganizationID: orgID,
        UserNodeID: user.NodeId, TraceID: manifestID, JobID: j.ID}
}
err = packagedelete.DeletePackages(ctx, sender, reqs)
```

**Why the row is written in the transaction, before the enqueue.** The database
flip and the SQS send cannot be made atomic, and we deliberately do not tie them
together. If the send is lost after the commit, the package is `DELETING` with no
message: nothing deletes it, its storage stays counted, and it is invisible. With
the `PENDING` row committed alongside the flip, the same failure leaves an aged
`PENDING` row, which the monitor alarms on and an operator can re-drive.

**Do not skip the row "because the enqueue will work".** A producer that sends a
message with no row is exactly the case the consumer's strict mode rejects.

## The `delete_jobs` record

One row per job, in each organization schema (migration:
`pennsieve-db-migrations`, `migrations/organization`). Key columns: `id` (the
message id), `package_id`, `package_node_id`, `dataset_id`, `user_node_id`,
`trace_id`, `source` (which producer: `API_DELETE`, `UPLOAD_REPLACE`), `state`,
`attempts`, `last_error`, and progress counters.

```
PENDING ──► IN_PROGRESS ──► SUCCEEDED
                │
                ├────────► PARTIAL ──► (re-drive) ──► IN_PROGRESS
                └────────► FAILED  ──► (re-drive) ──► IN_PROGRESS
```

- **Producer** writes `PENDING`. **Consumer** moves it to `IN_PROGRESS` on
  receipt (incrementing `attempts`) and records the outcome.
- `PARTIAL` is distinct from `FAILED`: the job committed but did not finish
  everything (today: some S3 keys failed to delete).
- The consumer's tracking and its monitor/alarms are Pennsieve/process-jobs-service
  PR #29; the consumer is lenient until strict mode is turned on (below).

### Ordering and rollout

- **The table must exist before a producer that calls `CreateDeleteJobs` is
  deployed.** Inside a Postgres transaction a failed statement aborts the whole
  transaction, so a missing table fails the caller's transaction (for the replace
  flow: the import) rather than being skippable.
- A producer may safely reach an environment *before* the consumer: the consumer
  simply finds the `PENDING` row and moves it to `IN_PROGRESS`.
- Once all producers write rows, the consumer's strict mode (a
  `REQUIRE_JOB_RECORD` switch, **planned**, requested in process-jobs-service#29) is
  turned on. Then a message with no row is rejected (distinct metric, DLQ, alarm)
  instead of being run untracked.

## The `DELETING` flip and the rename contract

Producers soft-delete before enqueueing, so the name is free for reuse
immediately and the user sees the item vanish:

- `state = DELETING`
- `name = '__DELETED__' || node_id || '_' || name`. **The `__DELETED__` prefix is
  a contract with the restore flow** (packages-service `GetOriginalName` requires
  it on every `DELETED` package).
- `pennsieve-api` does this for the **entire subtree** in one transaction
  (pennsieve-api#399), skipping descendants that are already `DELETING`/`DELETED`.
- `pgdb.Queries.PrepareReplace` does it for the single predecessor of a replace.
  It refuses a package that is already `DELETING`/`DELETED` (so a duplicate replace
  cannot rename twice).
- The consumer repeats the rename as an idempotent backstop (it only renames
  names that do not already start with the marker).

Deleting a package that is already `DELETING` is an idempotent success in the API,
not a second delete.

## Replace flow (upload-service-v2)

Upload with `onConflict = replace` does **not** overwrite in place. For each file
whose name collides with an existing, non-deleted package:

1. the predecessor is flipped to `DELETING` and renamed (`PrepareReplace`);
2. a **new** package is inserted with `replaces_package_id` = predecessor;
3. the predecessor's `replaced_by_package_id` = the new package.

`replaces_package_id` / `replaced_by_package_id` are **provenance only**. The
predecessor is still deleted by an ordinary `DeletePackageJob`, so a replace
produces one job (and one `delete_jobs` row, `source = UPLOAD_REPLACE`) per
replaced file. Only files can be replaced (a DB check forbids collections).

## Consumer behaviour worth knowing

- **Idempotent on redelivery.** SQS is at-least-once. The consumer's flip to
  `DELETED` skips rows already `DELETED`, and only the rows it actually flipped
  get their storage decremented. A job already `SUCCEEDED` is acked untouched.
- **Large deletes** (more files than `FILE_PROCESSING_LIMIT`) are handed to a
  Fargate task; the job stays `IN_PROGRESS` until the task records its outcome.
- **Storage counters** track the workspace footprint regardless of publish
  status: a published file's S3 object is retained, but its bytes are still
  decremented.

## Tests

`go test ./pkg/packagedelete` needs nothing. The `pgdb` tests (including
`CreateDeleteJobs`) need the Docker Compose stack, `make test`. Until the CI seed
database includes the `delete_jobs` migration, the tests create the table
themselves (`delete_jobs_test.go`, marked TODO).
