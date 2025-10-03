// Copyright The OpenTelemetry Authors
// SPDX-License-Identifier: Apache-2.0

package logsmatchconnector

import (
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/confmap/confmaptest"
	"go.opentelemetry.io/collector/confmap/xconfmap"

	"github.com/open-telemetry/opentelemetry-collector-contrib/connector/logsmatchconnector/internal/metadata"
)

func TestNewConfig(t *testing.T) {
	config := newConfig()
	require.NotNil(t, config)

	assert.Equal(t, "", config.Match.Attribute)
	assert.Equal(t, defaultMatchRegexp, config.Match.Regexp)
	assert.Equal(t, defaultMetricName, config.Metric.Name)
	assert.Equal(t, defaultMetricDescription, config.Metric.Description)
}

func TestConfigValidate(t *testing.T) {
	defaultConfig := newConfig()

	testCases := []struct {
		name   string
		config *Config
		expect string
	}{
		{
			name:   "default_config",
			config: defaultConfig,
			expect: "",
		},
		{
			name: "custom_config",
			config: &Config{
				Match: MatchConfig{
					Attribute: "message",
					Regexp:    `(?P<log_level>.+): .*`,
				},
				Metric: MetricConfig{
					Name:        "custom_logs_match",
					Description: "Custom description",
				},
			},
			expect: "",
		},
		{
			name: "invalid_regexp",
			config: &Config{
				Match: MatchConfig{
					Regexp: "(?P<log_level>.+: .*",
				},
			},
			expect: "match.regexp is invalid: error parsing regexp: missing closing ): `(?P<log_level>.+: .*`",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := xconfmap.Validate(tc.config)

			if tc.expect == "" {
				assert.NoError(t, err)
			} else {
				assert.ErrorContains(t, err, tc.expect)
			}
		})
	}
}

func TestLoadConfig(t *testing.T) {
	defaultConfig := newConfig()

	testCases := []struct {
		name   string
		expect *Config
	}{
		{
			name:   "default",
			expect: defaultConfig,
		},
		{
			name: "custom_regexp",
			expect: &Config{
				Match: MatchConfig{
					Regexp: "(?P<log_level>.+): .*",
				},
				Metric: defaultConfig.Metric,
			},
		},
		{
			name: "custom_attribute",
			expect: &Config{
				Match: MatchConfig{
					Attribute: "message",
					Regexp:    defaultMatchRegexp,
				},
				Metric: defaultConfig.Metric,
			},
		},
		{
			name: "custom_metric_name",
			expect: &Config{
				Match: defaultConfig.Match,
				Metric: MetricConfig{
					Name:          "custom_logs_match",
					Description:   defaultMetricDescription,
					KeepTimestamp: true,
				},
			},
		},
		{
			name: "custom_metric_description",
			expect: &Config{
				Match: defaultConfig.Match,
				Metric: MetricConfig{
					Name:          defaultMetricName,
					Description:   "Custom description",
					KeepTimestamp: true,
				},
			},
		},
		{
			name: "do_not_keep_timestamp",
			expect: &Config{
				Match: defaultConfig.Match,
				Metric: MetricConfig{
					Name:          defaultMetricName,
					Description:   defaultMetricDescription,
					KeepTimestamp: false,
				},
			},
		},
		{
			name: "custom_all",
			expect: &Config{
				Match: MatchConfig{
					Attribute: "message",
					Regexp:    "(?P<log_level>.+): .*",
				},
				Metric: MetricConfig{
					Name:          "custom_logs_match",
					Description:   "Custom description",
					KeepTimestamp: true,
				},
			},
		},
	}

	configMap, err := confmaptest.LoadConf(filepath.Join("testdata", "config.yaml"))
	require.NoError(t, err)

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			componentID := component.NewIDWithName(metadata.Type, tc.name).String()
			require.True(t, configMap.IsSet(componentID))

			sub, err := configMap.Sub(componentID)
			require.NoError(t, err)

			config := newConfig()
			require.NoError(t, sub.Unmarshal(config))

			assert.NoError(t, config.Validate())
			assert.Equal(t, tc.expect, config)
		})
	}
}
