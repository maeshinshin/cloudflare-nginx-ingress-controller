package controller

import (
	"context"
	"fmt"

	corev1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
	"k8s.io/apimachinery/pkg/api/equality"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	corev1apply "k8s.io/client-go/applyconfigurations/core/v1"
	metav1apply "k8s.io/client-go/applyconfigurations/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
)

func (r *IngressReconciler) reconcileExternalNameService(ctx context.Context, ing networkingv1.Ingress) error {
	var err error

	logger := logf.FromContext(ctx)
	externalName := r.getNginxIngressServiceDomain()
	labels := getLabels(ing)

	currentExternalNameService := &corev1.Service{}
	if err := r.Get(ctx, client.ObjectKey{Name: r.NginxIngressServiceName, Namespace: ing.Namespace}, currentExternalNameService); err != nil {
		if apierrors.IsNotFound(err) {
			logger.Info("ExternalName Service not found, creating a new one", "name", r.NginxIngressServiceName, "namespace", ing.Namespace)
		} else {
			return fmt.Errorf("failed to get ExternalName Service: %w", err)
		}
	}

	var ownerRef *metav1apply.OwnerReferenceApplyConfiguration
	if ownerRef, err = r.controllerReference(&ing); err != nil {
		return fmt.Errorf("failed to get controller reference: %w", err)
	}

	externalNameService := corev1apply.Service(
		r.NginxIngressServiceName,
		ing.Namespace,
	).
		WithOwnerReferences(ownerRef).
		WithLabels(labels).
		WithSpec(
			corev1apply.ServiceSpec().
				WithType(corev1.ServiceTypeExternalName).
				WithExternalName(externalName),
		)

	obj, err := runtime.DefaultUnstructuredConverter.ToUnstructured(externalNameService)
	if err != nil {
		return fmt.Errorf("failed to convert ExternalName Service to unstructured: %w", err)
	}

	patch := &unstructured.Unstructured{
		Object: obj,
	}

	currApplyConfig, err := corev1apply.ExtractService(currentExternalNameService, FIELDMANAGER_NAME)
	if err != nil {
		return fmt.Errorf("failed to extract current ExternalName Service apply configuration: %w", err)
	}

	if equality.Semantic.DeepEqual(currApplyConfig, externalNameService) {
		logger.Info("ExternalName Service is up to date, no changes needed", "name", r.NginxIngressServiceName, "namespace", ing.Namespace)
		return nil
	}

	if err := r.Patch(ctx, patch, client.Apply, client.FieldOwner(FIELDMANAGER_NAME), client.ForceOwnership); err != nil {
		return fmt.Errorf("failed to apply ExternalName Service: %w", err)
	}

	logger.Info("Created or updated ExternalName Service", "name", r.NginxIngressServiceName, "namespace", ing.Namespace)

	return nil
}
