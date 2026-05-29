package packagedelete

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/google/uuid"
)

// QueueSender sends one message to the delete queue. Caller supplies
// the implementation for loading messages on the queue.
// DeletePackages calls SendToQueue once per request.
type QueueSender interface {
	SendToQueue(ctx context.Context, body []byte) error
}

// DeleteRequest is a single package to delete.
//
// UserNodeID is the user's node id, e.g. "N:user:abc". (The wire field
// is named "userId" to match pennsieve-api — both sides carry a node id
// here.)
//
// TraceID is any string that helps trace this job in the consumer's
// logs. The upload service uses the manifest id; an API call would use
// its request trace id. Required.
type DeleteRequest struct {
	PackageID      int64  `json:"packageId"`
	OrganizationID int64  `json:"organizationId"`
	UserNodeID     string `json:"userId"`
	TraceID        string `json:"traceId"`
}

// DeletePackages sends one delete message per request through the given
// QueueSender. If a single request fails (bad input, marshal error, send
// error), the others still go through — the first error is returned so
// you know something went wrong without losing the good ones.
func DeletePackages(ctx context.Context, sender QueueSender, reqs []DeleteRequest) error {
	if sender == nil {
		return errors.New("packagedelete: QueueSender is nil")
	}
	if len(reqs) == 0 {
		return nil
	}

	var firstErr error
	recordErr := func(i int, err error) {
		if firstErr == nil {
			firstErr = fmt.Errorf("request %d: %w", i, err)
		}
	}

	for i, r := range reqs {
		if err := validate(r); err != nil {
			recordErr(i, err)
			continue
		}
		body, err := json.Marshal(map[string]any{
			"DeletePackageJob": struct {
				DeleteRequest
				Id string `json:"id"`
			}{r, uuid.NewString()},
		})
		if err != nil {
			recordErr(i, fmt.Errorf("marshal: %w", err))
			continue
		}
		if err := sender.SendToQueue(ctx, body); err != nil {
			recordErr(i, fmt.Errorf("send: %w", err))
			continue
		}
	}
	return firstErr
}

func validate(r DeleteRequest) error {
	if r.PackageID <= 0 {
		return errors.New("PackageID must be > 0")
	}
	if r.OrganizationID <= 0 {
		return errors.New("OrganizationID must be > 0")
	}
	if r.UserNodeID == "" {
		return errors.New("UserNodeID is required")
	}
	if r.TraceID == "" {
		return errors.New("TraceID is required")
	}
	return nil
}
