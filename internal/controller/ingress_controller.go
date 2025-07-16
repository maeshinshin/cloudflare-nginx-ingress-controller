/*
Copyright 2025.

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

package controller

import (
	"context"

	networkingv1 "k8s.io/api/networking/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
)

// IngressReconciler reconciles a Ingress object
type IngressReconciler struct {
	client.Client
	Scheme                           *runtime.Scheme
	IsDefaultIngressClassEnabled     bool
	IngressClassName                 string
	NginxIngressClassName            string
	NginxIngressServiceName          string
	CloudflareTunnelIngressClassName string
}

// +kubebuilder:rbac:groups=networking.k8s.io,resources=ingresses,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=networking.k8s.io,resources=ingresses/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=networking.k8s.io,resources=ingresses/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the Ingress object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.20.2/pkg/reconcile
func (r *IngressReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := logf.FromContext(ctx)
	logger.Info("Reconciling Integrated Ingress", "name", req.Name, "namespace", req.Namespace)

	var ing networkingv1.Ingress
	if err := r.Get(ctx, req.NamespacedName, &ing); err != nil {
		if apierrors.IsNotFound(err) {
			logger.Info("Ingress resource not found, skipping reconciliation", "name", req.Name, "namespace", req.Namespace)
			return ctrl.Result{}, nil
		} else {
			logger.Error(err, "Failed to get Ingress resource", "name", req.Name, "namespace", req.Namespace)
			return ctrl.Result{}, err
		}
	}

	if !r.IsDefaultIngressClassEnabled && ing.Spec.IngressClassName != nil && *ing.Spec.IngressClassName != r.IngressClassName {
		return ctrl.Result{}, nil
	}

	if !ing.DeletionTimestamp.IsZero() {
		logger.Info("Ingress resource is being deleted, skipping reconciliation", "name", req.Name, "namespace", req.Namespace)
		if err := r.deleteFinalizers(ctx, &ing); err != nil {
			logger.Error(err, "Failed to delete finalizers from Ingress", "name", req.Name, "namespace", req.Namespace)
			return ctrl.Result{}, err
		}
	}

	return r.reconcile(ctx, ing)
}

func (r *IngressReconciler) reconcile(ctx context.Context, ing networkingv1.Ingress) (ctrl.Result, error) {
	logger := logf.FromContext(ctx)
	if err := r.reconcileIngress(ctx, ing); err != nil {
		logger.Error(err, "Failed to reconcile Ingress", "name", ing.Name, "namespace", ing.Namespace)
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *IngressReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&networkingv1.Ingress{}).
		Named("ingress").
		Complete(r)
}
