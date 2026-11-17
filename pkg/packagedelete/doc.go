// Package packagedelete lets Go services ask process-jobs-service to
// delete a package. It builds the message and hands it off.
//
// See README.md in this directory for the full picture: the message, the
// delete_jobs record, the DELETING/rename contract, and rollout order.
//
// How to use it:
//
//   - Plain delete: record the job, then enqueue. Inside the transaction that
//     marks the package DELETING (and renames it), call
//     pgdb.Queries.CreateDeleteJobs to write a PENDING delete_jobs row. After
//     commit, build a []DeleteRequest with each job's id as JobID and call
//     DeletePackages.
//
//   - Replace flow (a new file is taking over an existing name): call
//     pgdb.Queries.AddPackagesWithConflict with conflictStrategy.Replace, which
//     renames the old row and marks it as deleting, frees the name, inserts the
//     new package, and links the two. In the same transaction, call
//     CreateDeleteJobs for each package with ReplacesPackageId set. After
//     commit, call DeletePackages with those job ids.
//     (Lower level: pgdb.Queries.PrepareReplace does only the make-room step.)
//
// Writing the delete_jobs row in the same transaction as the DELETING flip is
// what makes a lost enqueue detectable. Leaving JobID empty still works (a
// UUID is minted) but sends a job with no record.
//
// Note: pennsieve-api sends the same message to the same queue
package packagedelete
