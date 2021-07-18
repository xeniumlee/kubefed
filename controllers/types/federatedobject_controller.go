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

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"

	typesv1beta1 "github.com/xeniumlee/kubefed/apis/types/v1beta1"
	"github.com/xeniumlee/kubefed/util"
)

// FederatedObjectReconciler reconciles a FederatedObject object
type FederatedObjectReconciler struct {
	client.Client
	Scheme            *runtime.Scheme
	ClusterName       string
	TargetClusterName string
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
	//logger := log.FromContext(ctx)

	obj := &typesv1beta1.FederatedObject{}
	if err := r.Get(ctx, req.NamespacedName, obj); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	if r.ClusterName == util.FederationClusterName {
		// federation cluster
		if r.TargetClusterName != "" {
			// watch member cluster
			util.TryNotify(r.TargetClusterName, req.NamespacedName, obj)
		} else {
			// watch federation cluster
			util.TryNotify(r.ClusterName, req.NamespacedName, obj)
		}
		return ctrl.Result{}, nil
	} else {
		// member cluster
		if obj.Status != nil {
			return ctrl.Result{}, nil
		}

		obj.Status = append(obj.Status, typesv1beta1.ClusterStatus{
			Name:      r.ClusterName,
			Timestamp: metav1.Now(),
		})

		err := r.Status().Update(ctx, obj)
		return ctrl.Result{}, err
	}
}

// SetupWithManager sets up the controller with the Manager.
func (r *FederatedObjectReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&typesv1beta1.FederatedObject{}).
		WithOptions(controller.Options{MaxConcurrentReconciles: 1}).
		Complete(r)
}
