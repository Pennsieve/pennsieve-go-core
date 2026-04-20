package conflictStrategy

// Strategy controls how AddPackagesWithConflict resolves a name collision
// between a new upload and an existing non-deleted package under the same
// (dataset_id, parent_id, name) tuple.
type Strategy string

const (
	// KeepBoth appends " (N)" to the new upload's name so the existing
	// package survives unchanged. This is the legacy behavior.
	KeepBoth Strategy = "KEEP_BOTH"

	// Replace soft-deletes the existing package (state -> DELETING, name
	// prefixed with __DELETED__<nodeId>_) before inserting the new one,
	// and records the provenance link via replaces_package_id /
	// replaced_by_package_id. Folders (Collection type) cannot be replaced
	// — the DB CHECK constraint enforces this.
	Replace Strategy = "REPLACE"

	// Fail returns an error listing the conflicting names without inserting
	// anything. Lets the caller resolve the conflict out-of-band.
	Fail Strategy = "FAIL"
)