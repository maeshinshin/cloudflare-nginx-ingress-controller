package controller

import (
	"context"
	"fmt"

	networkingv1 "k8s.io/api/networking/v1"
	metav1apply "k8s.io/client-go/applyconfigurations/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client/apiutil"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
)

func (r *IngressReconciler) deleteFinalizers(ctx context.Context, ing *networkingv1.Ingress) error {
	logger := logf.FromContext(ctx)
	if controllerutil.ContainsFinalizer(ing, FINALIZER_NAME) {
		logger.Info("Removing finalizer from Ingress", "name", ing.Name, "namespace", ing.Namespace)
		controllerutil.RemoveFinalizer(ing, FINALIZER_NAME)
		if err := r.Update(ctx, ing); err != nil {
			logger.Error(err, "Failed to remove finalizer from Ingress", "name", ing.Name, "namespace", ing.Namespace)
			return err
		}
	} else {
		logger.Info("Finalizer not found on Ingress, nothing to remove", "name", ing.Name, "namespace", ing.Namespace)
	}
	return nil
}

func (r *IngressReconciler) getNginxIngressServiceDomain() string {
	return fmt.Sprintf("%s.%s.svc.cluster.local", r.NginxIngressServiceName, r.Namespace)
}

func (r *IngressReconciler) controllerReference(ing *networkingv1.Ingress) (*metav1apply.OwnerReferenceApplyConfiguration, error) {
	gvk, err := apiutil.GVKForObject(ing, r.Scheme)
	if err != nil {
		return nil, fmt.Errorf("failed to get GVK for Ingress: %w", err)
	}

	return metav1apply.OwnerReference().
			WithAPIVersion(gvk.GroupVersion().String()).
			WithKind(gvk.Kind).
			WithName(ing.Name).
			WithUID(ing.UID).
			WithBlockOwnerDeletion(true).
			WithController(true),
		nil
}
