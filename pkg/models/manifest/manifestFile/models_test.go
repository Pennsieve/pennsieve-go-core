package manifestFile

import "testing"

func TestIsInProgress(t *testing.T) {
	cases := []struct {
		status   Status
		expected string
	}{
		// Non-terminal — still expects automatic progress.
		{Local, "x"},
		{Registered, "x"},
		{Uploaded, "x"},
		{Changed, "x"},
		{Unknown, "x"},

		// Terminal — no background worker will change this automatically.
		{Imported, ""},
		{Finalized, ""},
		{Verified, ""},
		{Removed, ""},
		{Failed, ""},
		{FailedOrphan, ""},
	}
	for _, c := range cases {
		if got := c.status.IsInProgress(); got != c.expected {
			t.Errorf("IsInProgress(%s) = %q, want %q", c.status, got, c.expected)
		}
	}
}

func TestStatusStringRoundTrip(t *testing.T) {
	// Every defined Status must round-trip via String() -> ManifestFileStatusMap.
	all := []Status{Local, Registered, Imported, Finalized, Verified, Failed, Removed, Unknown, Changed, Uploaded, FailedOrphan}
	for _, s := range all {
		parsed := Status(0).ManifestFileStatusMap(s.String())
		if parsed != s {
			t.Errorf("round-trip failed: %s -> %q -> %s", s, s.String(), parsed)
		}
	}
}

func TestManifestFileStatusMapUnknownFallback(t *testing.T) {
	// Unknown strings must map to Unknown, not Local. Previously mapping to
	// Local caused callers (notably the agent) to misinterpret a
	// future-version status as "not yet synced" and queue a spurious
	// re-upload.
	if got := Status(0).ManifestFileStatusMap("SomeFutureStatus"); got != Unknown {
		t.Errorf("expected Unknown fallback for unknown string, got %s", got)
	}
	if got := Status(0).ManifestFileStatusMap(""); got != Unknown {
		t.Errorf("expected Unknown fallback for empty string, got %s", got)
	}
}