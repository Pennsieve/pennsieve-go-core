package realtime

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

const (
	orgUUID  = "11111111-2222-3333-4444-555555555555"
	userUUID = "aaaaaaaa-bbbb-cccc-dddd-eeeeeeeeeeee"
	runUUID  = "99999999-8888-7777-6666-555555555555"
	dsUUID   = "12345678-1234-1234-1234-123456789abc"
)

func TestChannelPaths(t *testing.T) {
	cases := []struct {
		name string
		ch   Channel
		want string
	}{
		{"dataset node id", Dataset("N:dataset:" + dsUUID), "/datasets/" + dsUUID},
		{"dataset bare uuid", Dataset(dsUUID), "/datasets/" + dsUUID},
		{"run in a workspace", Run("N:organization:"+orgUUID, "N:user:"+userUUID, runUUID), "/runs/org-" + orgUUID + "/" + runUUID},
		{"run without a workspace", Run("", "N:user:"+userUUID, runUUID), "/runs/user-" + userUUID + "/" + runUUID},
		{"workspace run scope", RunScope("N:organization:"+orgUUID, ""), "/runs/org-" + orgUUID + "/*"},
		{"user run scope", RunScope("", userUUID), "/runs/user-" + userUUID + "/*"},
		{"application", Application("app-" + dsUUID[:8]), "/applications/app-12345678"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			assert.Equal(t, c.want, c.ch.Path())
			assert.NoError(t, c.ch.Validate())
		})
	}
}

func TestValidateRejects(t *testing.T) {
	cases := map[string]Channel{
		"unknown namespace":         {Namespace: "orgs", Segments: []string{orgUUID}},
		"no segments":               {Namespace: NamespaceDatasets},
		"too many segments":         {Namespace: NamespaceRuns, Segments: []string{"a", "b", "c", "d", "e"}},
		"segment too long":          {Namespace: NamespaceDatasets, Segments: []string{strings.Repeat("a", 51)}},
		"illegal characters":        {Namespace: NamespaceDatasets, Segments: []string{"N:dataset:x"}},
		"wildcard outside runs":     {Namespace: NamespaceDatasets, Segments: []string{"*"}},
		"wildcard not at run scope": {Namespace: NamespaceRuns, Segments: []string{"*"}},
		"wildcard mid-path":         {Namespace: NamespaceRuns, Segments: []string{"org-x", "*", "y"}},
		"empty dataset id":          Dataset(""),
	}
	for name, ch := range cases {
		t.Run(name, func(t *testing.T) {
			assert.Error(t, ch.Validate(), ch.Path())
		})
	}
}

func TestSegmentFitsAppSyncLimit(t *testing.T) {
	// "org-" + a UUID is the longest segment the scheme produces.
	assert.LessOrEqual(t, len(Run(orgUUID, "", runUUID).Segments[0]), maxSegmentLength)
}
