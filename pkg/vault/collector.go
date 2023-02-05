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
	"bytes"
	"hash/fnv"
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/common/log"
)

const labelsSeparator = byte(255)

type ConstMetricCollector interface {
	Describe(chan<- *prometheus.Desc)
	Collect(chan<- prometheus.Metric)
	Store(time.Time, Sample)
	Clear(time.Time)
}

var (
	_ ConstMetricCollector = (*GaugeCollector)(nil)
)

type StampedGaugeMetric struct {
	Value float64

	LabelValues []string
	LastUpdate  time.Time
}

type GaugeCollector struct {
	mu sync.RWMutex

	collection map[uint64]StampedGaugeMetric
	desc       *prometheus.Desc
	mapping    Mapping
}

func NewConstGaugeCollector(mapping Mapping) *GaugeCollector {
	desc := prometheus.NewDesc(mapping.Name, mapping.Help, mapping.LabelNames, nil)
	return &GaugeCollector{mapping: mapping, collection: make(map[uint64]StampedGaugeMetric), desc: desc}
}

func (c *GaugeCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- c.desc
}

func (c *GaugeCollector) Collect(ch chan<- prometheus.Metric) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	for _, s := range c.collection {
		metric, err := prometheus.NewConstMetric(c.desc, prometheus.GaugeValue, s.Value, s.LabelValues...)
		if err != nil {
			// TODO(nabokihms): add counter for errors
			log.Warnf("prepare gauge: %v", err)
			continue
		}
		ch <- metric
	}
}

func (c *GaugeCollector) Store(timestamp time.Time, sample Sample) {
	labelsHash := hashLabels(sample.Labels)

	c.mu.Lock()
	defer c.mu.Unlock()

	storedMetric, ok := c.collection[labelsHash]
	if !ok {
		storedMetric = StampedGaugeMetric{LabelValues: sample.Labels, LastUpdate: timestamp}
	}

	// If sample contains last update information, that means it was collected from the metric source.
	// Consider this as a truth.
	if !sample.Timestamp.IsZero() {
		storedMetric.LastUpdate = sample.Timestamp
		// If a value has not been changed, this is a fake update.
		// The same event cannot be truly updated without changing its value.
		// For this kind of events timestamp should not be updated by exporter.
	} else if sample.Value != storedMetric.Value {
		storedMetric.LastUpdate = timestamp
	}

	storedMetric.Value = sample.Value
	c.collection[labelsHash] = storedMetric
}

func (c *GaugeCollector) Clear(now time.Time) {
	c.mu.Lock()
	defer c.mu.Unlock()

	for labelsHash, singleMetric := range c.collection {
		if singleMetric.LastUpdate.Add(c.mapping.TTL).Before(now) {
			delete(c.collection, labelsHash)
		}
	}
}

func hashLabels(labels []string) uint64 {
	// TODO(nabokihms): declare hasher once
	// TODO(nabokihms): consider better hashing
	hasher := fnv.New64a()
	var hashbuf bytes.Buffer

	for _, labelValue := range labels {
		hashbuf.WriteString(labelValue)
		hashbuf.WriteByte(labelsSeparator)
	}

	_, _ = hasher.Write(hashbuf.Bytes())
	return hasher.Sum64()
}
