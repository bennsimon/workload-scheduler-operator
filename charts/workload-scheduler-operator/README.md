workload-scheduler-operator helm chart
===========



## TL;DR

```bash
$ helm repo add workload-scheduler-operator https://bennsimon.github.io/workload-scheduler-operator/
$ helm install workload-scheduler-operator workload-scheduler-operator/workload-scheduler-operator
```

## Introduction

This chart bootstraps  [workload-scheduler-operator](https://github.com/bennsimon/workload-scheduler-operator) deployment on a [Kubernetes](http://kubernetes.io) cluster using the [Helm](https://helm.sh) package manager.

## Prerequisites

- Kubernetes 1.16+
- Helm 3.1.0

## Installing the Chart

To install the chart with the release name `workload-scheduler-operator`:

## Configuration

The following table lists the configurable parameters of the Workload-scheduler-operator chart and their default values.

| Parameter                            | Description | Default                                   |
|--------------------------------------|-------------|-------------------------------------------|
| `replicaCount`                       |             | `1`                                       |
| `image.repository`                   |             | `"bennsimon/workload-scheduler-operator"` |
| `image.pullPolicy`                   |             | `"IfNotPresent"`                          |
| `image.tag`                          |             | `""`                                      |
| `imagePullSecrets`                   |             | `[]`                                      |
| `nameOverride`                       |             | `""`                                      |
| `fullnameOverride`                   |             | `""`                                      |
| `serviceAccount.create`              |             | `true`                                    |
| `serviceAccount.annotations`         |             | `{}`                                      |
| `serviceAccount.name`                |             | `""`                                      |
| `podAnnotations`                     |             | `{}`                                      |
| `podSecurityContext`                 |             | `{}`                                      |
| `securityContext`                    |             | `{}`                                      |
| `resources`                          |             | `{}`                                      |
| `livenessProbe.httpGet.path`         |             | `"/healthz"`                              |
| `livenessProbe.httpGet.port`         |             | `8081`                                    |
| `livenessProbe.initialDelaySeconds`  |             | `15`                                      |
| `livenessProbe.periodSeconds`        |             | `20`                                      |
| `readinessProbe.httpGet.path`        |             | `"/readyz"`                               |
| `readinessProbe.httpGet.port`        |             | `8081`                                    |
| `readinessProbe.initialDelaySeconds` |             | `5`                                       |
| `readinessProbe.periodSeconds`       |             | `10`                                      |
| `autoscaling.enabled`                |             | `false`                                   |
| `nodeSelector`                       |             | `{}`                                      |
| `tolerations`                        |             | `[]`                                      |
| `affinity`                           |             | `{}`                                      |
| `crds.enabled`                       |             | `true`                                    |
| `env`                                |             | `null`                                    |
