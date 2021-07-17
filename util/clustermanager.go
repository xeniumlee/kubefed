/*
Copyright 2021.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package util

import (
	"context"
	"net/http"
	"net/url"
	"sync"

	"github.com/pkg/errors"
	"k8s.io/client-go/tools/clientcmd"

	corev1beta1 "github.com/xeniumlee/kubefed/apis/core/v1beta1"
	corev1 "k8s.io/api/core/v1"
	restclient "k8s.io/client-go/rest"

	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	ctrlclient "sigs.k8s.io/controller-runtime/pkg/client"
	ctrlmanager "sigs.k8s.io/controller-runtime/pkg/manager"
)

const (
	KubeAPIQPS   = 20.0
	KubeAPIBurst = 30
	TokenKey     = "token"
	CaCrtKey     = "ca.crt"
)

var (
	clusterClients map[string]ctrlclient.Client = make(map[string]ctrlclient.Client)
	clientLock     sync.RWMutex
)

func buildClusterConfig(
	cluster *corev1beta1.KubeFedCluster,
	namespace string,
	client ctrlclient.Client) (*restclient.Config, error) {

	clusterName := cluster.Name

	apiEndpoint := cluster.Spec.APIEndpoint
	if apiEndpoint == "" {
		return nil, errors.Errorf("The api endpoint of cluster %s is empty", clusterName)
	}

	secretName := cluster.Spec.SecretRef.Name
	if secretName == "" {
		return nil, errors.Errorf("Cluster %s does not have a secret name", clusterName)
	}
	secret := &corev1.Secret{}
	err := client.Get(context.TODO(), ctrlclient.ObjectKey{Namespace: namespace, Name: secretName}, secret)
	if err != nil {
		return nil, err
	}

	token, tokenFound := secret.Data[TokenKey]
	if !tokenFound || len(token) == 0 {
		return nil, errors.Errorf("The secret for cluster %s is missing a non-empty value for %q", clusterName, TokenKey)
	}

	clusterConfig, err := clientcmd.BuildConfigFromFlags(apiEndpoint, "")
	if err != nil {
		return nil, err
	}
	clusterConfig.CAData = cluster.Spec.CABundle
	clusterConfig.BearerToken = string(token)
	clusterConfig.QPS = KubeAPIQPS
	clusterConfig.Burst = KubeAPIBurst

	if cluster.Spec.ProxyURL != "" {
		proxyURL, err := url.Parse(cluster.Spec.ProxyURL)
		if err != nil {
			return nil, errors.Errorf("Failed to parse provided proxy URL %s: %v", cluster.Spec.ProxyURL, err)
		}
		clusterConfig.Proxy = http.ProxyURL(proxyURL)
	}

	return clusterConfig, nil
}

func NewManager(
	cluster *corev1beta1.KubeFedCluster,
	namespace string,
	client ctrlclient.Client,
	scheme *runtime.Scheme) (ctrlmanager.Manager, error) {

	config, err := buildClusterConfig(cluster, namespace, client)
	if err != nil {
		return nil, err
	}

	return ctrl.NewManager(config, ctrl.Options{
		Scheme:                 scheme,
		Namespace:              namespace,
		MetricsBindAddress:     "0",
		Port:                   0,
		HealthProbeBindAddress: "0",
		LeaderElection:         false,
	})
}

func AddclusterClient(
	clusterName string,
	client ctrlclient.Client) {
	clientLock.Lock()
	clusterClients[clusterName] = client
	clientLock.Unlock()
}

func GetclusterClient(clusterName string) ctrlclient.Client {
	clientLock.RLock()
	defer clientLock.RUnlock()
	if client, ok := clusterClients[clusterName]; ok {
		return client
	} else {
		return nil
	}
}
