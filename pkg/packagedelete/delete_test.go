package packagedelete

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// wireMessage decodes the JSON EnqueueJobs produces — used by tests
// that need to read fields back out of a sent message. Mirrors the
// shape: DeleteRequest's wire fields plus the per-message Id.
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

// panicDB fails the test if RecordPendingJobs reaches the database. Used
// to show that validation rejects a batch before any SQL runs.
type panicDB struct{ t *testing.T }

func (p panicDB) QueryContext(context.Context, string, ...any) (*sql.Rows, error) {
	p.t.Fatal("RecordPendingJobs should not query the database for an invalid batch")
	return nil, nil
}

func validRequest() DeleteRequest {
	return DeleteRequest{
		PackageID:      123,
		OrganizationID: 5,
		UserNodeID:     "N:user:00000000-0000-0000-0000-000000000001",
		TraceID:        "trace-abc",
		Source:         SourceAPIDelete,
	}
}

func validJob() Job {
	return Job{ID: uuid.NewString(), DeleteRequest: validRequest()}
}

// TestEnqueueJobs_EnvelopeShape checks the JSON we send: the keys, the
// types, and the {"DeletePackageJob": {...}} wrapper. If anyone renames
// a field, changes the shape, or leaks Source onto the wire, this fails.
func TestEnqueueJobs_EnvelopeShape(t *testing.T) {
	p := newFakeQueueSender()
	job := validJob()
	require.NoError(t, EnqueueJobs(context.Background(), p, []Job{job}))
	require.Len(t, p.bodies, 1)

	var outer map[string]json.RawMessage
	require.NoError(t, json.Unmarshal(p.bodies[0], &outer))
	require.Contains(t, outer, "DeletePackageJob",
		"payload must be wrapped under DeletePackageJob")
	require.Len(t, outer, 1, "envelope must have exactly one top-level key")

	var inner map[string]any
	require.NoError(t, json.Unmarshal(outer["DeletePackageJob"], &inner))

	for _, k := range []string{"packageId", "organizationId", "userId", "traceId", "id"} {
		assert.Contains(t, inner, k, "missing field %q", k)
	}
	assert.Len(t, inner, 5, "extra/missing fields (Source must not be sent): %v", inner)

	assert.IsType(t, float64(0), inner["packageId"])
	assert.IsType(t, float64(0), inner["organizationId"])
	assert.IsType(t, "", inner["userId"])
	assert.IsType(t, "", inner["traceId"])
	assert.IsType(t, "", inner["id"])

	assert.EqualValues(t, 123, inner["packageId"])
	assert.EqualValues(t, 5, inner["organizationId"])
	assert.Equal(t, "N:user:00000000-0000-0000-0000-000000000001", inner["userId"])
	assert.Equal(t, "trace-abc", inner["traceId"])
	assert.Equal(t, job.ID, inner["id"], "message id must be the job id (the delete_jobs primary key)")
}

// TestEnqueueJobs_GoldenEnvelope pins the full JSON with a fixed id.
func TestEnqueueJobs_GoldenEnvelope(t *testing.T) {
	p := newFakeQueueSender()
	require.NoError(t, EnqueueJobs(context.Background(), p, []Job{{
		ID: "6f1c2b0e-8f4e-4a77-9d43-2b8f1f5b7c11",
		DeleteRequest: DeleteRequest{
			PackageID:      42,
			OrganizationID: 7,
			UserNodeID:     "N:user:abc",
			TraceID:        "trace-xyz",
			Source:         SourceUploadReplace,
		},
	}}))
	require.Len(t, p.bodies, 1)
	assert.JSONEq(t,
		`{"DeletePackageJob":{"packageId":42,"organizationId":7,"userId":"N:user:abc","traceId":"trace-xyz","id":"6f1c2b0e-8f4e-4a77-9d43-2b8f1f5b7c11"}}`,
		string(p.bodies[0]))
}

