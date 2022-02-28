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

### Docker Container

Ready-to-use Docker images are [available on GitHub](https://github.com/nabokihms/events_exporter/pkgs/container/events_exporter).

```bash
docker pull ghcr.io/nabokihms/events_exporter:latest
```

### Helm Chart

The first version of helm chart is available.
1. Clone this repo.
2. Install the chart:
    ```bash
    helm upgrade --install events-exporter ./charts/events_exporter
    ```
    __Note__: If you want to install the chart to the custom namespace, you need to create it first.


3. After the installation, metrics will be available on address `http://events-exporter.default:9000/metrics`



## Alerts and Dashboards

TBA
