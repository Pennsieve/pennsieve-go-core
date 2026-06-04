package packagedelete

import (
	"context"
	"encoding/json"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// wireMessage decodes the JSON DeletePackages produces — used by tests
// that need to read fields back out of a sent message. Mirrors the
// shape: DeleteRequest's fields plus the per-message Id.
type wireMessage struct {
	DeleteRequest
	Id string `json:"id"`
}

// fakeQueueSender records every message it's handed and counts how many
// times it was called. Set fail to make SendToQueue return an error.
type fakeQueueSender struct {
	bodies [][]byte
	calls  int
	fail   bool
}

func newFakeQueueSender() *fakeQueueSender {
	return &fakeQueueSender{}
}

func (f *fakeQueueSender) SendToQueue(_ context.Context, bodies [][]byte) error {
	f.calls++
	if f.fail {
		return errors.New("fake: send boom")
	}
	for _, b := range bodies {
		// Copy the bytes — the caller may reuse the slice.
		cp := make([]byte, len(b))
		copy(cp, b)
		f.bodies = append(f.bodies, cp)
	}
	return nil
}

func validRequest() DeleteRequest {
	return DeleteRequest{
		PackageID:      123,
		OrganizationID: 5,
		UserNodeID:     "N:user:00000000-0000-0000-0000-000000000001",
		TraceID:        "trace-abc",
	}
}

// TestDeletePackages_EnvelopeShape checks the JSON we send: the keys,
// the types, and the {"DeletePackageJob": {...}} wrapper. If anyone
// renames a field or changes the shape, this fails.
func TestDeletePackages_EnvelopeShape(t *testing.T) {
	p := newFakeQueueSender()
	err := DeletePackages(context.Background(), p, []DeleteRequest{validRequest()})
	require.NoError(t, err)
	require.Len(t, p.bodies, 1)

	// Decode into a map so we check the JSON shape directly rather than
	// trusting our own Go struct.
	var outer map[string]json.RawMessage
	require.NoError(t, json.Unmarshal(p.bodies[0], &outer))
	require.Contains(t, outer, "DeletePackageJob",
		"payload must be wrapped under DeletePackageJob")
	require.Len(t, outer, 1, "envelope must have exactly one top-level key")

	var inner map[string]any
	require.NoError(t, json.Unmarshal(outer["DeletePackageJob"], &inner))

	// Field names + casing.
	for _, k := range []string{"packageId", "organizationId", "userId", "traceId", "id"} {
		assert.Contains(t, inner, k, "missing field %q", k)
	}
	assert.Len(t, inner, 5, "extra/missing fields: %v", inner)

	// Types. JSON numbers decode to float64 in map[string]any.
	assert.IsType(t, float64(0), inner["packageId"])
	assert.IsType(t, float64(0), inner["organizationId"])
	assert.IsType(t, "", inner["userId"])
	assert.IsType(t, "", inner["traceId"])
	assert.IsType(t, "", inner["id"])

	// Values come from the request.
	assert.EqualValues(t, 123, inner["packageId"])
	assert.EqualValues(t, 5, inner["organizationId"])
	assert.Equal(t, "N:user:00000000-0000-0000-0000-000000000001", inner["userId"])
	assert.Equal(t, "trace-abc", inner["traceId"])

	// id is a fresh UUID.
	idStr, _ := inner["id"].(string)
	_, err = uuid.Parse(idStr)
	assert.NoError(t, err, "id must be a valid UUID, got %q", idStr)
}

// TestDeletePackages_OneBatchCall checks that all requests are handed to
// the sender in a single call (so the sender can batch), each with a
// unique id.
func TestDeletePackages_OneBatchCall(t *testing.T) {
	p := newFakeQueueSender()
	reqs := []DeleteRequest{
		{PackageID: 1, OrganizationID: 5, UserNodeID: "N:user:a", TraceID: "t"},
		{PackageID: 2, OrganizationID: 5, UserNodeID: "N:user:a", TraceID: "t"},
		{PackageID: 3, OrganizationID: 5, UserNodeID: "N:user:a", TraceID: "t"},
	}
	require.NoError(t, DeletePackages(context.Background(), p, reqs))
	assert.Equal(t, 1, p.calls, "all messages should go in a single SendToQueue call")
	require.Len(t, p.bodies, 3)

	ids := map[string]struct{}{}
	for _, b := range p.bodies {
		var msg map[string]wireMessage
		require.NoError(t, json.Unmarshal(b, &msg))
		ids[msg["DeletePackageJob"].Id] = struct{}{}
	}
	assert.Len(t, ids, 3, "each message should get a distinct id")
}

func TestDeletePackages_NilQueueSender(t *testing.T) {
	err := DeletePackages(context.Background(), nil, []DeleteRequest{validRequest()})
	assert.Error(t, err)
}

// TestDeletePackages_CollectsAllErrors checks that every invalid request
// is reported, not just the first — and the valid one in between still
// gets sent.
func TestDeletePackages_CollectsAllErrors(t *testing.T) {
	p := newFakeQueueSender()
	reqs := []DeleteRequest{
		{PackageID: 0, OrganizationID: 5, UserNodeID: "N:user:a", TraceID: "t"},   // bad: PackageID
		{PackageID: 1, OrganizationID: 5, UserNodeID: "N:user:a", TraceID: "t"},   // good
		{PackageID: 2, OrganizationID: 5, UserNodeID: "", TraceID: "t"},           // bad: UserNodeID
	}
	err := DeletePackages(context.Background(), p, reqs)
	require.Error(t, err)
	// Both bad requests should be named in the joined error.
	assert.Contains(t, err.Error(), "request 0")
	assert.Contains(t, err.Error(), "request 2")
	// The good one in the middle still went out.
	assert.Len(t, p.bodies, 1, "the valid request should still be sent")
}

func TestDeletePackages_EmptySlice(t *testing.T) {
	p := newFakeQueueSender()
	assert.NoError(t, DeletePackages(context.Background(), p, nil))
	assert.Empty(t, p.bodies)
}

// TestDeletePackages_ReturnsSendError checks that a sender failure is
// surfaced. Per-message partial-failure handling is the sender's job now
// (e.g. SQS batch entry failures); from here a failed send is one error.
func TestDeletePackages_ReturnsSendError(t *testing.T) {
	p := newFakeQueueSender()
	p.fail = true
	reqs := []DeleteRequest{
		{PackageID: 1, OrganizationID: 5, UserNodeID: "N:user:a", TraceID: "t"},
		{PackageID: 2, OrganizationID: 5, UserNodeID: "N:user:a", TraceID: "t"},
	}
	err := DeletePackages(context.Background(), p, reqs)
	assert.Error(t, err, "send failure should be returned")
}

// TestDeletePackages_ValidationFailsPerRequest checks that an invalid
// request is reported but doesn't block the valid ones around it.
func TestDeletePackages_ValidationFailsPerRequest(t *testing.T) {
	tests := []struct {
		name string
		mod  func(*DeleteRequest)
	}{
		{"zero PackageID", func(r *DeleteRequest) { r.PackageID = 0 }},
		{"negative PackageID", func(r *DeleteRequest) { r.PackageID = -1 }},
		{"zero OrganizationID", func(r *DeleteRequest) { r.OrganizationID = 0 }},
		{"empty UserNodeID", func(r *DeleteRequest) { r.UserNodeID = "" }},
		{"empty TraceID", func(r *DeleteRequest) { r.TraceID = "" }},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bad := validRequest()
			tt.mod(&bad)
			good := validRequest()
			good.PackageID = 999

			p := newFakeQueueSender()
			err := DeletePackages(context.Background(), p, []DeleteRequest{bad, good})
			assert.Error(t, err, "the bad request should return an error")
			assert.Len(t, p.bodies, 1, "the good request should still be sent")

			var msg map[string]wireMessage
			require.NoError(t, json.Unmarshal(p.bodies[0], &msg))
			assert.Equal(t, int64(999), msg["DeletePackageJob"].PackageID,
				"the sent message should be the valid request")
		})
	}
}

// TestDeletePackages_GoldenEnvelope pins the JSON shape with a fixed
// fixture. The id is generated per message, so we drop it before
// comparing.
func TestDeletePackages_GoldenEnvelope(t *testing.T) {
	p := newFakeQueueSender()
	require.NoError(t, DeletePackages(context.Background(), p, []DeleteRequest{{
		PackageID:      42,
		OrganizationID: 7,
		UserNodeID:     "N:user:abc",
		TraceID:        "trace-xyz",
	}}))
	require.Len(t, p.bodies, 1)

	var got map[string]map[string]any
	require.NoError(t, json.Unmarshal(p.bodies[0], &got))
	delete(got["DeletePackageJob"], "id")

	want := map[string]map[string]any{
		"DeletePackageJob": {
			"packageId":      float64(42),
			"organizationId": float64(7),
			"userId":         "N:user:abc",
			"traceId":        "trace-xyz",
		},
	}
	assert.Equal(t, want, got)
}
