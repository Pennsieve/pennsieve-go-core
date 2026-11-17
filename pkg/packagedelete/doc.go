// Package packagedelete lets Go services ask process-jobs-service to
// delete a package. Every delete is a durable job: a row in the
// organization's delete_jobs table plus a DeletePackageJob message on
// the jobs queue, both carrying the same job id.
//
// How to use it:
//
//  1. Inside the transaction that sets the package(s) to DELETING, call
//     RecordPendingJobs with that transaction. It writes one PENDING
//     delete_jobs row per request and returns the jobs.
//  2. Commit.
//  3. Call EnqueueJobs with the returned jobs.
//
// If the enqueue is lost, the committed PENDING row is what the monitor
// and re-drive use to find it, so never enqueue a job without first
// recording it.
//
// Replace flow (a new file is taking over an existing name): call
// pgdb.Queries.PrepareReplace (or AddPackagesWithConflict with
// conflictStrategy.Replace, which calls it) inside your transaction.
// That renames the old row and marks it DELETING, which frees the name
// so your new insert can use it. Then, still in that transaction, call
// RecordPendingJobs with Source SourceUploadReplace for the replaced
// packages; commit; then EnqueueJobs.
//
// Note: pennsieve-api writes the same row and sends the same message to
// the same queue.
//
// The delete_jobs table is created by pennsieve-db-migrations
// V20261117092140__create_delete_jobs.sql; it must exist before this
// code runs.
package packagedelete
