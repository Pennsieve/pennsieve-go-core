package domain

import "github.com/pusher/pusher-http-go/v5"

type PusherAPI interface {
	TriggerBatch(batch []pusher.Event) (*pusher.TriggerBatchChannelsList, error)
	Trigger(channel string, eventName string, data interface{}) error
}
