# namespacelabel-operator

This Kubernetes operator takes labels as spec and syncs those labels to exist on the specified namespace.

## Example
```sh
apiVersion: dana.io.dana.io/v1alpha1
kind: NamespaceLabel
metadata:
  labels:
  namespace: example-namespace
  name: namespacelabel-sample
spec:
  labels:
        one: one
        two: two
        m: h
        f: c
```

## Getting Started

You’ll need a Kubernetes cluster to run against. You can use [KIND](https://sigs.k8s.io/kind) to get a local cluster for testing, or run against a remote cluster.
**Note:** Your controller will automatically use the current context in your kubeconfig file (i.e. whatever cluster `kubectl cluster-info` shows).

### Running on the cluster

1. Install Instances of Custom Resources:

```sh
kubectl apply -f config/samples/
```

2. Build and push your image to the location specified by `IMG`:

```sh
make docker-build docker-push IMG=<some-registry>/namespacelabel-operator:tag
```

3. Deploy the controller to the cluster with the image specified by `IMG`:

```sh
make deploy IMG=<some-registry>/namespacelabel-operator:tag
```

### Uninstall CRDs

To delete the CRDs from the cluster:

```sh
make uninstall
```

# API Reference

## Packages
- [dana.io.dana.io/v1alpha1](#danaiodanaiov1alpha1)


## dana.io.dana.io/v1alpha1

Package v1alpha1 contains API Schema definitions for the  v1alpha1 API group

### Resource Types
- [NamespaceLabel](#namespacelabel)
- [NamespaceLabelList](#namespacelabellist)



#### NamespaceLabel



NamespaceLabel is the Schema for the namespacelabels API

_Appears in:_
- [NamespaceLabelList](#namespacelabellist)

| Field | Description |
| --- | --- |
| `apiVersion` _string_ | `dana.io.dana.io/v1alpha1`
| `kind` _string_ | `NamespaceLabel`
| `metadata` _[ObjectMeta](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.25/#objectmeta-v1-meta)_ | Refer to Kubernetes API documentation for fields of `metadata`. |
| `spec` _[NamespaceLabelSpec](#namespacelabelspec)_ |  |


#### NamespaceLabelList



NamespaceLabelList contains a list of NamespaceLabels



| Field | Description |
| --- | --- |
| `apiVersion` _string_ | `dana.io.dana.io/v1alpha1`
| `kind` _string_ | `NamespaceLabelList`
| `metadata` _[ListMeta](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.25/#listmeta-v1-meta)_ | Refer to Kubernetes API documentation for fields of `metadata`. |
| `items` _[NamespaceLabel](#namespacelabel) array_ |  |


#### NamespaceLabelSpec



NamespaceLabelSpec defines the desired state of NamespaceLabel

_Appears in:_
- [NamespaceLabel](#namespacelabel)

| Field | Description |
| --- | --- |
| `labels` _object (keys:string, values:string)_ | Labels is a map of key-value pairs representing the labels the NamespaceLabel syncs |






### License

Copyright 2022.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
