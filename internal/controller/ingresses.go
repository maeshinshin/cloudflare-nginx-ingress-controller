package controller

import (
	"context"
	"fmt"
	"maps"

	networkingv1 "k8s.io/api/networking/v1"
	"k8s.io/apimachinery/pkg/api/equality"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	corev1apply "k8s.io/client-go/applyconfigurations/core/v1"
	metav1apply "k8s.io/client-go/applyconfigurations/meta/v1"
	networkingv1apply "k8s.io/client-go/applyconfigurations/networking/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
)

func (r *IngressReconciler) reconcileIngress(ctx context.Context, ing networkingv1.Ingress) error {
	annotations := getAnnotations(ing)

	annotationsForNginxIngress := make(map[string]string)
	annotationsForCloudflareTunnelIngress := make(map[string]string)
	maps.Copy(annotationsForNginxIngress, annotations[nginxIngressAnnotationPrefix])
	maps.Copy(annotationsForNginxIngress, annotations[otherwizeAnnotation])
	maps.Copy(annotationsForCloudflareTunnelIngress, annotations[cloudflareTunnelIngressAnnotationPrefix])

	if err := r.reconcileNginxIngress(ctx, ing, annotationsForNginxIngress); err != nil {
		return err
	}

	if err := r.reconcileCloudflareTunnelIngress(ctx, ing, annotationsForCloudflareTunnelIngress); err != nil {
		return err
	}

	return nil
}

func (r *IngressReconciler) reconcileNginxIngress(ctx context.Context, ing networkingv1.Ingress, annotations map[string]string) error {
	var err error

	logger := logf.FromContext(ctx)
	nginxIngressName := getNginxIngressName(ing)
	labels := getLabels(ing)

	currentNginxIngress := &networkingv1.Ingress{}
	if err = r.Get(ctx, client.ObjectKey{Name: nginxIngressName, Namespace: ing.Namespace}, currentNginxIngress); err != nil {
		if apierrors.IsNotFound(err) {
			logger.Info("Nginx Ingress not found, creating a new one", "name", nginxIngressName, "namespace", ing.Namespace)
		} else {
			return fmt.Errorf("failed to get Nginx Ingress: %w", err)
		}
	}

	var ownerRef *metav1apply.OwnerReferenceApplyConfiguration
	if ownerRef, err = r.controllerReference(&ing); err != nil {
		return fmt.Errorf("failed to get controller reference: %w", err)
	}

	nginxIngressSpec := networkingv1apply.IngressSpec().
		WithIngressClassName(r.nginxIngressClassName)

	if ing.Spec.DefaultBackend != nil {
		defaultBackend := networkingv1apply.IngressBackend()
		if ing.Spec.DefaultBackend.Service != nil {
			defaultBackend.
				WithService(
					networkingv1apply.IngressServiceBackend().
						WithName(ing.Spec.DefaultBackend.Service.Name).
						WithPort(
							networkingv1apply.ServiceBackendPort().
								WithName(ing.Spec.DefaultBackend.Service.Port.Name).
								WithNumber(ing.Spec.DefaultBackend.Service.Port.Number),
						),
				)
		} else if ing.Spec.DefaultBackend.Resource != nil {
			defaultBackend.
				WithResource(
					corev1apply.TypedLocalObjectReference().
						WithName(ing.Spec.DefaultBackend.Resource.Name).
						WithKind(ing.Spec.DefaultBackend.Resource.Kind).
						WithAPIGroup(*ing.Spec.DefaultBackend.Resource.APIGroup),
				)
		}
		nginxIngressSpec.WithDefaultBackend(defaultBackend)
	}

	if ing.Spec.Rules != nil {
		rules := make([]*networkingv1apply.IngressRuleApplyConfiguration, 0, len(ing.Spec.Rules))
		for _, rule := range ing.Spec.Rules {

			var paths []*networkingv1apply.HTTPIngressPathApplyConfiguration
			for _, path := range rule.HTTP.Paths {
				httpBackend := networkingv1apply.IngressBackend()

				if path.Backend.Service != nil {
					httpBackend.WithService(
						networkingv1apply.IngressServiceBackend().
							WithName(path.Backend.Service.Name).
							WithPort(
								networkingv1apply.ServiceBackendPort().
									WithName(path.Backend.Service.Port.Name).
									WithNumber(path.Backend.Service.Port.Number),
							),
					)
				} else if path.Backend.Resource != nil {
					httpBackend.WithResource(
						corev1apply.TypedLocalObjectReference().
							WithName(path.Backend.Resource.Name).
							WithKind(path.Backend.Resource.Kind).
							WithAPIGroup(*path.Backend.Resource.APIGroup),
					)
				}

				paths = append(paths, networkingv1apply.HTTPIngressPath().
					WithPath(path.Path).
					WithBackend(httpBackend).
					WithPathType(*path.PathType),
				)
			}

			rules = append(rules, networkingv1apply.IngressRule().
				WithHost(rule.Host).
				WithHTTP(
					networkingv1apply.HTTPIngressRuleValue().
						WithPaths(paths...),
				),
			)
		}

		nginxIngressSpec.WithRules(rules...)
	}

	if ing.Spec.TLS != nil {
		ingressTLS := make([]*networkingv1apply.IngressTLSApplyConfiguration, 0, len(ing.Spec.TLS))
		for _, tls := range ing.Spec.TLS {
			ingressTLS = append(ingressTLS, networkingv1apply.IngressTLS().
				WithHosts(tls.Hosts...).
				WithSecretName(tls.SecretName),
			)
		}
		nginxIngressSpec.WithTLS(ingressTLS...)
	}

	nginxIngress := networkingv1apply.Ingress(
		nginxIngressName,
		ing.Namespace,
	).
		WithOwnerReferences(ownerRef).
		WithAnnotations(annotations).
		WithLabels(labels).
		WithSpec(nginxIngressSpec)

	obj, err := runtime.DefaultUnstructuredConverter.ToUnstructured(nginxIngress)
	if err != nil {
		return fmt.Errorf("failed to convert Nginx Ingress to unstructured: %w", err)
	}

	patch := &unstructured.Unstructured{
		Object: obj,
	}

	currApplyConfig, err := networkingv1apply.ExtractIngress(currentNginxIngress, FIELDMANAGER_NAME)
	if err != nil {
		return fmt.Errorf("failed to extract current Nginx Ingress apply configuration: %w", err)
	}

	if equality.Semantic.DeepEqual(currApplyConfig, nginxIngress) {
		logger.Info("Nginx Ingress is up to date, no changes needed", "name", nginxIngressName, "namespace", ing.Namespace)
		return nil
	}

	if err := r.Patch(ctx, patch, client.Apply, client.FieldOwner(FIELDMANAGER_NAME), client.ForceOwnership); err != nil {
		return fmt.Errorf("failed to apply Nginx Ingress: %w", err)
	}

	logger.Info("Created or updated Nginx Ingress", "name", nginxIngressName, "namespace", ing.Namespace)

	return nil
}

