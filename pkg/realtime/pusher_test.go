package realtime

import (
	"context"
	"errors"
	"testing"

	"github.com/pennsieve/pennsieve-go-core/pkg/domain"
	"github.com/pusher/pusher-http-go/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type fakePusher struct {
	triggers []pusher.Event
	batches  [][]pusher.Event
	err      error
}

func (f *fakePusher) Trigger(channel, name string, data interface{}) error {
	f.triggers = append(f.triggers, pusher.Event{Channel: channel, Name: name, Data: data})
	return f.err
}

func (f *fakePusher) TriggerBatch(batch []pusher.Event) (*pusher.TriggerBatchChannelsList, error) {
	f.batches = append(f.batches, batch)
	return nil, f.err
}

var _ domain.PusherAPI = (*fakePusher)(nil)

func TestPusherPublishMapsChannelsAndKeepsPayload(t *testing.T) {
	f := &fakePusher{}
	p, err := NewPusher(f)
	require.NoError(t, err)

	data := map[string]string{"runId": runUUID, "status": "SUCCEEDED"}
	require.NoError(t, p.Publish(context.Background(), Run(orgUUID, userUUID, runUUID), "workflow-run-status", data))

	require.Len(t, f.triggers, 1)
	assert.Equal(t, "organization-"+orgUUID+"-analytics", f.triggers[0].Channel)
	assert.Equal(t, "workflow-run-status", f.triggers[0].Name)
	assert.Equal(t, data, f.triggers[0].Data, "no AppSync envelope on Pusher")
}

func TestPusherPublishBatchChunksByTen(t *testing.T) {
	f := &fakePusher{}
	p, _ := NewPusher(f)
	var events []Event
	for i := 0; i < 12; i++ {
		events = append(events, Event{Channel: Dataset(dsUUID), Name: "upload-event", Data: i})
	}
	require.NoError(t, p.PublishBatch(context.Background(), events))
	require.Len(t, f.batches, 2)
	assert.Len(t, f.batches[0], 10)
	assert.Len(t, f.batches[1], 2)
	assert.Equal(t, "dataset-"+dsUUID, f.batches[0][0].Channel)
}

func TestPusherErrorsPropagate(t *testing.T) {
	f := &fakePusher{err: errors.New("boom")}
	p, _ := NewPusher(f)
	assert.Error(t, p.Publish(context.Background(), Dataset(dsUUID), "upload-event", nil))
	assert.Error(t, p.Publish(context.Background(), RunScope(orgUUID, ""), "x", nil), "wildcards are rejected before Pusher")
	_, err := NewPusher(nil)
	assert.Error(t, err)
}
