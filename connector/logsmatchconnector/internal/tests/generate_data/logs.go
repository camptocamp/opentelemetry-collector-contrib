// Copyright The OpenTelemetry Authors
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"fmt"
	"os"
	"time"

	"go.opentelemetry.io/collector/pdata/pcommon"
	"go.opentelemetry.io/collector/pdata/plog"

	"github.com/open-telemetry/opentelemetry-collector-contrib/connector/logsmatchconnector/internal/tests"
	"github.com/open-telemetry/opentelemetry-collector-contrib/pkg/golden"
)

type LogsGenerator struct {
	time time.Time
}

func newGenerator() *LogsGenerator {
	return &LogsGenerator{
		time: time.Unix(0, 0),
	}
}

func (generator *LogsGenerator) generateLogs() error {
	dataDir := tests.DataDir(tests.LogsToMetricsDataKind)

	err := os.MkdirAll(dataDir.Path(), 0o755)
	if err != nil {
		return fmt.Errorf("failed to create data directory: %s", err)
	}

	logs := plog.NewLogs()

	generator.generateResourceLogs(logs)

	err = golden.WriteLogsToFile(dataDir.InputFilePath(), logs)
	if err != nil {
		return fmt.Errorf("failed to write logs to file: %s", err)
	}

	return nil
}

func (generator *LogsGenerator) generateResourceLogs(logs plog.Logs) {
	resources := []struct {
		environment string
		service     string
		version     string
		component   string
	}{
		{
			environment: "production",
			service:     "frontend",
			version:     "v1",
		},
		{
			environment: "production",
			service:     "backend",
			component:   "api",
		},
		{
			environment: "development",
			service:     "backend",
			component:   "api",
		},
	}

	for _, resource := range resources {
		generator.time = generator.time.Add(10000000 * time.Nanosecond).Round(10000000 * time.Nanosecond)

		resourceLogs := logs.ResourceLogs().AppendEmpty()
		resourceLogs.Resource().Attributes().PutStr("environment", resource.environment)
		resourceLogs.Resource().Attributes().PutStr("service", resource.service)

		if resource.version != "" {
			resourceLogs.Resource().Attributes().PutStr("version", resource.version)
		}

		if resource.component != "" {
			resourceLogs.Resource().Attributes().PutStr("component", resource.component)
		}

		generator.generateScopeLogs(resourceLogs)
	}
}

func (generator *LogsGenerator) generateScopeLogs(resourceLogs plog.ResourceLogs) {
	scopes := []struct {
		name       string
		version    string
		attributes map[string]string
	}{
		{
			name:    "sdk",
			version: "v1",
			attributes: map[string]string{
				"url":    "example.com/sdk",
				"source": "code.example.com/sdk",
			},
		},
		{
			name:    "lib",
			version: "v2",
			attributes: map[string]string{
				"source": "code.example.com/lib",
				"commit": "01ba4719c80b6fe911b091a7c05124b64eeece964e09c058ef8f9805daca546b",
			},
		},
	}

	for _, scope := range scopes {
		generator.time = generator.time.Add(1000000 * time.Nanosecond).Round(1000000 * time.Nanosecond)

		scopeLogs := resourceLogs.ScopeLogs().AppendEmpty()
		scopeLogs.Scope().SetName(scope.name)
		scopeLogs.Scope().SetVersion(scope.version)

		for k, v := range scope.attributes {
			scopeLogs.Scope().Attributes().PutStr(k, v)
		}

		generator.generateLogRecords(scopeLogs)
	}
}

func (generator *LogsGenerator) generateLogRecords(scopeLogs plog.ScopeLogs) {
	modules := []string{
		"",
		"auth",
		"payment",
	}

	for _, module := range modules {
		generator.time = generator.time.Add(100000 * time.Nanosecond).Round(100000 * time.Nanosecond)

		for _, isPanic := range []bool{false, true} {
			generator.time = generator.time.Add(10000 * time.Nanosecond).Round(10000 * time.Nanosecond)

			for _, withTimestamp := range []bool{false, true} {
				generator.time = generator.time.Add(1000 * time.Nanosecond).Round(1000 * time.Nanosecond)

				for _, withObservedTimestamp := range []bool{false, true} {
					generator.time = generator.time.Add(100 * time.Nanosecond).Round(100 * time.Nanosecond)

					for _, useAttribute := range []bool{false, true} {
						generator.time = generator.time.Add(10 * time.Nanosecond).Round(10 * time.Nanosecond)

						var message string

						if module != "" {
							message = fmt.Sprintf("[%s] ", module)
						}

						if isPanic {
							message += "panic: unrecoverable error"
						} else {
							message += "INFO: operation completed successfully"
						}

						logRecord := scopeLogs.LogRecords().AppendEmpty()

						if withTimestamp {
							logRecord.SetTimestamp(pcommon.NewTimestampFromTime(generator.time))
						}

						if withObservedTimestamp {
							logRecord.SetObservedTimestamp(pcommon.NewTimestampFromTime(generator.time.Add(time.Nanosecond)))
						}

						if useAttribute {
							logRecord.Attributes().PutStr("message", message)
						} else {
							logRecord.Body().SetStr(message)
						}
					}
				}
			}
		}
	}
}
