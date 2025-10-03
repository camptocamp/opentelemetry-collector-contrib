// Copyright The OpenTelemetry Authors
// SPDX-License-Identifier: Apache-2.0

package logsmatchconnector // import "github.com/open-telemetry/opentelemetry-collector-contrib/connector/logsmatchconnector"

import (
	"context"
	"regexp"
	"time"

	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/consumer"
	"go.opentelemetry.io/collector/pdata/pcommon"
	"go.opentelemetry.io/collector/pdata/plog"
	"go.opentelemetry.io/collector/pdata/pmetric"
	"go.uber.org/zap"
)

type Connector struct {
	component.StartFunc
	component.ShutdownFunc

	logger   *zap.Logger
	match    Match
	metric   Metric
	consumer consumer.Metrics
}

type Match struct {
	attribute   string
	regexp      *regexp.Regexp
	subexpnames []string
}

type Metric struct {
	name          string
	description   string
	keepTimestamp bool
}

func newConnector(logger *zap.Logger, config *Config, nextConsumer consumer.Metrics) (*Connector, error) {
	connector := &Connector{
		match: Match{
			attribute: config.Match.Attribute,
			regexp:    regexp.MustCompile(config.Match.Regexp),
		},
		metric: Metric{
			name:          config.Metric.Name,
			description:   config.Metric.Description,
			keepTimestamp: config.Metric.KeepTimestamp,
		},
		logger:   logger,
		consumer: nextConsumer,
	}

	connector.match.subexpnames = connector.match.regexp.SubexpNames()

	return connector, nil
}

func (_ *Connector) Capabilities() consumer.Capabilities {
	return consumer.Capabilities{MutatesData: false}
}

func (connector *Connector) ConsumeLogs(ctx context.Context, logs plog.Logs) error {
	metrics := pmetric.NewMetrics()

	for _, resourceLogs := range logs.ResourceLogs().All() {
		resourceMetrics := metrics.ResourceMetrics().AppendEmpty()
		resourceLogs.Resource().CopyTo(resourceMetrics.Resource())

		for _, scopeLogs := range resourceLogs.ScopeLogs().All() {
			scopeMetrics := resourceMetrics.ScopeMetrics().AppendEmpty()
			scopeLogs.Scope().CopyTo(scopeMetrics.Scope())

			metric := scopeMetrics.Metrics().AppendEmpty()
			metric.SetName(connector.metric.name)
			metric.SetDescription(connector.metric.description)
			metric.SetUnit("1")

			sum := metric.SetEmptySum()
			sum.SetIsMonotonic(true)
			sum.SetAggregationTemporality(pmetric.AggregationTemporalityDelta)

			for _, logRecord := range scopeLogs.LogRecords().All() {
				var matches []string

				if connector.match.attribute != "" {
					attributeValue, exists := logRecord.Attributes().Get(connector.match.attribute)

					if !exists {
						connector.logger.Warn("Log record does not have the attribute to match, skipping.", zap.String("attribute", connector.match.attribute), zap.Any("logRecord", logRecord))

						continue
					}

					matches = connector.match.regexp.FindStringSubmatch(attributeValue.Str())
				} else {
					matches = connector.match.regexp.FindStringSubmatch(logRecord.Body().Str())
				}

				if matches != nil {
					var timestamp time.Time

					if connector.metric.keepTimestamp {
						timestamp = logRecord.Timestamp().AsTime()

						if timestamp.UnixNano() == 0 {
							timestamp = logRecord.ObservedTimestamp().AsTime()
						}

						if timestamp.UnixNano() == 0 {
							connector.logger.Warn("Log record has no timestamp, skipping.", zap.Any("logRecord", logRecord))
							continue
						}
					} else {
						timestamp = time.Now()
					}

					dataPoint := sum.DataPoints().AppendEmpty()
					dataPoint.SetTimestamp(pcommon.NewTimestampFromTime(timestamp))
					dataPoint.SetIntValue(1)

					for i, subexpNames := 1, connector.match.subexpnames; i < len(subexpNames); i++ {
						dataPoint.Attributes().PutStr(subexpNames[i], matches[i])
					}
				}
			}
		}
	}

	return connector.consumer.ConsumeMetrics(ctx, metrics)
}
