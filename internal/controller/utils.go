package controller

import (
	"fmt"
	"maps"
	"strings"

	networkingv1 "k8s.io/api/networking/v1"
)

type (
	component int
)

func getNginxIngressName(ing networkingv1.Ingress) string {
	return fmt.Sprintf("%s-%s-nginx-ingress", ing.Namespace, ing.Name)
}

func getCloudflareTunnelIngressName(ing networkingv1.Ingress) string {
	return fmt.Sprintf("%s-%s-cf-tunnel-ingress", ing.Namespace, ing.Name)
}

func getAnnotations(ing networkingv1.Ingress) map[string]map[string]string {
	annotations := make(map[string]map[string]string)
	for key, value := range ing.Annotations {
		if strings.HasPrefix(key, nginxIngressAnnotationPrefix) {
			if _, exists := annotations[nginxIngressAnnotationPrefix]; !exists {
				annotations[nginxIngressAnnotationPrefix] = make(map[string]string)
			}
			annotations[nginxIngressAnnotationPrefix][key] = value
		} else if strings.HasPrefix(key, cloudflareTunnelIngressAnnotationPrefix) {
			if _, exists := annotations[cloudflareTunnelIngressAnnotationPrefix]; !exists {
				annotations[cloudflareTunnelIngressAnnotationPrefix] = make(map[string]string)
			}
			annotations[cloudflareTunnelIngressAnnotationPrefix][key] = value
		} else {
			if _, exists := annotations[otherwizeAnnotation]; !exists {
				annotations[otherwizeAnnotation] = make(map[string]string)
			}
			annotations[otherwizeAnnotation][key] = value
		}
	}
	return annotations
}

func getLabels(ing networkingv1.Ingress) map[string]string {
	labels := make(map[string]string)
	labels["managed-by"] = "integrated-ingress-controller"
	if ing.Labels != nil {
		maps.Copy(labels, ing.Labels)
	}

	return labels
}
