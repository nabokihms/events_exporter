// Copyright 2021 The Events Exporter authors
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package vault

import (
	"sort"
	"testing"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	dto "github.com/prometheus/client_model/go"
	"github.com/stretchr/testify/require"
)

func TestCollector(t *testing.T) {
	curTime := time.Now()

	tests := []struct {
		Name    string
		Samples []Sample
		Result  []string
	}{
		{
			Name: "Normal sample",
			Samples: []Sample{
				{
					ID:        "metric-1",
					Labels:    []string{"test-1"},
					Timestamp: curTime,
				},
				{
					ID:        "metric-2",
					Labels:    []string{"test-2"},
					Timestamp: curTime,
				},
			},
			Result: []string{"test-1", "test-2"},
		},
		{
			Name: "Normal sample and Expired sample",
			Samples: []Sample{
				{
					ID:        "metric-1",
					Labels:    []string{"test-ok"},
					Timestamp: curTime,
				},
				{
					ID:        "metric-2",
					Labels:    []string{"test-expired"},
					Timestamp: curTime.Add(-3 * time.Hour),
				},
			},
			Result: []string{"test-ok"},
		},
	}

	for _, tc := range tests {
		t.Run(tc.Name, func(t *testing.T) {
			// To be sure that we have clean registry all the time
			prometheus.DefaultRegisterer = prometheus.NewRegistry()

			vault := NewVault()
			err := vault.RegisterMappings([]Mapping{
				{
					Name:       "test_metric",
					Help:       "Test",
					LabelNames: []string{"name"},
					TTL:        time.Hour,
				},
			})
			require.NoError(t, err)

			for _, s := range tc.Samples {
				err = vault.Store("test_metric", s)
				require.NoError(t, err)
			}

			vault.RemoveStaleMetrics()

			collector := vault.metrics["test_metric"]
			names := make([]string, 0, len(tc.Result))

			namesCh := make(chan prometheus.Metric)
			go func() {
				collector.Collect(namesCh)
				close(namesCh)
			}()

			for metric := range namesCh {
				var convertedMetric dto.Metric
				err := metric.Write(&convertedMetric)
				require.NoError(t, err)

				names = append(names, *convertedMetric.Label[0].Value)
			}

			sort.Strings(names)
			require.Equal(t, tc.Result, names)
		})
	}
}
