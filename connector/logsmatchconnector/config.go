// Copyright The OpenTelemetry Authors
// SPDX-License-Identifier: Apache-2.0

package logsmatchconnector // import "github.com/open-telemetry/opentelemetry-collector-contrib/connector/logsmatchconnector"

import (
	"fmt"
	"regexp"
)

const (
	defaultMatchRegexp       = ".*"
	defaultMetricName        = "log.record.match.count"
	defaultMetricDescription = "The number of logs matched by the regexp"
)

type Config struct {
	Match  MatchConfig  `mapstructure:"match"`
	Metric MetricConfig `mapstructure:"metric"`
}

type MatchConfig struct {
	Attribute string `mapstructure:"attribute"`
	Regexp    string `mapstructure:"regexp"`
}

type MetricConfig struct {
	Name          string `mapstructure:"name"`
	Description   string `mapstructure:"description"`
	KeepTimestamp bool   `mapstructure:"keep_timestamp"`
}

func newConfig() *Config {
	return &Config{
		Match: MatchConfig{
			Regexp: defaultMatchRegexp,
		},
		Metric: MetricConfig{
			Name:          defaultMetricName,
			Description:   defaultMetricDescription,
			KeepTimestamp: true,
		},
	}
}

func (config *Config) Validate() error {
	if _, err := regexp.Compile(config.Match.Regexp); err != nil {
		return fmt.Errorf("match.regexp is invalid: %s", err)
	}

	return nil
}
