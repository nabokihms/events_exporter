# Events Exporter

## Overview

Prometheus-exporter, which converts Kubernetes events to Prometheus samples.

## Usage
```
Usage of events_exporter:
  -kube.config string
        Path to kubeconfig (optional)
  -kube.field-selector string
        Events filter as for kubectl
  -kube.omit-events-messages
        Do not expose message field from events (it reduces cardinality)
  -server.exporter-address string
        Address to export prometheus metrics (default ":9000")
  -server.log-level string
        Log level (default "info")
```

## Install

Ready-to-use Docker images are [available on GitHub](https://github.com/nabokihms/events_exporter/pkgs/container/events_exporter).

To run the exporter on top of Kubernetes, grant following permissions to the pod:
```yaml
---
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRole
metadata:
  name: events-exporter
rules:
- apiGroups: ["", "events.k8s.io"]
  resources: ["events"]
  verbs: ["get", "list", "watch"]
```

Helm chart will be available in the future releases.
