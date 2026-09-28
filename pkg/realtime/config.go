package realtime

import (
	"context"
	"net/http"
	"os"
	"strings"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/pennsieve/pennsieve-go-core/pkg/domain"
	log "github.com/sirupsen/logrus"
)

// Environment variables that select and configure the publisher.
const (
	EnvProvider       = "REALTIME_PROVIDER"        // appsync | pusher | noop (default noop)
	EnvEventsEndpoint = "REALTIME_EVENTS_ENDPOINT" // Event API HTTP DNS name, for appsync
)

// Provider names accepted in REALTIME_PROVIDER.
const (
	ProviderAppSync = "appsync"
	ProviderPusher  = "pusher"
	ProviderNoop    = "noop"
)

type options struct {
	pusher     func(context.Context) (domain.PusherAPI, error)
	httpClient *http.Client
	getenv     func(string) string
}

// Option configures FromEnv.
type Option func(*options)

// WithPusher supplies the Pusher client for REALTIME_PROVIDER=pusher. It is
// only called when Pusher is selected. Services already build this client from
// /ops/pusher-config; this package doesn't read that parameter itself.
func WithPusher(factory func(context.Context) (domain.PusherAPI, error)) Option {
	return func(o *options) { o.pusher = factory }
}

// WithHTTPClient sets the HTTP client for the AppSync publisher.
func WithHTTPClient(client *http.Client) Option {
	return func(o *options) { o.httpClient = client }
}

// withGetenv is for tests.
func withGetenv(getenv func(string) string) Option {
	return func(o *options) { o.getenv = getenv }
}

// FromEnv returns the publisher REALTIME_PROVIDER selects.
//
// It never fails and never exits the process: a missing or broken
// configuration returns Noop with a warning, so a service always starts and
// only loses live updates. (A nil Pusher client or unparsable config used to
// crash callers at start-up.)
func FromEnv(ctx context.Context, opts ...Option) Publisher {
	o := options{getenv: os.Getenv}
	for _, opt := range opts {
		opt(&o)
	}
	provider := strings.ToLower(strings.TrimSpace(o.getenv(EnvProvider)))
	switch provider {
	case "", ProviderNoop, "disabled", "none":
		return Noop{}

	case ProviderAppSync:
		endpoint := o.getenv(EnvEventsEndpoint)
		if endpoint == "" {
			log.Warnf("realtime: %s=%s but %s is empty; live updates disabled", EnvProvider, provider, EnvEventsEndpoint)
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

	case ProviderPusher:
		if o.pusher == nil {
			log.Warnf("realtime: %s=%s but no Pusher client was supplied; live updates disabled", EnvProvider, provider)
			return Noop{}
		}
		client, err := o.pusher(ctx)
		if err != nil {
			log.Warnf("realtime: building Pusher client: %v; live updates disabled", err)
			return Noop{}
		}
		p, err := NewPusher(client)
		if err != nil {
			log.Warnf("realtime: %v; live updates disabled", err)
			return Noop{}
		}
		return p

	default:
		log.Warnf("realtime: unknown %s=%q; live updates disabled", EnvProvider, provider)
		return Noop{}
	}
}
