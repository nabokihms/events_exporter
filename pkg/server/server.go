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

package server

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"time"

	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/prometheus/common/log"
)

// MetricsServer is a http server which serves prometheus metrics from the metrics vault.
type MetricsServer struct {
	srv *http.Server
}

// NewMetricsServer returns a metrics server instance.
func NewMetricsServer() *MetricsServer {
	return &MetricsServer{srv: &http.Server{}}
}

// Start runs metrics server with typical exporter handlers.
func (m *MetricsServer) Start(address string, errorCh chan error) {
	http.Handle("/metrics", promhttp.Handler())

	http.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("ok"))
	})

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		_, err := fmt.Fprintf(w, `<!DOCTYPE html>
			<title>Events Exporter</title>
			<h1>Events Exporter</h1>
			<p><a href=%q>Metrics</a></p>`,
			"/metrics")
		if err != nil {
			log.Warnf("Error while sending a response for the '/' path: %v", err)
			return
		}
	})

	listener, err := net.Listen("tcp", address)
	if err != nil {
		errorCh <- err
		return
	}

	log.Infof("start exporting metrics on %q", address)
	errorCh <- m.srv.Serve(listener)
}

func (m *MetricsServer) Close() {
	log.Info("closing metrics server ...")
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	_ = m.srv.Shutdown(ctx)
}
