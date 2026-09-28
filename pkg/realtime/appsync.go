package realtime

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	v4 "github.com/aws/aws-sdk-go-v2/aws/signer/v4"
)

// AppSync Events publish limits: 5 events per request (fixed) and 1.2 MB per
// request body.
const (
	maxEventsPerPublish = 5
	maxPublishBytes     = 1_200_000
)

// AppSync publishes events to an AppSync Event API over HTTP, signed with
// SigV4 using the caller's AWS credentials (appsync:EventPublish on the API's
// channel namespaces).
type AppSync struct {
	url         string
	region      string
	credentials aws.CredentialsProvider
	signer      *v4.Signer
	client      *http.Client
	now         func() time.Time
}

// NewAppSync publishes to endpoint, the Event API's HTTP DNS name
// (e.g. abc123.appsync-api.us-east-1.amazonaws.com, with or without https://).
func NewAppSync(endpoint, region string, credentials aws.CredentialsProvider, client *http.Client) (*AppSync, error) {
	endpoint = strings.TrimSuffix(strings.TrimSpace(endpoint), "/")
	if endpoint == "" {
		return nil, errors.New("realtime: AppSync endpoint is required")
	}
	if region == "" {
		return nil, errors.New("realtime: AppSync region is required")
	}
	if credentials == nil {
		return nil, errors.New("realtime: AWS credentials are required")
	}
	if !strings.HasPrefix(endpoint, "https://") && !strings.HasPrefix(endpoint, "http://") {
		endpoint = "https://" + endpoint
	}
	if client == nil {
		client = &http.Client{Timeout: 10 * time.Second}
	}
	return &AppSync{
		url:         endpoint + "/event",
		region:      region,
		credentials: credentials,
		signer:      v4.NewSigner(),
		client:      client,
		now:         time.Now,
	}, nil
}

func (a *AppSync) Publish(ctx context.Context, channel Channel, name string, data any) error {
	return a.PublishBatch(ctx, []Event{{Channel: channel, Name: name, Data: data}})
}

// PublishBatch sends events in order, grouped by channel (AppSync takes one
// channel per request) and split into requests of at most 5 events. It stops
// at the first failed request.
func (a *AppSync) PublishBatch(ctx context.Context, events []Event) error {
	type group struct {
		path   string
		events []string
	}
	var groups []*group
	byPath := map[string]*group{}
	for _, e := range events {
		if err := e.Channel.Validate(); err != nil {
			return err
		}
		if e.Channel.IsWildcard() {
			return fmt.Errorf("realtime: cannot publish to wildcard channel %s", e.Channel.Path())
		}
		body, err := json.Marshal(envelope{Event: e.Name, Data: e.Data})
		if err != nil {
			return fmt.Errorf("realtime: encoding %s event: %w", e.Name, err)
		}
		path := e.Channel.Path()
		g, ok := byPath[path]
		if !ok {
			g = &group{path: path}
			byPath[path] = g
			groups = append(groups, g)
		}
		// AppSync takes each event as a JSON-encoded string.
		g.events = append(g.events, string(body))
	}
	for _, g := range groups {
		for start := 0; start < len(g.events); start += maxEventsPerPublish {
			end := start + maxEventsPerPublish
			if end > len(g.events) {
				end = len(g.events)
			}
			if err := a.send(ctx, g.path, g.events[start:end]); err != nil {
				return err
			}
		}
	}
	return nil
}

type publishRequest struct {
	Channel string   `json:"channel"`
	Events  []string `json:"events"`
}

type publishResponse struct {
	Failed []struct {
		Identifier string `json:"identifier"`
		Index      int    `json:"index"`
		Code       any    `json:"code"`
		Message    string `json:"message"`
	} `json:"failed"`
}

func (a *AppSync) send(ctx context.Context, channel string, events []string) error {
	body, err := json.Marshal(publishRequest{Channel: channel, Events: events})
	if err != nil {
		return err
	}
	if len(body) > maxPublishBytes {
		return fmt.Errorf("realtime: publish to %s is %d bytes, over AppSync's %d-byte limit", channel, len(body), maxPublishBytes)
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, a.url, bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	creds, err := a.credentials.Retrieve(ctx)
	if err != nil {
		return fmt.Errorf("realtime: retrieving AWS credentials: %w", err)
	}
	sum := sha256.Sum256(body)
	if err := a.signer.SignHTTP(ctx, creds, req, hex.EncodeToString(sum[:]), "appsync", a.region, a.now()); err != nil {
		return fmt.Errorf("realtime: signing publish: %w", err)
	}

	resp, err := a.client.Do(req)
	if err != nil {
		return fmt.Errorf("realtime: publish to %s: %w", channel, err)
	}
	defer resp.Body.Close()
	respBody, _ := io.ReadAll(io.LimitReader(resp.Body, 64<<10))
	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		return fmt.Errorf("realtime: publish to %s: HTTP %d: %s", channel, resp.StatusCode, strings.TrimSpace(string(respBody)))
	}
	var parsed publishResponse
	if len(respBody) > 0 && json.Unmarshal(respBody, &parsed) == nil && len(parsed.Failed) > 0 {
		f := parsed.Failed[0]
		return fmt.Errorf("realtime: publish to %s: %d of %d events failed (first: index %d, %v %s)",
			channel, len(parsed.Failed), len(events), f.Index, f.Code, f.Message)
	}
	return nil
}
