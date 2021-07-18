## Prepare
```
curl -Lo /usr/local/bin/kind https://kind.sigs.k8s.io/dl/v0.11.1/kind-linux-amd64
chmod +x /usr/local/bin/kind

curl -Lo /usr/local/bin/kubefedctl https://github.com/kubernetes-sigs/kubefed/releases/download/v0.8.1/kubefedctl-0.8.1-linux-amd64.tgz
chmod +x /usr/local/bin/kubefedctl
```

## Init
```
sed -i 's/IPADDR/<your ip address>/g' config/samples/kind-config.yaml

kind create cluster --name cluster-fed --config=config/samples/kind-config.yaml
kind create cluster --name cluster-1 --config=config/samples/kind-config.yaml
kind create cluster --name cluster-2 --config=config/samples/kind-config.yaml
kind create cluster --name cluster-3 --config=config/samples/kind-config.yaml
```

## Setup
### Setup fed cluster
```
helm upgrade -i kubefed ./charts --create-namespace -n kube-federation-system --kube-context kind-cluster-fed


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

## TODO
- [Finalizers](https://book.kubebuilder.io/reference/using-finalizers.html)
- [Event filter](https://stuartleeks.com/posts/kubebuilder-event-filters-part-2-update)
- [Http client](https://www.loginradius.com/blog/async/tune-the-go-http-client-for-high-performance)
- [Leader election]()

## Develop
```
## Init project
kubebuilder init --domain kubefed.io --repo github.com/xeniumlee/kubefed
kubebuilder edit --multigroup=true

kubebuilder create api --group types --version v1beta1 --kind FederatedObject
kubebuilder create api --group core --version v1beta1 --kind KubeFedCluster

## Install CRD to federatioin cluster
make install

## For kubefedctl join
kubectl apply -f config/samples/crd-kubefed-config.yaml
kubectl apply -f config/samples/kubefed-config.yaml

kubefedctl --host-cluster-context kind-cluster-fed join cluster-1 --cluster-context kind-cluster-1 --kubefed-namespace test -v 2
kubefedctl --host-cluster-context kind-cluster-fed join cluster-2 --cluster-context kind-cluster-2 --kubefed-namespace test -v 2
kubefedctl --host-cluster-context kind-cluster-fed join cluster-3 --cluster-context kind-cluster-3 --kubefed-namespace test -v 2

kubectl --context kind-cluster-fed -n test get kubefedclusters

kubectl --context kind-cluster-1 apply -f config/crd/bases/types.kubefed.io_federatedobjects.yaml
kubectl --context kind-cluster-2 apply -f config/crd/bases/types.kubefed.io_federatedobjects.yaml
kubectl --context kind-cluster-3 apply -f config/crd/bases/types.kubefed.io_federatedobjects.yaml

kubectl --context kind-cluster-fed apply -f config/samples/types_v1beta1_federatedobject.yaml

kubectl --context kind-cluster-fed -n test delete federatedobjects.types.kubefed.io federatedobject-1
kubectl --context kind-cluster-1 -n test delete federatedobjects.types.kubefed.io federatedobject-1
kubectl --context kind-cluster-2 -n test delete federatedobjects.types.kubefed.io federatedobject-1
kubectl --context kind-cluster-3 -n test delete federatedobjects.types.kubefed.io federatedobject-1

kubectl --context kind-cluster-fed -n test get federatedobjects.types.kubefed.io federatedobject-1 -oyaml
kubectl --context kind-cluster-1 -n test get federatedobjects.types.kubefed.io federatedobject-1 -oyaml
kubectl --context kind-cluster-2 -n test get federatedobjects.types.kubefed.io federatedobject-1 -oyaml
kubectl --context kind-cluster-3 -n test get federatedobjects.types.kubefed.io federatedobject-1 -oyaml


./bin/manager --kubeconfig /root/.kube/config --clustername cluster-fed
./bin/manager --kubeconfig /root/.kube/cluster-1 --clustername cluster-1
./bin/manager --kubeconfig /root/.kube/cluster-2 --clustername cluster-2
./bin/manager --kubeconfig /root/.kube/cluster-3 --clustername cluster-3

```