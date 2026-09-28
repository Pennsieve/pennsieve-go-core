package realtime

import (
	"context"
	"errors"
	"testing"

	"github.com/pennsieve/pennsieve-go-core/pkg/domain"
	"github.com/stretchr/testify/assert"
)

func env(vars map[string]string) Option {
	return withGetenv(func(k string) string { return vars[k] })
}

func TestFromEnvDefaultsToNoop(t *testing.T) {
	for _, v := range []string{"", "noop", "disabled", "NONE", "carrier-pigeon"} {
		t.Run(v, func(t *testing.T) {
			assert.IsType(t, Noop{}, FromEnv(context.Background(), env(map[string]string{EnvProvider: v})))
		})
	}
}

func TestFromEnvAppSync(t *testing.T) {
	// Static credentials from the environment, so no AWS call is made.
	t.Setenv("AWS_REGION", "us-east-1")
	t.Setenv("AWS_ACCESS_KEY_ID", "AKIDTEST")
	t.Setenv("AWS_SECRET_ACCESS_KEY", "secret")
	t.Setenv("AWS_CONFIG_FILE", "/nonexistent")
	t.Setenv("AWS_SHARED_CREDENTIALS_FILE", "/nonexistent")

	p := FromEnv(context.Background(), env(map[string]string{
		EnvProvider:       "appsync",
		EnvEventsEndpoint: "abc.appsync-api.us-east-1.amazonaws.com",
	}))
	a, ok := p.(*AppSync)
	if assert.True(t, ok, "got %T", p) {
		assert.Equal(t, "us-east-1", a.region)
		assert.Equal(t, "https://abc.appsync-api.us-east-1.amazonaws.com/event", a.url)
	}

	assert.IsType(t, Noop{}, FromEnv(context.Background(), env(map[string]string{EnvProvider: "appsync"})),
		"no endpoint: live updates off, service still starts")
}

func TestFromEnvPusher(t *testing.T) {
	vars := env(map[string]string{EnvProvider: "pusher"})

	p := FromEnv(context.Background(), vars, WithPusher(func(context.Context) (domain.PusherAPI, error) {
		return &fakePusher{}, nil
	}))
	assert.IsType(t, &Pusher{}, p)

	assert.IsType(t, Noop{}, FromEnv(context.Background(), vars), "no factory")
	assert.IsType(t, Noop{}, FromEnv(context.Background(), vars, WithPusher(func(context.Context) (domain.PusherAPI, error) {
		return nil, errors.New("bad /ops/pusher-config")
	})), "a broken Pusher config must not crash the service")
	assert.IsType(t, Noop{}, FromEnv(context.Background(), vars, WithPusher(func(context.Context) (domain.PusherAPI, error) {
		return nil, nil
	})), "a nil client is never used")
}

func TestNoopDropsEverything(t *testing.T) {
	var p Publisher = Noop{}
	assert.NoError(t, p.Publish(context.Background(), Dataset(dsUUID), "upload-event", nil))
	assert.NoError(t, p.PublishBatch(context.Background(), []Event{{Channel: Dataset(dsUUID)}}))
}
