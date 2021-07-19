## Prepare
```
curl -Lo /usr/local/bin/kind https://kind.sigs.k8s.io/dl/v0.11.1/kind-linux-amd64
chmod +x /usr/local/bin/kind

curl -Lo /usr/local/bin/kubefedctl https://github.com/kubernetes-sigs/kubefed/releases/download/v0.8.1/kubefedctl-0.8.1-linux-amd64.tgz
chmod +x /usr/local/bin/kubefedctl
```

## Init
```
sed -i 's/IPADDR/<your ip address>/g' test/kind-config.yaml

kind create cluster --name cluster-fed --config=test/kind-config.yaml
kind create cluster --name cluster-1 --config=test/kind-config.yaml
kind create cluster --name cluster-2 --config=test/kind-config.yaml
kind create cluster --name cluster-3 --config=test/kind-config.yaml
```

## Install
```
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -o bin/manager main.go
cd bin
docker build -t kubefed:0.0.1 -f ../Dockerfile.prebuilt .
kind load docker-image kubefed:0.0.1 --name cluster-fed
kind load docker-image kubefed:0.0.1 --name cluster-1
kind load docker-image kubefed:0.0.1 --name cluster-2
kind load docker-image kubefed:0.0.1 --name cluster-3

docker exec -it cluster-3-control-plane crictl images

helm upgrade -i kubefed ./charts --create-namespace -n test --kube-context kind-cluster-fed --set config.cluster=cluster-fed
helm upgrade -i kubefed ./charts --create-namespace -n test --kube-context kind-cluster-1 --set config.cluster=cluster-1
helm upgrade -i kubefed ./charts --create-namespace -n test --kube-context kind-cluster-2 --set config.cluster=cluster-2
helm upgrade -i kubefed ./charts --create-namespace -n test --kube-context kind-cluster-3 --set config.cluster=cluster-3
```

## Setup
```
kubefedctl --host-cluster-context kind-cluster-fed join cluster-1 --cluster-context kind-cluster-1 --kubefed-namespace test -v 2
kubefedctl --host-cluster-context kind-cluster-fed join cluster-2 --cluster-context kind-cluster-2 --kubefed-namespace test -v 2
kubefedctl --host-cluster-context kind-cluster-fed join cluster-3 --cluster-context kind-cluster-3 --kubefed-namespace test -v 2

kubectl --context kind-cluster-fed -n test get kubefedclusters
```

## Test
```
test/test.sh 1000
test/test.sh list
test/test.sh clean
```

## Cleanup
```
kind delete cluster --name cluster-fed
kind delete cluster --name cluster-1
kind delete cluster --name cluster-2
kind delete cluster --name cluster-3
```

## Design
![Design](/test/kubefed.jpg)

## TODO
- [Finalizers](https://book.kubebuilder.io/reference/using-finalizers.html)
- [Http client](https://www.loginradius.com/blog/async/tune-the-go-http-client-for-high-performance)
- Reconcile/Retry

## Develop
```
kubebuilder init --domain kubefed.io --repo github.com/xeniumlee/kubefed
kubebuilder edit --multigroup=true

kubebuilder create api --group types --version v1beta1 --kind FederatedObject
kubebuilder create api --group core --version v1beta1 --kind KubeFedCluster
```