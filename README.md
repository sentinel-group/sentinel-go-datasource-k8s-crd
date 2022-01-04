# Sentinel Go Kubernetes CRD data-source

## Generate CRD definitions

Generate Sentinel rule CRDs with:

```shell
make manifests
```

## Install CRD definitions:

Install Sentinel rule CRDs with:

```shell
kubectl apply -f config/crd/bases
```

## Samples

- [Sentinel CRD YAML samples](./config/samples)