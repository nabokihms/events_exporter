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
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/prometheus/common/log"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

func getClient(kubeconfigPath string) (kubernetes.Interface, error) {
	var (
		cfg *rest.Config
		err error
	)
	if kubeconfigPath != "" {
		cfg, err = clientcmd.BuildConfigFromFlags("", kubeconfigPath)
		if err != nil {
			return nil, fmt.Errorf("new kubernetes client from config %s: %w", kubeconfigPath, err)
		}
		log.Infof("using kubeconfig from file: %q", kubeconfigPath)
	} else {
		cfg, err = rest.InClusterConfig()
		switch {
		case err == nil:
			log.Infof("using in-cluster kubeconfig")
		case !errors.Is(err, rest.ErrNotInCluster):
			return nil, fmt.Errorf("new kubernetes from cluster: %w", err)
		default:
			home, _ := os.UserHomeDir()
			userKubeconfigPath := filepath.Join(home, ".kube", "config")

			cfg, err = clientcmd.BuildConfigFromFlags("", userKubeconfigPath)
			if err != nil {
				return nil, fmt.Errorf("new kubernetes client from homedir %s: %w", userKubeconfigPath, err)
			}
			log.Infof("using kubeconfig from homedir: %q", userKubeconfigPath)
		}
	}

	return kubernetes.NewForConfig(cfg)
}
