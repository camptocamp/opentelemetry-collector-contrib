// Copyright The OpenTelemetry Authors
// SPDX-License-Identifier: Apache-2.0

package tests // import "github.com/open-telemetry/opentelemetry-collector-contrib/connector/logsmatchconnector/internal/tests"

import "path/filepath"

const (
	LogsToMetricsDataKind = "logstometrics"
)

type _DataDir string

func DataDir(kind string) _DataDir {
	return _DataDir(filepath.Join("testdata", kind))
}

func (dataDir _DataDir) Path() string {
	return string(dataDir)
}

func (dataDir _DataDir) InputFilePath() string {
	return filepath.Join(string(dataDir), "input.yaml")
}

func (dataDir _DataDir) OutputFilePath(name string) string {
	return filepath.Join(string(dataDir), name+".yaml")
}
