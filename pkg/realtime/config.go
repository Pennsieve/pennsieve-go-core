package realtime

import (
	"context"
	"net/http"
	"os"
	"strings"

	"github.com/aws/aws-sdk-go-v2/config"
	log "github.com/sirupsen/logrus"
)

// EnvEventsEndpoint is the Event API's HTTP DNS name. When it is empty, live
// updates are off.
const EnvEventsEndpoint = "REALTIME_EVENTS_ENDPOINT"

type options struct {
	httpClient *http.Client
	getenv     func(string) string
}

// Option configures FromEnv.
type Option func(*options)

// WithHTTPClient sets the HTTP client for the AppSync publisher.
func WithHTTPClient(client *http.Client) Option {
	return func(o *options) { o.httpClient = client }
}

// withGetenv is for tests.
func withGetenv(getenv func(string) string) Option {
	return func(o *options) { o.getenv = getenv }
}

// FromEnv returns an AppSync publisher when REALTIME_EVENTS_ENDPOINT is set,
// and Noop otherwise.
//
// It never fails and never exits the process: a broken configuration returns
// Noop with a warning, so a service always starts and only loses live updates.
func FromEnv(ctx context.Context, opts ...Option) Publisher {
	o := options{getenv: os.Getenv}
	for _, opt := range opts {
		opt(&o)
	}
	endpoint := strings.TrimSpace(o.getenv(EnvEventsEndpoint))
	if endpoint == "" {
		return Noop{}
	}
	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		log.Warnf("realtime: loading AWS config: %v; live updates disabled", err)
		return Noop{}
	}
	p, err := NewAppSync(endpoint, cfg.Region, cfg.Credentials, o.httpClient)
	if err != nil {
		log.Warnf("realtime: %v; live updates disabled", err)
		return Noop{}
	}
	return p
}
