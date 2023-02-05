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
package main

import (
	"flag"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/prometheus/common/log"

	"github.com/nabokihms/events_exporter/pkg/kube"
	"github.com/nabokihms/events_exporter/pkg/server"
	"github.com/nabokihms/events_exporter/pkg/vault"
)

func main() {
	var (
		exporterAddress    = ":9000"
		logLevel           = "info"
		kubeconfig         = ""
		fieldSelector      = ""
		omitEventsMessages = false
		eventsTTL          = time.Hour
	)

	flag.StringVar(&exporterAddress, "server.exporter-address", exporterAddress, "Address to export prometheus metrics")
	flag.StringVar(&logLevel, "server.log-level", logLevel, "Log level (logs all incoming events if debug)")
	flag.StringVar(&kubeconfig, "kube.config", kubeconfig, "Path to kubeconfig (optional)")
	flag.StringVar(&fieldSelector, "kube.field-selector", fieldSelector, "Events filter as for kubectl")
	flag.BoolVar(&omitEventsMessages, "kube.omit-events-messages", omitEventsMessages, "Do not expose message field from events (it reduces cardinality)")
	flag.DurationVar(&eventsTTL, "kube.events-ttl", eventsTTL, "For how long to keep stale events")

	flag.Parse()

	if err := log.Base().SetFormat("logger:stdout?json=true"); err != nil {
		log.Fatalf("error formating logger: %v", err)
	}

	if err := log.Base().SetLevel(logLevel); err != nil {
		log.Fatalf("set log level: %v", err)
	}

	errorCh := make(chan error)
	stopCh := make(chan struct{})

	metricsVault := vault.NewVault()
	err := metricsVault.RegisterMappings([]vault.Mapping{kube.EventMapping(eventsTTL)})
	if err != nil {
		log.Fatalf("mappings registration: %v", err)
	}

	informer, err := kube.NewEventsInformer(kubeconfig, fieldSelector, kube.EventCallback(metricsVault, omitEventsMessages))
	if err != nil {
		log.Fatalf("kubernetes informer: %v", err)
	}

	// TODO(nabokihms): Ensure that after starting informer we clear all stale events before starting web server
	go func() {
		informer.Run(stopCh, errorCh)
		metricsVault.RemoveStaleMetrics()
	}()

	metricsServer := server.NewMetricsServer()
	go metricsServer.Start(exporterAddress, errorCh)

	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, os.Interrupt, syscall.SIGTERM)

	// TODO (nabokihms): check that every concurrent task stops correctly
	tick := time.NewTicker(time.Second)
	for {
		select {
		case <-tick.C:
			// TODO(nabokihms): think about setting tombstones instead of deleting
			metricsVault.RemoveStaleMetrics()
		case s := <-signalChan:
			log.Warnf("signal received: %v, exiting...", s)
			close(stopCh)
			metricsServer.Close()
			tick.Stop()
			os.Exit(0)
		case e := <-errorCh:
			log.Errorf("error received: %v", e)
			close(stopCh)
			metricsServer.Close()
			tick.Stop()
			os.Exit(1)
		}
	}
}