func (r *IngressReconciler) reconcileCloudflareTunnelIngress(ctx context.Context, ing networkingv1.Ingress, annotations map[string]string) error {
	var err error

	logger := logf.FromContext(ctx)
	cloudflareTunnelIngressName := getCloudflareTunnelIngressName(ing)
	labels := getLabels(ing)

	currentCloudflareTunnelIngress := &networkingv1.Ingress{}
	if err = r.Get(ctx, client.ObjectKey{Name: cloudflareTunnelIngressName, Namespace: ing.Namespace}, currentCloudflareTunnelIngress); err != nil {
		if apierrors.IsNotFound(err) {
			logger.Info("Cloudflare Tunnel Ingress not found, creating a new one", "name", cloudflareTunnelIngressName, "namespace", ing.Namespace)
		} else {
			return fmt.Errorf("failed to get Cloudflare Tunnel Ingress: %w", err)
		}
	}

	var ownerRef *metav1apply.OwnerReferenceApplyConfiguration
	if ownerRef, err = r.controllerReference(&ing); err != nil {
		return fmt.Errorf("failed to get controller reference: %w", err)
	}

	cloudflareTunnelIngressSpec := networkingv1apply.IngressSpec().
		WithIngressClassName(r.cloudflareTunnelIngressClassName)

	tlsHosts := make(map[string]struct{}, 0)
	if ing.Spec.TLS != nil {
		for _, tls := range ing.Spec.TLS {
			for _, host := range tls.Hosts {
				tlsHosts[host] = struct{}{}
			}
		}
	}

	rules := make([]*networkingv1apply.IngressRuleApplyConfiguration, 0, len(ing.Spec.Rules))
	for _, rule := range ing.Spec.Rules {
		host := rule.Host
		_, isTLSEnabled := tlsHosts[host]

		var backendPort int32 = 80
		if isTLSEnabled {
			backendPort = 443
		}

		rules = append(
			rules,
			networkingv1apply.IngressRule().
				WithHost(host).
				WithHTTP(
					networkingv1apply.HTTPIngressRuleValue().
						WithPaths(
							networkingv1apply.HTTPIngressPath().
								WithPath("/").
								WithBackend(
									networkingv1apply.IngressBackend().
										WithService(
											networkingv1apply.IngressServiceBackend().
												WithName(r.nginxIngressServiceName).
												WithPort(
													networkingv1apply.ServiceBackendPort().
														WithNumber(backendPort),
												),
										),
								).
								WithPathType(networkingv1.PathTypePrefix),
						),
				),
		)
	}
	cloudflareTunnelIngressSpec.WithRules(rules...)

	cloudflareTunnelIngress := networkingv1apply.Ingress(
		cloudflareTunnelIngressName,
		ing.Namespace,
	).
		WithOwnerReferences(ownerRef).
		WithAnnotations(annotations).
		WithLabels(labels).
		WithSpec(cloudflareTunnelIngressSpec)

	obj, err := runtime.DefaultUnstructuredConverter.ToUnstructured(cloudflareTunnelIngress)
	if err != nil {
		return fmt.Errorf("failed to convert Cloudflare Tunnel Ingress to unstructured: %w", err)
	}

	patch := &unstructured.Unstructured{
		Object: obj,
	}

	currApplyConfig, err := networkingv1apply.ExtractIngress(currentCloudflareTunnelIngress, FIELDMANAGER_NAME)
	if err != nil {
		return fmt.Errorf("failed to extract current Cloudflare Tunnel Ingress apply configuration: %w", err)
	}

	if equality.Semantic.DeepEqual(currApplyConfig, cloudflareTunnelIngress) {
		logger.Info("Cloudflare Tunnel Ingress is up to date, no changes needed", "name", cloudflareTunnelIngressName, "namespace", ing.Namespace)
		return nil
	}

	if err := r.Patch(ctx, patch, client.Apply, client.FieldOwner(FIELDMANAGER_NAME), client.ForceOwnership); err != nil {
		return fmt.Errorf("failed to apply Cloudflare Tunnel Ingress: %w", err)
	}

	logger.Info("Created or updated Cloudflare Tunnel Ingress", "name", cloudflareTunnelIngressName, "namespace", ing.Namespace)

	return nil
}
