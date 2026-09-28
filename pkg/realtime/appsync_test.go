package realtime

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type captured struct {
	auth    string
	date    string
	channel string
	events  []envelope
}

func newTestServer(t *testing.T, status int, response string) (*httptest.Server, *[]captured) {
	var mu sync.Mutex
	var got []captured
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPost, r.Method)
		assert.Equal(t, "/event", r.URL.Path)
		assert.Equal(t, "application/json", r.Header.Get("Content-Type"))
		body, _ := io.ReadAll(r.Body)
		var req publishRequest
		require.NoError(t, json.Unmarshal(body, &req))
		c := captured{auth: r.Header.Get("Authorization"), date: r.Header.Get("X-Amz-Date"), channel: req.Channel}
		for _, e := range req.Events {
			var env envelope
			require.NoError(t, json.Unmarshal([]byte(e), &env), "each event is a JSON-encoded string")
			c.events = append(c.events, env)
		}
		mu.Lock()
		got = append(got, c)
		mu.Unlock()
		w.WriteHeader(status)
		_, _ = w.Write([]byte(response))
	}))
	t.Cleanup(srv.Close)
	return srv, &got
}

func newTestPublisher(t *testing.T, url string) *AppSync {
	p, err := NewAppSync(url, "us-east-1", credentials.NewStaticCredentialsProvider("AKIDTEST", "secret", ""), nil)
	require.NoError(t, err)
	p.now = func() time.Time { return time.Date(2026, 9, 28, 12, 0, 0, 0, time.UTC) }
	return p
}

func TestPublishSignsAndWrapsTheEvent(t *testing.T) {
	srv, got := newTestServer(t, http.StatusOK, `{"successful":[{"identifier":"x","index":0}],"failed":[]}`)
	p := newTestPublisher(t, srv.URL)

	rows := []map[string]string{{"name": "a.csv", "node_id": "N:package:1"}}
	require.NoError(t, p.Publish(context.Background(), Dataset(dsUUID), "upload-event", rows))

	require.Len(t, *got, 1)
	c := (*got)[0]
	assert.Equal(t, "/datasets/"+dsUUID, c.channel)
	assert.True(t, strings.HasPrefix(c.auth, "AWS4-HMAC-SHA256 Credential=AKIDTEST/20260928/us-east-1/appsync/aws4_request"), c.auth)
	assert.Equal(t, "20260928T120000Z", c.date)
	require.Len(t, c.events, 1)
	assert.Equal(t, "upload-event", c.events[0].Event)
	data, _ := json.Marshal(c.events[0].Data)
	assert.JSONEq(t, `[{"name":"a.csv","node_id":"N:package:1"}]`, string(data))
}

func TestPublishBatchSplitsByChannelAndLimit(t *testing.T) {
	srv, got := newTestServer(t, http.StatusOK, `{"failed":[]}`)
	p := newTestPublisher(t, srv.URL)

	run := Run(orgUUID, userUUID, runUUID)
	var events []Event
	for i := 0; i < 6; i++ {
		events = append(events, Event{Channel: run, Name: "workflow-processor-status", Data: map[string]int{"i": i}})
	}
	events = append(events, Event{Channel: Dataset(dsUUID), Name: "upload-event", Data: []string{}})
	require.NoError(t, p.PublishBatch(context.Background(), events))

	require.Len(t, *got, 3, "6 run events → 5 + 1, plus one dataset request")
	assert.Equal(t, run.Path(), (*got)[0].channel)
	assert.Len(t, (*got)[0].events, 5)
	assert.Len(t, (*got)[1].events, 1)
	assert.Equal(t, "/datasets/"+dsUUID, (*got)[2].channel)
}

func TestPublishRejectsWildcardsAndInvalidChannels(t *testing.T) {
	srv, got := newTestServer(t, http.StatusOK, `{}`)
	p := newTestPublisher(t, srv.URL)
	assert.Error(t, p.Publish(context.Background(), RunScope(orgUUID, ""), "workflow-run-status", nil))
	assert.Error(t, p.Publish(context.Background(), Channel{Namespace: "orgs", Segments: []string{orgUUID}}, "x", nil))
	assert.Empty(t, *got, "nothing is sent for an invalid channel")
}

func TestPublishReportsFailures(t *testing.T) {
	t.Run("http error", func(t *testing.T) {
		srv, _ := newTestServer(t, http.StatusUnauthorized, `{"errors":[{"errorType":"UnauthorizedException"}]}`)
		err := newTestPublisher(t, srv.URL).Publish(context.Background(), Dataset(dsUUID), "upload-event", nil)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "HTTP 401")
	})
	t.Run("failed events in a 200", func(t *testing.T) {
		srv, _ := newTestServer(t, http.StatusOK, `{"successful":[],"failed":[{"identifier":"x","index":0,"code":400,"message":"bad"}]}`)
		err := newTestPublisher(t, srv.URL).Publish(context.Background(), Dataset(dsUUID), "upload-event", nil)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "1 of 1 events failed")
	})
}

func TestPublishRejectsOversizedRequests(t *testing.T) {
	srv, got := newTestServer(t, http.StatusOK, `{}`)
	big := strings.Repeat("x", maxPublishBytes)
	err := newTestPublisher(t, srv.URL).Publish(context.Background(), Dataset(dsUUID), "upload-event", big)
	require.Error(t, err)
	assert.Empty(t, *got)
}

func TestNewAppSyncNormalisesEndpoint(t *testing.T) {
	creds := credentials.NewStaticCredentialsProvider("a", "b", "")
	p, err := NewAppSync("abc.appsync-api.us-east-1.amazonaws.com/", "us-east-1", creds, nil)
	require.NoError(t, err)
	assert.Equal(t, "https://abc.appsync-api.us-east-1.amazonaws.com/event", p.url)

	_, err = NewAppSync("", "us-east-1", creds, nil)
	assert.Error(t, err)
	_, err = NewAppSync("x", "", creds, nil)
	assert.Error(t, err)
	_, err = NewAppSync("x", "us-east-1", nil, nil)
	assert.Error(t, err)
}
