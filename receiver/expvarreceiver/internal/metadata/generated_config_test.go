// Code generated by mdatagen. DO NOT EDIT.

package metadata

import (
	"path/filepath"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/stretchr/testify/require"

	"go.opentelemetry.io/collector/confmap"
	"go.opentelemetry.io/collector/confmap/confmaptest"
)

func TestMetricsBuilderConfig(t *testing.T) {
	tests := []struct {
		name string
		want MetricsBuilderConfig
	}{
		{
			name: "default",
			want: DefaultMetricsBuilderConfig(),
		},
		{
			name: "all_set",
			want: MetricsBuilderConfig{
				Metrics: MetricsConfig{
					ProcessRuntimeMemstatsBuckHashSys:   MetricConfig{Enabled: true},
					ProcessRuntimeMemstatsFrees:         MetricConfig{Enabled: true},
					ProcessRuntimeMemstatsGcCPUFraction: MetricConfig{Enabled: true},
					ProcessRuntimeMemstatsGcSys:         MetricConfig{Enabled: true},
					ProcessRuntimeMemstatsHeapAlloc:     MetricConfig{Enabled: true},
					ProcessRuntimeMemstatsHeapIdle:      MetricConfig{Enabled: true},
					ProcessRuntimeMemstatsHeapInuse:     MetricConfig{Enabled: true},
					ProcessRuntimeMemstatsHeapObjects:   MetricConfig{Enabled: true},
					ProcessRuntimeMemstatsHeapReleased:  MetricConfig{Enabled: true},
					ProcessRuntimeMemstatsHeapSys:       MetricConfig{Enabled: true},
					ProcessRuntimeMemstatsLastPause:     MetricConfig{Enabled: true},
					ProcessRuntimeMemstatsLookups:       MetricConfig{Enabled: true},
					ProcessRuntimeMemstatsMallocs:       MetricConfig{Enabled: true},
					ProcessRuntimeMemstatsMcacheInuse:   MetricConfig{Enabled: true},
					ProcessRuntimeMemstatsMcacheSys:     MetricConfig{Enabled: true},
					ProcessRuntimeMemstatsMspanInuse:    MetricConfig{Enabled: true},
					ProcessRuntimeMemstatsMspanSys:      MetricConfig{Enabled: true},
					ProcessRuntimeMemstatsNextGc:        MetricConfig{Enabled: true},
					ProcessRuntimeMemstatsNumForcedGc:   MetricConfig{Enabled: true},
					ProcessRuntimeMemstatsNumGc:         MetricConfig{Enabled: true},
					ProcessRuntimeMemstatsOtherSys:      MetricConfig{Enabled: true},
					ProcessRuntimeMemstatsPauseTotal:    MetricConfig{Enabled: true},
					ProcessRuntimeMemstatsStackInuse:    MetricConfig{Enabled: true},
					ProcessRuntimeMemstatsStackSys:      MetricConfig{Enabled: true},
					ProcessRuntimeMemstatsSys:           MetricConfig{Enabled: true},
					ProcessRuntimeMemstatsTotalAlloc:    MetricConfig{Enabled: true},
				},
			},
		},
		{
			name: "none_set",
			want: MetricsBuilderConfig{
				Metrics: MetricsConfig{
					ProcessRuntimeMemstatsBuckHashSys:   MetricConfig{Enabled: false},
					ProcessRuntimeMemstatsFrees:         MetricConfig{Enabled: false},
					ProcessRuntimeMemstatsGcCPUFraction: MetricConfig{Enabled: false},
					ProcessRuntimeMemstatsGcSys:         MetricConfig{Enabled: false},
					ProcessRuntimeMemstatsHeapAlloc:     MetricConfig{Enabled: false},
					ProcessRuntimeMemstatsHeapIdle:      MetricConfig{Enabled: false},
					ProcessRuntimeMemstatsHeapInuse:     MetricConfig{Enabled: false},
					ProcessRuntimeMemstatsHeapObjects:   MetricConfig{Enabled: false},
					ProcessRuntimeMemstatsHeapReleased:  MetricConfig{Enabled: false},
					ProcessRuntimeMemstatsHeapSys:       MetricConfig{Enabled: false},
					ProcessRuntimeMemstatsLastPause:     MetricConfig{Enabled: false},
					ProcessRuntimeMemstatsLookups:       MetricConfig{Enabled: false},
					ProcessRuntimeMemstatsMallocs:       MetricConfig{Enabled: false},
					ProcessRuntimeMemstatsMcacheInuse:   MetricConfig{Enabled: false},
					ProcessRuntimeMemstatsMcacheSys:     MetricConfig{Enabled: false},
					ProcessRuntimeMemstatsMspanInuse:    MetricConfig{Enabled: false},
					ProcessRuntimeMemstatsMspanSys:      MetricConfig{Enabled: false},
					ProcessRuntimeMemstatsNextGc:        MetricConfig{Enabled: false},
					ProcessRuntimeMemstatsNumForcedGc:   MetricConfig{Enabled: false},
					ProcessRuntimeMemstatsNumGc:         MetricConfig{Enabled: false},
					ProcessRuntimeMemstatsOtherSys:      MetricConfig{Enabled: false},
					ProcessRuntimeMemstatsPauseTotal:    MetricConfig{Enabled: false},
					ProcessRuntimeMemstatsStackInuse:    MetricConfig{Enabled: false},
					ProcessRuntimeMemstatsStackSys:      MetricConfig{Enabled: false},
					ProcessRuntimeMemstatsSys:           MetricConfig{Enabled: false},
					ProcessRuntimeMemstatsTotalAlloc:    MetricConfig{Enabled: false},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := loadMetricsBuilderConfig(t, tt.name)
			diff := cmp.Diff(tt.want, cfg, cmpopts.IgnoreUnexported(MetricConfig{}))
			require.Emptyf(t, diff, "Config mismatch (-expected +actual):\n%s", diff)
		})
	}
}

func loadMetricsBuilderConfig(t *testing.T, name string) MetricsBuilderConfig {
	cm, err := confmaptest.LoadConf(filepath.Join("testdata", "config.yaml"))
	require.NoError(t, err)
	sub, err := cm.Sub(name)
	require.NoError(t, err)
	cfg := DefaultMetricsBuilderConfig()
	require.NoError(t, sub.Unmarshal(&cfg, confmap.WithIgnoreUnused()))
	return cfg
}
