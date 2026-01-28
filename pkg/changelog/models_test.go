package changelog

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestType_MarshalText(t *testing.T) {
	tests := []struct {
		name     string
		t        Type
		expected string
	}{
		{"CreatePackage", CreatePackage, "CREATE_PACKAGE"},
		{"DeletePackage", DeletePackage, "DELETE_PACKAGE"},
		{"RestorePackage", RestorePackage, "RESTORE_PACKAGE"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := tt.t.MarshalText()
			require.NoError(t, err)
			assert.Equal(t, tt.expected, string(result))
		})
	}
}

func TestType_UnmarshalText(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected Type
	}{
		{"CREATE_PACKAGE", "CREATE_PACKAGE", CreatePackage},
		{"DELETE_PACKAGE", "DELETE_PACKAGE", DeletePackage},
		{"RESTORE_PACKAGE", "RESTORE_PACKAGE", RestorePackage},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var result Type
			err := result.UnmarshalText([]byte(tt.input))
			require.NoError(t, err)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestType_UnmarshalText_Unknown(t *testing.T) {
	var result Type
	err := result.UnmarshalText([]byte("INVALID_TYPE"))
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unknown changelog type")
}

func TestMessage_Marshal(t *testing.T) {
	timestamp := time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC)
	msg := Message{
		DatasetChangelogEventJob: MessageParams{
			OrganizationId: 1,
			DatasetId:      2,
			UserId:         "user-123",
			TraceId:        "trace-456",
			Id:             "msg-789",
			Events: []Event{
				{
					EventType: RestorePackage,
					EventDetail: PackageRestoreEvent{
						Id:     100,
						Name:   "restored-package",
						NodeId: "N:package:abc",
					},
					Timestamp: timestamp,
				},
			},
		},
	}

	data, err := json.Marshal(msg)
	require.NoError(t, err)

	assert.Contains(t, string(data), `"eventType":"RESTORE_PACKAGE"`)
	assert.Contains(t, string(data), `"datasetId":2`)
	assert.Contains(t, string(data), `"userId":"user-123"`)
}

func TestMessage_Unmarshal(t *testing.T) {
	jsonStr := `{
		"DatasetChangelogEventJob": {
			"organizationId": 1,
			"datasetId": 2,
			"userId": "user-123",
			"traceId": "trace-456",
			"id": "msg-789",
			"events": [
				{
					"eventType": "RESTORE_PACKAGE",
					"eventDetail": {
						"id": 100,
						"name": "restored-package",
						"nodeId": "N:package:abc"
					},
					"timestamp": "2024-01-15T10:30:00Z"
				}
			]
		}
	}`

	var msg Message
	err := json.Unmarshal([]byte(jsonStr), &msg)
	require.NoError(t, err)

	assert.Equal(t, int64(1), msg.DatasetChangelogEventJob.OrganizationId)
	assert.Equal(t, int64(2), msg.DatasetChangelogEventJob.DatasetId)
	assert.Equal(t, "user-123", msg.DatasetChangelogEventJob.UserId)
	assert.Len(t, msg.DatasetChangelogEventJob.Events, 1)
	assert.Equal(t, RestorePackage, msg.DatasetChangelogEventJob.Events[0].EventType)

	// EventDetail is map[string]interface{} since we didn't implement custom unmarshalling
	detail := msg.DatasetChangelogEventJob.Events[0].EventDetail.(map[string]interface{})
	assert.Equal(t, float64(100), detail["id"]) // JSON numbers unmarshal as float64
	assert.Equal(t, "restored-package", detail["name"])
}

func TestMessage_RoundTrip(t *testing.T) {
	timestamp := time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC)
	original := Message{
		DatasetChangelogEventJob: MessageParams{
			OrganizationId: 1,
			DatasetId:      2,
			UserId:         "user-123",
			TraceId:        "trace-456",
			Id:             "msg-789",
			Events: []Event{
				{
					EventType:   RestorePackage,
					EventDetail: map[string]interface{}{"id": float64(100), "name": "test"},
					Timestamp:   timestamp,
				},
			},
		},
	}

	data, err := json.Marshal(original)
	require.NoError(t, err)

	var unmarshalled Message
	err = json.Unmarshal(data, &unmarshalled)
	require.NoError(t, err)

	assert.Equal(t, original.DatasetChangelogEventJob.OrganizationId, unmarshalled.DatasetChangelogEventJob.OrganizationId)
	assert.Equal(t, original.DatasetChangelogEventJob.DatasetId, unmarshalled.DatasetChangelogEventJob.DatasetId)
	assert.Equal(t, original.DatasetChangelogEventJob.Events[0].EventType, unmarshalled.DatasetChangelogEventJob.Events[0].EventType)
}
