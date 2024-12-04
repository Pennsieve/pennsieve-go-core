package pgdb

import (
	"context"
	"fmt"
	"github.com/pennsieve/pennsieve-go-core/pkg/models/pgdb"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"slices"
	"testing"
)

func TestQueries_GetFeatureFlags(t *testing.T) {
	for scenario, fn := range map[string]func(
		tt *testing.T, store *SQLStore,
	){
		"no feature flags":   testNoFeatureFlags,
		"some feature flags": testSomeFeatureFlags,
	} {
		t.Run(scenario, func(t *testing.T) {
			store := NewSQLStore(testDB[0])
			fn(t, store)
		})
	}
}

func testNoFeatureFlags(t *testing.T, store *SQLStore) {
	orgId := int64(2)
	featureFlags, err := store.GetFeatureFlags(context.Background(), orgId)
	require.NoError(t, err)
	assert.Empty(t, featureFlags)
}

func testSomeFeatureFlags(t *testing.T, store *SQLStore) {
	orgId := int64(402)
	featureFlags, err := store.GetFeatureFlags(context.Background(), orgId)
	require.NoError(t, err)

	assert.Len(t, featureFlags, 5)
	disabledIndex := slices.IndexFunc(featureFlags, func(flag pgdb.FeatureFlags) bool {
		return flag.Feature == "disabled feature" && flag.Enabled == false
	})
	assert.True(t, disabledIndex >= 0, "expected disabled feature not found in %s", featureFlags)
	enabledFeatures := []string{"one", "two", "three", "four"}
	for _, enabledFeature := range enabledFeatures {
		index := slices.IndexFunc(featureFlags, func(flag pgdb.FeatureFlags) bool {
			return flag.Enabled && flag.Feature == fmt.Sprintf("feature %s", enabledFeature)
		})
		assert.True(t, index >= 0, "expected enabled feature %s not found in %s", enabledFeature, featureFlags)
	}
}

func TestQueries_GetEnabledFeatureFlags(t *testing.T) {
	for scenario, fn := range map[string]func(
		tt *testing.T, store *SQLStore,
	){
		"no feature flags":           testNoFeatureFlagsGetEnabled,
		"no enabled feature flags":   testNoEnabledFeatureFlags,
		"some enabled feature flags": testSomeEnabledFeatureFlags,
	} {
		t.Run(scenario, func(t *testing.T) {
			store := NewSQLStore(testDB[0])
			fn(t, store)
		})
	}
}

func testNoFeatureFlagsGetEnabled(t *testing.T, store *SQLStore) {
	orgId := int64(2)
	featureFlags, err := store.GetEnabledFeatureFlags(context.Background(), orgId)
	require.NoError(t, err)
	assert.Empty(t, featureFlags)
}

func testNoEnabledFeatureFlags(t *testing.T, store *SQLStore) {
	orgId := int64(403)
	featureFlags, err := store.GetEnabledFeatureFlags(context.Background(), orgId)
	require.NoError(t, err)
	assert.Empty(t, featureFlags)
}

func testSomeEnabledFeatureFlags(t *testing.T, store *SQLStore) {
	orgId := int64(402)
	featureFlags, err := store.GetEnabledFeatureFlags(context.Background(), orgId)
	require.NoError(t, err)

	assert.Len(t, featureFlags, 4)
	enabledFeatures := []string{"one", "two", "three", "four"}
	for _, enabledFeature := range enabledFeatures {
		index := slices.IndexFunc(featureFlags, func(flag pgdb.FeatureFlags) bool {
			return flag.Enabled && flag.Feature == fmt.Sprintf("feature %s", enabledFeature)
		})
		assert.True(t, index >= 0, "expected enabled feature %s not found in %s", enabledFeature, featureFlags)
	}
}
