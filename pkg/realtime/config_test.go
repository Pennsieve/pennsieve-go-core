package realtime

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func env(vars map[string]string) Option {
	return withGetenv(func(k string) string { return vars[k] })
}

func TestFromEnvWithoutEndpointIsNoop(t *testing.T) {
	for _, v := range []string{"", "  "} {
		assert.IsType(t, Noop{}, FromEnv(context.Background(), env(map[string]string{EnvEventsEndpoint: v})))
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
		EnvEventsEndpoint: "abc.appsync-api.us-east-1.amazonaws.com",
	}))
	a, ok := p.(*AppSync)
	if assert.True(t, ok, "got %T", p) {
		assert.Equal(t, "us-east-1", a.region)
		assert.Equal(t, "https://abc.appsync-api.us-east-1.amazonaws.com/event", a.url)
	}
}

func TestFromEnvWithoutRegionIsNoop(t *testing.T) {
	t.Setenv("AWS_REGION", "")
	t.Setenv("AWS_DEFAULT_REGION", "")
	t.Setenv("AWS_ACCESS_KEY_ID", "AKIDTEST")
	t.Setenv("AWS_SECRET_ACCESS_KEY", "secret")
	t.Setenv("AWS_CONFIG_FILE", "/nonexistent")
	t.Setenv("AWS_SHARED_CREDENTIALS_FILE", "/nonexistent")

	p := FromEnv(context.Background(), env(map[string]string{
		EnvEventsEndpoint: "abc.appsync-api.us-east-1.amazonaws.com",
	}))
	assert.IsType(t, Noop{}, p, "a broken configuration must not crash the service")
}

func TestNoopDropsEverything(t *testing.T) {
	var p Publisher = Noop{}
	assert.NoError(t, p.Publish(context.Background(), Dataset(dsUUID), "upload-event", nil))
	assert.NoError(t, p.PublishBatch(context.Background(), []Event{{Channel: Dataset(dsUUID)}}))
}
