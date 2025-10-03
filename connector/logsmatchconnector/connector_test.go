// Copyright The OpenTelemetry Authors
// SPDX-License-Identifier: Apache-2.0

package logsmatchconnector

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/collector/component/componenttest"
	"go.opentelemetry.io/collector/confmap"
	"go.opentelemetry.io/collector/connector/connectortest"
	"go.opentelemetry.io/collector/consumer/consumertest"
	"go.uber.org/zap"

	"github.com/open-telemetry/opentelemetry-collector-contrib/connector/logsmatchconnector/internal/metadata"
	"github.com/open-telemetry/opentelemetry-collector-contrib/connector/logsmatchconnector/internal/tests"
	"github.com/open-telemetry/opentelemetry-collector-contrib/pkg/golden"
	"github.com/open-telemetry/opentelemetry-collector-contrib/pkg/pdatatest/pmetrictest"
)

func TestConnectorCapabilities(t *testing.T) {
	connector, err := newConnector(zap.NewNop(), &Config{}, nil)
	require.NoError(t, err)

	capabilities := connector.Capabilities()
	assert.False(t, capabilities.MutatesData)
}

func TestLogsToMetrics(t *testing.T) {
	const (
		MatchAttribute                = "message"
		MatchRegexp                   = `panic: .*`
		MatchRegexpWithCapturingGroup = `(?:\[(?P<module>[a-z]+)\] )?panic: .*`
		MetricName                    = "fatal_errors"
		MetricDescription             = "Counts of fatal errors by compoment"
	)

	testCases := []struct {
		name   string
		config map[string]any
	}{
		{
			name: "match_body",
			config: map[string]any{
				"match": map[string]any{
					"regexp": MatchRegexp,
				},
				"metric": map[string]any{
					"name":        MetricName,
					"description": MetricDescription,
				},
			},
		},
		{
			name: "match_body_do_not_keep_timestamp",
			config: map[string]any{
				"match": map[string]any{
					"regexp": MatchRegexp,
				},
				"metric": map[string]any{
					"name":           MetricName,
					"description":    MetricDescription,
					"keep_timestamp": false,
				},
			},
		},
		{
			name: "match_attribute",
			config: map[string]any{
				"match": map[string]any{
					"attribute": MatchAttribute,
					"regexp":    MatchRegexp,
				},
				"metric": map[string]any{
					"name":        MetricName,
					"description": MetricDescription,
				},
			},
		},
		{
			name: "match_attribute_with_regex_capturing_group",
			config: map[string]any{
				"match": map[string]any{
					"attribute": MatchAttribute,
					"regexp":    MatchRegexpWithCapturingGroup,
				},
				"metric": map[string]any{
					"name":        MetricName,
					"description": MetricDescription,
				},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			dataDir := tests.DataDir(tests.LogsToMetricsDataKind)

			config := newConfig()

			require.NoError(t, confmap.NewFromStringMap(tc.config).Unmarshal(&config))

			require.NoError(t, config.Validate())

			host := componenttest.NewNopHost()
			settings := connectortest.NewNopSettings(metadata.Type)

			sink := &consumertest.MetricsSink{}

			connector, err := newConnector(settings.Logger, config, sink)
			require.NoError(t, err)

			require.NoError(t, connector.Start(t.Context(), host))

			defer func() {
				require.NoError(t, connector.Shutdown(t.Context()))
			}()

			logs, err := golden.ReadLogs(dataDir.InputFilePath())
			require.NoError(t, err)

			require.NoError(t, connector.ConsumeLogs(t.Context(), logs))

			allMetrics := sink.AllMetrics()
			require.Len(t, allMetrics, 1)

			metrics := allMetrics[0]

			// require.NoError(t, golden.WriteMetrics(t, dataDir.OutputFilePath(tc.name), metrics, golden.SkipMetricTimestampNormalization()))

			expectedMetrics, err := golden.ReadMetrics(dataDir.OutputFilePath(tc.name))
			require.NoError(t, err)

			compareMetricsOptions := []pmetrictest.CompareMetricsOption{
				pmetrictest.IgnoreResourceMetricsOrder(),
				pmetrictest.IgnoreScopeMetricsOrder(),
			}

			if !config.Metric.KeepTimestamp {
				compareMetricsOptions = append(compareMetricsOptions, pmetrictest.IgnoreTimestamp())
			}

			assert.NoError(t, pmetrictest.CompareMetrics(expectedMetrics, metrics, compareMetricsOptions...))
		})
	}
}
