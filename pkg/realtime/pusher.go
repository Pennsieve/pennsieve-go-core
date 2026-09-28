package realtime

import (
	"context"
	"errors"

	"github.com/pennsieve/pennsieve-go-core/pkg/domain"
	"github.com/pusher/pusher-http-go/v5"
)

// Pusher's batch trigger takes at most 10 events per call.
const maxPusherBatch = 10

// Pusher publishes through an existing Pusher client, mapping channels back
// to the names today's Pusher subscribers use (dataset-<uuid>,
// organization-<uuid>-analytics, user-<uuid>-analytics, application-<uuid>).
// It keeps environments on Pusher working while publishers move to this
// package. Payloads are sent as-is, without the AppSync envelope, because
// Pusher carries the event name itself.
type Pusher struct {
	client domain.PusherAPI
}

func NewPusher(client domain.PusherAPI) (*Pusher, error) {
	if client == nil {
		return nil, errors.New("realtime: Pusher client is required")
	}
	return &Pusher{client: client}, nil
}

func (p *Pusher) Publish(_ context.Context, channel Channel, name string, data any) error {
	pusherName, err := pusherChannel(channel)
	if err != nil {
		return err
	}
	return p.client.Trigger(pusherName, name, data)
}

func (p *Pusher) PublishBatch(_ context.Context, events []Event) error {
	batch := make([]pusher.Event, 0, len(events))
	for _, e := range events {
		channel, err := pusherChannel(e.Channel)
		if err != nil {
			return err
		}
		batch = append(batch, pusher.Event{Channel: channel, Name: e.Name, Data: e.Data})
	}
	for start := 0; start < len(batch); start += maxPusherBatch {
		end := start + maxPusherBatch
		if end > len(batch) {
			end = len(batch)
		}
		if _, err := p.client.TriggerBatch(batch[start:end]); err != nil {
			return err
		}
	}
	return nil
}
