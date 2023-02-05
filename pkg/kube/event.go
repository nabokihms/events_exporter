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

const maxMessageLen = 200

func trimMessage(message string) string {
	if len(message) <= maxMessageLen {
		return message
	}
	return message[:maxMessageLen]
}

// EventToSample converts Kubernetes core v1.Event to the prometheus metric sample.
func EventToSample(event *v1.Event, omitEventsMessages bool) vault.Sample {
	var message string
	if !omitEventsMessages {
		message = trimMessage(event.Message)
	}

	return vault.Sample{
		ID:    string(event.UID),
		Value: float64(event.Count),
		Labels: []string{
			/* type */ event.Type,
			/* source_component */ event.Source.Component,
			/* source_host */ event.Source.Host,
			/* involved_kind */ event.InvolvedObject.Kind,
			/* involved_name */ event.InvolvedObject.Name,
			/* involved_namespace */ event.InvolvedObject.Namespace,
			/* reporting_controller */ event.ReportingController,
			/* reporting_instance */ event.ReportingInstance,
			/* reason */ event.Reason,
			/* message */ message,
		},
		Timestamp: event.LastTimestamp.Local(),
	}
}

// EventCallback generates the handler to connect prometheus metrics vault to the shared event informer.
func EventCallback(vault *vault.MetricsVault, omitEventsMessages bool) func(obj interface{}) {
	return func(obj interface{}) {
		log.With("event", obj).Debug("received event")

		event := obj.(*v1.Event)
		if err := vault.Store("kube_event_info", EventToSample(event, omitEventsMessages)); err != nil {
			log.Errorf("collecting event: %v", err)
		}
	}
}

// EventMapping creates the mapping for the prometheus metrics vault. The order of the labels here should match the one
// from the sample converter function.
func EventMapping(ttl time.Duration) vault.Mapping {
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
			"reporting_controller",
			"reporting_instance",
			"reason",
			"message",
		},
		TTL: ttl,
	}
}
