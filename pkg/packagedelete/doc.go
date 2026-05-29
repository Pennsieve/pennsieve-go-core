// Package packagedelete lets Go services ask process-jobs-service to
// delete a package. It builds the message and hands it off
//
// How to use it:
//
//   - Plain delete: build a []DeleteRequest and call DeletePackages.
//
//   - Replace flow (a new file is taking over an existing name): first
//     call pgdb.Queries.PrepareReplace inside your transaction. That
//     renames the old row and marks it as deleting, which frees the name
//     so your new insert can use it. After you commit, call
//     DeletePackages to enqueue the actual delete.
//
// Note: pennsieve-api sends the same message to the same queue
package packagedelete
