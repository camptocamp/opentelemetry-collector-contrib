// Copyright The OpenTelemetry Authors
// SPDX-License-Identifier: Apache-2.0

package logsmatchconnector // import "github.com/open-telemetry/opentelemetry-collector-contrib/connector/logsmatchconnector"

import (
	"context"

	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/connector"
	"go.opentelemetry.io/collector/consumer"

	"github.com/open-telemetry/opentelemetry-collector-contrib/connector/logsmatchconnector/internal/metadata"
)

func NewFactory() connector.Factory {
	return connector.NewFactory(
		metadata.Type,
		createDefaultConfig,
		connector.WithLogsToMetrics(createLogsToMetricsConnector, metadata.LogsToMetricsStability))
}

func createDefaultConfig() component.Config {
	return &Config{
		Match: MatchConfig{
			Regexp: defaultMatchRegexp,
		},
		Metric: MetricConfig{
			Name:        defaultMetricName,
			Description: defaultMetricDescription,
		},
	}
}

func createLogsToMetricsConnector(ctx context.Context, settings connector.Settings, config component.Config, nextConsumer consumer.Metrics) (connector.Logs, error) {
	return newConnector(settings.Logger, config.(*Config), nextConsumer)
}
