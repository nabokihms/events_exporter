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
package kube

import (
	"time"

	"github.com/prometheus/common/log"
	v1 "k8s.io/api/core/v1"

	"github.com/nabokihms/events_exporter/pkg/vault"
)

func EventToSample(event *v1.Event) vault.Sample {
	return vault.Sample{
		Value: float64(event.Count),
		Labels: []string{
			/* type */ event.Type,
			/* source_component */ event.Source.Component,
			/* source_host */ event.Source.Host,
			/* involved_kind */ event.InvolvedObject.Kind,
			/* involved_name */ event.InvolvedObject.Name,
			/* involved_namespace */ event.InvolvedObject.Namespace,
			/* reason */ event.Reason,
			/* message */ event.Message,
		},
	}
}

func EventCallback(vault *vault.MetricsVault) func(obj interface{}) {
	return func(obj interface{}) {
		event := obj.(*v1.Event)
		if err := vault.Store(0, EventToSample(event)); err != nil {
			log.Errorf("collecting event: %v", err)
		}
	}
}

func EventMapping() vault.Mapping {
	return vault.Mapping{
		Name: "kube_event_info",
		Help: "Expose Kubernetes events information",
		LabelNames: []string{
			"type",
			"source_component",
			"source_host",
			"involved_kind",
			"involved_name",
			"involved_namespace",
			"reason",
			"message",
		},
		TTL: time.Hour,
	}
}
