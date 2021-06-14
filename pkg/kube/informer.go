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
	"context"
	"fmt"
	"time"

	"github.com/prometheus/common/log"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
)

const defaultSyncPeriod = 10 * time.Minute

type EventsInformer struct {
	client   kubernetes.Interface
	informer cache.SharedIndexInformer

	eventHandler func(object interface{})
}

func NewEventsInformer(kubeconfigPath, fieldSelector string, handler func(object interface{})) (*EventsInformer, error) {
	client, err := getClient(kubeconfigPath)
	if err != nil {
		return nil, err
	}

	informer := cache.NewSharedIndexInformer(
		// TODO: allow filter events based on metav1.ListOptions
		&cache.ListWatch{
			ListFunc: func(options metav1.ListOptions) (runtime.Object, error) {
				opts := metav1.ListOptions{FieldSelector: fieldSelector}
				return client.CoreV1().Events(metav1.NamespaceAll).List(context.TODO(), opts)
			},
			WatchFunc: func(options metav1.ListOptions) (watch.Interface, error) {
				opts := metav1.ListOptions{FieldSelector: fieldSelector}
				return client.CoreV1().Events(metav1.NamespaceAll).Watch(context.TODO(), opts)
			},
		},
		&v1.Event{},
		defaultSyncPeriod,
		cache.Indexers{cache.NamespaceIndex: cache.MetaNamespaceIndexFunc},
	)

	return &EventsInformer{client: client, informer: informer, eventHandler: handler}, nil
}

func (e *EventsInformer) Run(stopCh <-chan struct{}, errorCh chan<- error) {
	e.informer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: e.eventHandler,
		UpdateFunc: func(act, new interface{}) {
			e.eventHandler(new)
		},
	})
	err := e.informer.SetWatchErrorHandler(func(_ *cache.Reflector, err error) {
		errorCh <- fmt.Errorf("watch handler: %w", err)
	})
	if err != nil {
		errorCh <- fmt.Errorf("set watch handler: %w", err)
	}

	go e.informer.Run(stopCh)

	if ok := cache.WaitForCacheSync(stopCh, e.informer.HasSynced); !ok {
		errorCh <- fmt.Errorf("informer cache is not synced")
		return
	}

	<-stopCh
	log.Info("stopping informer ...")
}
