package controller

const (
	FINALIZER_NAME    = "integrated-ingress-controller.maeshinshin.github.io/finalizer"
	FIELDMANAGER_NAME = "integrated-ingress-controller"

	nginxIngressAnnotationPrefix            = "nginx.ingress.kubernetes.io"
	cloudflareTunnelIngressAnnotationPrefix = "cloudflare-tunnel-ingress-controller.strrl.dev"
	otherwizeAnnotation                     = "otherwize"
)