// TestEnqueueJobs_OneBatchCall checks that all jobs are handed to the
// sender in a single call (so the sender can batch), each with its own id.
func TestEnqueueJobs_OneBatchCall(t *testing.T) {
	p := newFakeQueueSender()
	jobs := []Job{validJob(), validJob(), validJob()}
	require.NoError(t, EnqueueJobs(context.Background(), p, jobs))
	assert.Equal(t, 1, p.calls, "all messages should go in a single SendToQueue call")
	require.Len(t, p.bodies, 3)

	for i, b := range p.bodies {
		var msg map[string]wireMessage
		require.NoError(t, json.Unmarshal(b, &msg))
		assert.Equal(t, jobs[i].ID, msg["DeletePackageJob"].Id)
	}
}

func TestEnqueueJobs_NilQueueSender(t *testing.T) {
	assert.Error(t, EnqueueJobs(context.Background(), nil, []Job{validJob()}))
}

func TestEnqueueJobs_Empty(t *testing.T) {
	p := newFakeQueueSender()
	assert.NoError(t, EnqueueJobs(context.Background(), p, nil))
	assert.Zero(t, p.calls)
}

// TestEnqueueJobs_RejectsJobWithoutID guards against hand-built jobs
// that never went through RecordPendingJobs: they have no row, so
// sending them would break the consumer's strict mode.
func TestEnqueueJobs_RejectsJobWithoutID(t *testing.T) {
	p := newFakeQueueSender()
	bad := validJob()
	bad.ID = ""
	err := EnqueueJobs(context.Background(), p, []Job{validJob(), bad})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "job 1")
	assert.Zero(t, p.calls, "nothing should be sent if any job is malformed")
}

func TestEnqueueJobs_ReturnsSendError(t *testing.T) {
	p := newFakeQueueSender()
	p.fail = true
	assert.Error(t, EnqueueJobs(context.Background(), p, []Job{validJob(), validJob()}))
}

func TestRecordPendingJobs_NilDB(t *testing.T) {
	_, err := RecordPendingJobs(context.Background(), nil, []DeleteRequest{validRequest()})
	assert.Error(t, err)
}

func TestRecordPendingJobs_Empty(t *testing.T) {
	jobs, err := RecordPendingJobs(context.Background(), panicDB{t}, nil)
	assert.NoError(t, err)
	assert.Empty(t, jobs)
}

// TestRecordPendingJobs_ValidationRejectsWholeBatch checks that any
// invalid request fails the whole call before SQL runs. The caller is
// inside a transaction and should roll back, so a partial insert would
// be useless.
func TestRecordPendingJobs_ValidationRejectsWholeBatch(t *testing.T) {
	tests := []struct {
		name string
		mod  func(*DeleteRequest)
	}{
		{"zero PackageID", func(r *DeleteRequest) { r.PackageID = 0 }},
		{"negative PackageID", func(r *DeleteRequest) { r.PackageID = -1 }},
		{"zero OrganizationID", func(r *DeleteRequest) { r.OrganizationID = 0 }},
		{"empty UserNodeID", func(r *DeleteRequest) { r.UserNodeID = "" }},
		{"empty TraceID", func(r *DeleteRequest) { r.TraceID = "" }},
		{"empty Source", func(r *DeleteRequest) { r.Source = "" }},
		{"UNKNOWN Source is consumer-only", func(r *DeleteRequest) { r.Source = "UNKNOWN" }},
		{"made-up Source", func(r *DeleteRequest) { r.Source = "CLEANUP" }},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			good := validRequest()
			good.PackageID = 999
			bad := validRequest()
			tt.mod(&bad)

			jobs, err := RecordPendingJobs(context.Background(), panicDB{t}, []DeleteRequest{good, bad})
			require.Error(t, err)
			assert.Contains(t, err.Error(), "request 1")
			assert.Nil(t, jobs)
		})
	}
}

func TestRecordPendingJobs_RejectsDuplicatePackage(t *testing.T) {
	r := validRequest()
	_, err := RecordPendingJobs(context.Background(), panicDB{t}, []DeleteRequest{r, r})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "duplicate")
}
