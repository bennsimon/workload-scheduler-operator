# workload-scheduler-operator

[![License](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](https://opensource.org/licenses/Apache-2.0)  [![Go](https://github.com/bennsimon/workload-scheduler-operator/actions/workflows/go.yaml/badge.svg?branch=main)](https://github.com/bennsimon/workload-scheduler-operator/actions/workflows/go.yaml)

This operator scales various kubernetes workloads to a desired number of replicas based on a schedule.

## Description

The operator introduces 3 custom resources to handle its logic:

### Schedule

In this resource one defines a period. It takes in list of `scheduleUnits` which are used to define part(s) of a period. The scheduleUnits are described by the days and start and end time and/or date.

> If the days section in a `scheduleUnit` is not specified or left empty then all the days of the week will take its place. i.e. it will behave like a wildcard.

The name for the schedule i.e. value of `metadata.name` is used in the [WorkloadSchedule](#workloadschedule) CR.

> Date format: yyyy-MM-dd e.g. 2023-07-25

> Time Format: HH:mm:ss e.g. 9:00:00

Dates can take placeholders i.e `y`, `m` and `d` each will be replaced by the value of that day. e.g. `y-07-12` will be converted to `2023-07-12` if the year of that day is 2023.

The custom resource takes the form below:

```yaml
apiVersion: workload-scheduler.bennsimon.github.io/v1
kind: Schedule
metadata:
  labels:
    app.kubernetes.io/name: schedule
    app.kubernetes.io/instance: schedule-sample
    app.kubernetes.io/part-of: workload-scheduler-operator
    app.kubernetes.io/managed-by: kustomize
    app.kubernetes.io/created-by: workload-scheduler-operator
  name: weekday
spec:
  scheduleUnits:
    - days: # optional if not specified it will be replaced with *
        - "Monday"
        - "Tuesday"
        - "Wednesday"
        - "Thursday"
        - "Friday"
      start:
        time: "9:00:00" #  optional if not specified defaults to 00:00:00
        date: "2023-09-08" #  optional if not specified defaults to that day's date.
      end:
        time: "9:00:00" #  optional if not specified defaults to 00:00:00
        date: "2023-09-08" #  optional if not specified defaults to that day's date.
```

### WorkloadSchedule

This is the resource where one specifies the workload(s) and schedule(s) with which action to perform on a particular schedule. It takes in selectors and schedules; currently the supported selectors are `namespace`, `name`, `kind` and `labels`. The schedules section is used to specify the list of schedules with the desired (replica count) value of that particular period.
The order of definition in the schedules section determines which schedule has more priority (FIFO).

The custom resource takes the form below:

```yaml
apiVersion: workload-scheduler.bennsimon.github.io/v1
kind: WorkloadSchedule
metadata:
  labels:
    app.kubernetes.io/name: workloadschedule
    app.kubernetes.io/instance: workloadschedule-sample
    app.kubernetes.io/part-of: workload-scheduler-operator
    app.kubernetes.io/managed-by: kustomize
    app.kubernetes.io/created-by: workload-scheduler-operator
  name: workloadschedule-sample
spec:
  selector:
    namespaces: # optional, if not specified it would be replaced with *, i.e. act on all namespaces
      - "default"
    names: # optional, if not specified it would be replaced with *, i.e. act on all names
      - "server-a"
    kinds: # optional, if not specified it would be replaced with *, i.e. act on all kinds, currently supported kinds are StatefulSet and Deployment
      - "deployment"
#    labels: # optional, if not specified its null
#      app.kubernetes.io/name: "redis"
  schedules:
    - schedule: "always-up"
      desired: 1
    - schedule: "holidays"
      desired: 0
    - schedule: "weekday"
      desired: 2
```

### Configuration

#### Container Environment Configuration

| Configuration             | Description                                                                                              | Default       |
|---------------------------|----------------------------------------------------------------------------------------------------------|---------------|
| `TZ`                      | Specifies the timezone used.                                                                             | `UTC`         |
| `NAMESPACES_OFF_LIMITS`   | Specifies lists of namespaces (comma separated) that should be ignored by the operator.                  | `kube-system` |
| `RECONCILIATION_DURATION` | Specifies the duration in seconds at which cluster workloads are reconciled with the workload schedules. | `60`          |
| `DEBUG`                   | Shows the additional info logs for debugging purposes.                                                   | `false`       |

## Deployment

### On existing cluster

To deploy the operator you will need the following manifests:

*   serviceaccount
*   clusterrole
*   clusterrolebinding
*   deployment
*   schedule
*   workloadschedule
*   crds
    *   [schedules.yaml](config/crd/bases/workload-scheduler.bennsimon.github.io\_schedules.yaml)
    *   [workloadschedules.yaml](config/crd/bases/workload-scheduler.bennsimon.github.io\_workloadschedules.yaml)
    
    * Use the command below at the root on this repository (i.e. after cloning) to deploy crds:
    ````
    kubectl apply -k config/crd/
    ````

Below is the snippet of the yaml files you would need to deploy the operator.  (for crds check command above)

```yaml
---
apiVersion: v1
kind: ServiceAccount
metadata:
  labels:
    app.kubernetes.io/name: serviceaccount
    app.kubernetes.io/instance: controller-manager
    app.kubernetes.io/component: rbac
    app.kubernetes.io/created-by: workload-scheduler-operator
    app.kubernetes.io/part-of: workload-scheduler-operator
  name: workload-scheduler-operator
---
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRole
metadata:
  creationTimestamp: null
  name: workload-scheduler-operator
rules:
  - apiGroups:
      - apps
    resources:
      - deployments
    verbs:
      - get
      - list
      - update
      - watch
  - apiGroups:
      - apps
    resources:
      - statefulsets
    verbs:
      - get
      - list
      - update
      - watch
  - apiGroups:
      - workload-scheduler.bennsimon.github.io
    resources:
      - schedules
    verbs:
      - create
      - delete
      - get
      - list
      - patch
      - update
      - watch
  - apiGroups:
      - workload-scheduler.bennsimon.github.io
    resources:
      - schedules/finalizers
    verbs:
      - update
  - apiGroups:
      - workload-scheduler.bennsimon.github.io
    resources:
      - schedules/status
    verbs:
      - get
      - patch
      - update
  - apiGroups:
      - workload-scheduler.bennsimon.github.io
    resources:
      - workloadschedules
    verbs:
      - create
      - delete
      - get
      - list
      - patch
      - update
      - watch
  - apiGroups:
      - workload-scheduler.bennsimon.github.io
    resources:
      - workloadschedules/finalizers
    verbs:
      - update
  - apiGroups:
      - workload-scheduler.bennsimon.github.io
    resources:
      - workloadschedules/status
    verbs:
      - get
      - patch
      - update
---
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRoleBinding
metadata:
  labels:
    app.kubernetes.io/name: clusterrolebinding
    app.kubernetes.io/instance: manager-rolebinding
    app.kubernetes.io/component: rbac
    app.kubernetes.io/created-by: workload-scheduler-operator
    app.kubernetes.io/part-of: workload-scheduler-operator
  name: workload-scheduler-operator
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: workload-scheduler-operator
subjects:
  - kind: ServiceAccount
    name: workload-scheduler-operator
    namespace: default # update this to preferred namespace
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: workload-scheduler-operator
  labels:
    control-plane: controller-manager
    app.kubernetes.io/name: deployment
    app.kubernetes.io/instance: controller-manager
    app.kubernetes.io/component: manager
    app.kubernetes.io/created-by: workload-scheduler-operator
    app.kubernetes.io/part-of: workload-scheduler-operator
spec:
  selector:
    matchLabels:
      control-plane: controller-manager
  replicas: 1
  template:
    metadata:
      annotations:
        kubectl.kubernetes.io/default-container: manager
      labels:
        control-plane: controller-manager
    spec:
      containers:
        - command:
            - /manager
          env:
#            - name: TZ
#              value: "Africa/Nairobi"
#            - name: NAMESPACES_OFF_LIMITS
#              value: "cert-manager"
#            - name: RECONCILIATION_DURATION
#              value: "60"
          image: bennsimon/workload-scheduler-operator:tag
          name: manager
          securityContext:
            allowPrivilegeEscalation: false
            capabilities:
              drop:
                - "ALL"
          livenessProbe:
            httpGet:
              path: /healthz
              port: 8081
            initialDelaySeconds: 15
            periodSeconds: 20
          readinessProbe:
            httpGet:
              path: /readyz
              port: 8081
            initialDelaySeconds: 5
            periodSeconds: 10
          # TODO(user): Configure the resources accordingly based on the project requirements.
          # More info: https://kubernetes.io/docs/concepts/configuration/manage-resources-containers/
          resources:
            limits:
              cpu: 500m
              memory: 128Mi
            requests:
              cpu: 10m
              memory: 64Mi
      serviceAccountName: workload-scheduler-operator
      terminationGracePeriodSeconds: 10
```

### Getting Started with development

You’ll need a Kubernetes cluster to run against. You can use [KIND](https://sigs.k8s.io/kind) to get a local cluster for testing, or run against a remote cluster.
**Note:** Your controller will automatically use the current context in your kubeconfig file (i.e. whatever cluster `kubectl cluster-info` shows).

### Running on the cluster

1.  Install Instances of Custom Resources:

```sh
kubectl apply -f config/samples/
```

2.  Build and push your image to the location specified by `IMG`:

```sh
make docker-build docker-push IMG=<some-registry>/workload-scheduler-operator:tag
```

3.  Deploy the controller to the cluster with the image specified by `IMG`:

```sh
make deploy IMG=<some-registry>/workload-scheduler-operator:tag
```

### Uninstall CRDs

To delete the CRDs from the cluster:

```sh
make uninstall
```

### Undeploy controller

UnDeploy the controller from the cluster:

```sh
make undeploy
```

## Contributing

// TODO(user): Add detailed information on how you would like others to contribute to this project

### How it works

This project aims to follow the Kubernetes [Operator pattern](https://kubernetes.io/docs/concepts/extend-kubernetes/operator/).

It uses [Controllers](https://kubernetes.io/docs/concepts/architecture/controller/),
which provide a reconcile function responsible for synchronizing resources until the desired state is reached on the cluster.

### Test It Out

1.  Install the CRDs into the cluster:

```sh
make install
```

2.  Run your controller (this will run in the foreground, so switch to a new terminal if you want to leave it running):

```sh
make run
```

**NOTE:** You can also run this in one step by running: `make install run`

### Modifying the API definitions

If you are editing the API definitions, generate the manifests such as CRs or CRDs using:

```sh
make manifests
```

**NOTE:** Run `make --help` for more information on all potential `make` targets

More information can be found via the [Kubebuilder Documentation](https://book.kubebuilder.io/introduction.html)

## License

Copyright 2023.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
