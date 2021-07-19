#!/bin/bash

fedcluster=kind-cluster-fed
memberlist=(
    "kind-cluster-1"
    "kind-cluster-2"
    "kind-cluster-3"
    )
namespace=test



if [ "$1" = "clean" ]; then

    kubectl --context $fedcluster -n $namespace delete federatedobjects.types.kubefed.io --all
    for ctx in "${memberlist[@]}"; do
        kubectl --context $ctx -n $namespace delete federatedobjects.types.kubefed.io --all
    done

elif [ "$1" = "list" ]; then

    kubectl --context $fedcluster -n $namespace get federatedobjects.types.kubefed.io -oyaml
    for ctx in "${memberlist[@]}"; do
        kubectl --context $ctx -n $namespace get federatedobjects.types.kubefed.io -oyaml
    done

else
    echo "Starting at $(date)"
    for c in $(eval echo {1..$1}); do
cat <<EOF | kubectl --context $fedcluster apply -f - &
apiVersion: types.kubefed.io/v1beta1
kind: FederatedObject
metadata:
    name: federatedobject-$c
    namespace: $namespace
spec:
    placement:
        clusters:
        - name: cluster-1
        - name: cluster-2
        - name: cluster-3
EOF
      sleep 0.1
    done

fi

