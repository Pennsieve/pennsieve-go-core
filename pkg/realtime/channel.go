// Package realtime publishes live events (upload rows, run status, app
// builds) to subscribed clients.
//
// Channels are typed so every publisher, the events authorizer and the Go
// subscribers (the agent) agree on the same paths. See
// docs/realtime-appsync-design.md for the channel scheme and who may
// subscribe to what.
package realtime

import (
	"fmt"
	"regexp"
	"strings"
)

// Channel namespaces on the Event API.
const (
	NamespaceDatasets     = "datasets"
	NamespaceRuns         = "runs"
	NamespaceApplications = "applications"
)

// wildcard is the only non-identifier segment, allowed as the last segment of
// a run scope subscription (/runs/<scope>/*).
const wildcard = "*"

// AppSync Events: at most 5 segments per channel (the namespace counts) and
// 50 characters per segment.
const (
	maxSegments      = 5
	maxSegmentLength = 50
)

var segmentPattern = regexp.MustCompile(`^[A-Za-z0-9-]+$`)

// Channel is a namespace plus its path segments, e.g. /datasets/<uuid>.
type Channel struct {
	Namespace string
	Segments  []string
}

// Path is the channel as AppSync Events expects it: /namespace/seg/seg.
func (c Channel) Path() string {
	return "/" + strings.Join(append([]string{c.Namespace}, c.Segments...), "/")
}

func (c Channel) String() string { return c.Path() }

// IsWildcard reports whether the channel subscribes to a whole scope.
func (c Channel) IsWildcard() bool {
	return len(c.Segments) > 0 && c.Segments[len(c.Segments)-1] == wildcard
}

// Validate checks the channel against the namespaces this package defines and
// AppSync's segment limits. Wildcards are valid only as /runs/<scope>/*.
func (c Channel) Validate() error {
	switch c.Namespace {
	case NamespaceDatasets, NamespaceRuns, NamespaceApplications:
	default:
		return fmt.Errorf("realtime: unknown namespace %q", c.Namespace)
	}
	if len(c.Segments) == 0 || len(c.Segments)+1 > maxSegments {
		return fmt.Errorf("realtime: channel %s must have 1-%d segments after the namespace", c.Path(), maxSegments-1)
	}
	for i, s := range c.Segments {
		if s == wildcard {
			if c.Namespace != NamespaceRuns || i != len(c.Segments)-1 || len(c.Segments) != 2 {
				return fmt.Errorf("realtime: wildcard only allowed as /runs/<scope>/*, got %s", c.Path())
			}
			continue
		}
		if len(s) > maxSegmentLength || !segmentPattern.MatchString(s) {
			return fmt.Errorf("realtime: invalid segment %q in %s", s, c.Path())
		}
	}
	return nil
}

// Dataset is the channel for a dataset's events (upload rows):
// /datasets/<datasetUuid>. Accepts a node ID (N:dataset:<uuid>) or a bare
// UUID.
func Dataset(datasetId string) Channel {
	return Channel{Namespace: NamespaceDatasets, Segments: []string{stripNodePrefix(datasetId)}}
}

// Run is the channel for one workflow run's status and processor events:
// /runs/org-<orgUuid>/<runUuid>, or /runs/user-<createdByUuid>/<runUuid> for
// a run with no workspace.
func Run(organizationId, createdBy, runId string) Channel {
	return Channel{Namespace: NamespaceRuns, Segments: []string{runScope(organizationId, createdBy), stripNodePrefix(runId)}}
}

// RunScope subscribes to every run in a workspace (or of a user with no
// workspace): /runs/<scope>/*. For subscribers only; publishing to it fails.
func RunScope(organizationId, createdBy string) Channel {
	return Channel{Namespace: NamespaceRuns, Segments: []string{runScope(organizationId, createdBy), wildcard}}
}

// Application is the channel for an application's build/deploy status:
// /applications/<appUuid>. Apps can be shared across workspaces, so the
// channel carries no workspace; access is decided by app-deploy-service.
func Application(applicationId string) Channel {
	return Channel{Namespace: NamespaceApplications, Segments: []string{stripNodePrefix(applicationId)}}
}

func runScope(organizationId, createdBy string) string {
	if org := stripNodePrefix(organizationId); org != "" {
		return "org-" + org
	}
	return "user-" + stripNodePrefix(createdBy)
}

// stripNodePrefix turns N:<kind>:<uuid> into <uuid>; bare IDs pass through.
func stripNodePrefix(id string) string {
	id = strings.TrimSpace(id)
	if strings.HasPrefix(id, "N:") {
		if i := strings.LastIndex(id, ":"); i >= 0 {
			return id[i+1:]
		}
	}
	return id
}

// pusherChannel maps a channel to the name today's Pusher clients subscribe
// to, so the Pusher adapter keeps dev/prod working during the move.
func pusherChannel(c Channel) (string, error) {
	if err := c.Validate(); err != nil {
		return "", err
	}
	if c.IsWildcard() {
		return "", fmt.Errorf("realtime: cannot publish to wildcard channel %s", c.Path())
	}
	switch c.Namespace {
	case NamespaceDatasets:
		return "dataset-" + c.Segments[0], nil
	case NamespaceApplications:
		return "application-" + c.Segments[0], nil
	case NamespaceRuns:
		scope := c.Segments[0]
		switch {
		case strings.HasPrefix(scope, "org-"):
			return "organization-" + strings.TrimPrefix(scope, "org-") + "-analytics", nil
		case strings.HasPrefix(scope, "user-"):
			return "user-" + strings.TrimPrefix(scope, "user-") + "-analytics", nil
		}
	}
	return "", fmt.Errorf("realtime: no Pusher mapping for %s", c.Path())
}
