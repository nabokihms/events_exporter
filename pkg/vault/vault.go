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
	"fmt"
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

type MetricsVault struct {
	now func() time.Time

	metrics []ConstMetricCollector
}

type Mapping struct {
	Name string `yaml:"name"`
	Help string `yaml:"help,omitempty"`

	LabelNames []string      `yaml:"labels,omitempty"`
	TTL        time.Duration `yaml:"ttl,omitempty"`
}

type Sample struct {
	Labels []string
	Value  float64

	Timestamp time.Time
}

func NewVault() *MetricsVault {
	return &MetricsVault{now: time.Now}
}

func (v *MetricsVault) RegisterMappings(mappings []Mapping) error {
	for _, mapping := range mappings {
		collector := NewConstGaugeCollector(mapping)
		v.metrics = append(v.metrics, collector)

		if err := prometheus.Register(collector); err != nil {
			return fmt.Errorf("mapping registration: %v", err)
		}
	}
	return nil
}

func (v *MetricsVault) Store(index int, sample Sample) error {
	binding := v.metrics[index]
	binding.Store(v.now(), sample)
	return nil
}

func (v *MetricsVault) RemoveStaleMetrics() {
	currentTime := v.now()

	for _, m := range v.metrics {
		m.Clear(currentTime)
	}
}
