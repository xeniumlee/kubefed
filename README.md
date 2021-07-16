## Prepare
```
curl -Lo /usr/local/bin/kind https://kind.sigs.k8s.io/dl/v0.11.1/kind-linux-amd64
chmod +x /usr/local/bin/kind

curl -Lo /usr/local/bin/kubefedctl https://github.com/kubernetes-sigs/kubefed/releases/download/v0.8.1/kubefedctl-0.8.1-linux-amd64.tgz
chmod +x /usr/local/bin/kubefedctl
```

## Init
```
sed -i 's/IPADDR/<your ip address>/g' manifest/kind-config.yaml

kind create cluster --name cluster-fed --config=manifest/kind-config.yaml
kind create cluster --name cluster-1 --config=manifest/kind-config.yaml
kind create cluster --name cluster-2 --config=manifest/kind-config.yaml
kind create cluster --name cluster-3 --config=manifest/kind-config.yaml
```

## Setup
### Setup fed cluster
```
helm upgrade -i kubefed ./charts --create-namespace -n kube-federation-system --kube-context kind-cluster-fed
kubefedctl --host-cluster-context kind-cluster-fed join cluster-1 --cluster-context kind-cluster-1 --v=2
kubefedctl --host-cluster-context kind-cluster-fed join cluster-2 --cluster-context kind-cluster-2 --v=2
kubefedctl --host-cluster-context kind-cluster-fed join cluster-3 --cluster-context kind-cluster-3 --v=2

kubectl --context kind-cluster-fed -n kube-federation-system get kubefedclusters
```

### Setup member clusters
```
helm upgrade -i kubefed ./charts --create-namespace -n kube-federation-system --kube-context kind-cluster-1
helm upgrade -i kubefed ./charts --create-namespace -n kube-federation-system --kube-context kind-cluster-2
helm upgrade -i kubefed ./charts --create-namespace -n kube-federation-system --kube-context kind-cluster-3
```

## Test
```
```

## Cleanup
```
kind delete cluster --name cluster-fed
kind delete cluster --name cluster-1
kind delete cluster --name cluster-2
kind delete cluster --name cluster-3
```

## Develop
```
kubebuilder init --domain kubefed.io --repo github.com/xeniumlee/kubefed
kubebuilder edit --multigroup=true

kubebuilder create api --group types --version v1beta1 --kind FederatedObject
kubebuilder create api --group core --version v1beta1 --kind KubeFedCluster
```