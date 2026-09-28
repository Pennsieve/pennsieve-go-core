package realtime

import "context"

// Event is one live event: a name subscribers dispatch on (e.g.
// "upload-event", "workflow-run-status") and its JSON-serialisable payload.
type Event struct {
	Channel Channel
	Name    string
	Data    any
}

// Publisher sends events to a channel's subscribers.
//
// A failed publish returns an error; callers log it and carry on. A live
// update is never worth failing an upload or a run status update for.
type Publisher interface {
	Publish(ctx context.Context, channel Channel, name string, data any) error
	PublishBatch(ctx context.Context, events []Event) error
}

// envelope is what subscribers receive: the event name travels with the
// payload, since an AppSync channel carries no event names of its own.
type envelope struct {
	Event string `json:"event"`
	Data  any    `json:"data"`
}

// Noop drops every event. Used when realtime is disabled for an environment.
type Noop struct{}

func (Noop) Publish(context.Context, Channel, string, any) error { return nil }
func (Noop) PublishBatch(context.Context, []Event) error         { return nil }
