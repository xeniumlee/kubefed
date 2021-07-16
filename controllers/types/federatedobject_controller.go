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

package types

import (
	"context"

	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/log"

	typesv1beta1 "github.com/xeniumlee/kubefed/apis/types/v1beta1"
	"github.com/xeniumlee/kubefed/util"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// FederatedObjectReconciler reconciles a FederatedObject object
type FederatedObjectReconciler struct {
	client.Client
	Scheme      *runtime.Scheme
	ClusterName string
}

//+kubebuilder:rbac:groups=types.kubefed.io,resources=federatedobjects,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=types.kubefed.io,resources=federatedobjects/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=types.kubefed.io,resources=federatedobjects/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the FederatedObject object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.8.3/pkg/reconcile
func (r *FederatedObjectReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.FromContext(ctx)

	obj := &typesv1beta1.FederatedObject{}
	if err := r.Get(ctx, req.NamespacedName, obj); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	if r.ClusterName == util.FederationClusterName {

		clusters := obj.Spec.Placement.Clusters

		if obj.Status == nil {
			obj.Status = make([]typesv1beta1.ClusterStatus, len(clusters))
			for i, c := range clusters {
				obj.Status[i] = typesv1beta1.ClusterStatus{
					Name: c.Name,
				}
			}
			err := r.Status().Update(ctx, obj)
			logger.Info("Got", "target clusters", clusters, "status", obj.Status)
			return ctrl.Result{}, err
		}

		return ctrl.Result{}, nil

	} else {

		if obj.Status == nil {
			return ctrl.Result{}, nil
		}

		clusters := obj.Spec.Placement.Clusters
		for i, c := range clusters {
			if c.Name == r.ClusterName && obj.Status[i].Timestamp.IsZero() {
				obj.Status[i].Timestamp = metav1.Now()

				err := r.Status().Update(ctx, obj)
				logger.Info("Update timestamp", "cluster", r.ClusterName)
				return ctrl.Result{}, err
			}
		}

		return ctrl.Result{}, nil
	}
}

// SetupWithManager sets up the controller with the Manager.
func (r *FederatedObjectReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&typesv1beta1.FederatedObject{}).
		WithOptions(controller.Options{MaxConcurrentReconciles: 5}).
		Complete(r)
}
