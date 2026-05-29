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

// fakeQueueSender records every message it's asked to send. Set failAt
// to a call index (0-based) to make that one call return an error.
type fakeQueueSender struct {
	bodies   [][]byte
	failAt   int  // -1 = never fail
	gotCalls int
}

func newFakeQueueSender() *fakeQueueSender {
	return &fakeQueueSender{failAt: -1}
}

func (f *fakeQueueSender) SendToQueue(_ context.Context, body []byte) error {
	idx := f.gotCalls
	f.gotCalls++
	if idx == f.failAt {
		return errors.New("fake: send boom")
	}
	// Copy the bytes — the caller may reuse the slice.
	cp := make([]byte, len(body))
	copy(cp, body)
	f.bodies = append(f.bodies, cp)
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

// TestDeletePackages_SendsEach checks that each request results in one
// SendToQueue call, with a unique id per message.
func TestDeletePackages_SendsEach(t *testing.T) {
	p := newFakeQueueSender()
	reqs := []DeleteRequest{
		{PackageID: 1, OrganizationID: 5, UserNodeID: "N:user:a", TraceID: "t"},
		{PackageID: 2, OrganizationID: 5, UserNodeID: "N:user:a", TraceID: "t"},
		{PackageID: 3, OrganizationID: 5, UserNodeID: "N:user:a", TraceID: "t"},
	}
	require.NoError(t, DeletePackages(context.Background(), p, reqs))
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

func TestDeletePackages_EmptySlice(t *testing.T) {
	p := newFakeQueueSender()
	assert.NoError(t, DeletePackages(context.Background(), p, nil))
	assert.Empty(t, p.bodies)
}

// TestDeletePackages_ContinuesPastSendError checks that one bad send
// doesn't stop the rest of the batch.
func TestDeletePackages_ContinuesPastSendError(t *testing.T) {
	p := newFakeQueueSender()
	p.failAt = 0 // first call fails
	reqs := []DeleteRequest{
		{PackageID: 1, OrganizationID: 5, UserNodeID: "N:user:a", TraceID: "t"},
		{PackageID: 2, OrganizationID: 5, UserNodeID: "N:user:a", TraceID: "t"},
	}
	err := DeletePackages(context.Background(), p, reqs)
	assert.Error(t, err, "first send error should be returned")
	assert.Len(t, p.bodies, 1, "second request should still have been sent")
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
