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
	"fmt"

	"github.com/go-logr/logr"
	"github.com/prometheus/common/log"
	"k8s.io/klog/v2"
)

func init() {
	klog.SetLogger(&logger{})
}

// logger is an implementation of logr.Logger interface.
// It allows us to override standard klog output with desired logger (from prometheus).
var _ logr.Logger = (*logger)(nil)

type logger struct{}

func (l *logger) Enabled() bool {
	return true
}

func (l *logger) Info(msg string, keysAndValues ...interface{}) {
	log.Infof(msg, keysAndValues...)
}

func (l *logger) Error(err error, msg string, keysAndValues ...interface{}) {
	message := fmt.Sprintf(msg, keysAndValues...)
	if err != nil {
		message = fmt.Sprintf("%s: %v", message, err)
	}
	log.Error(message)
}

func (l *logger) V(_ int) logr.Logger {
	return l
}

func (l *logger) WithValues(_ ...interface{}) logr.Logger {
	return l
}

func (l *logger) WithName(_ string) logr.Logger {
	return l
}
