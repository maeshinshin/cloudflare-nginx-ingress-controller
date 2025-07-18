# Integrated Ingress Controller

integrated-ingress-controller is a custom Kubernetes controller designed to orchestrate Cloudflare Tunnel and NGINX Ingress Controller to work in harmony.

Developers can define a single Ingress resource to automatically configure secure external access via Cloudflare and flexible internal routing and authentication via NGINX.

## 🌟 Features

- Simplified Ingress Management: Developers can expose services without worrying about complex configurations by simply specifying a single IngressClass provided by this controller.
- Automatic DNS Record Creation: Through integration with Cloudflare Tunnel, DNS records for hostnames defined in the Ingress resource are automatically created and managed on Cloudflare.
- Flexible Internal Routing & Authentication: It leverages NGINX Ingress Controller internally, allowing you to use its powerful features like path-based routing, Basic Authentication, and rewrite rules.

## 🔗 Dependencies

This controller leverages two powerful, community-maintained Ingress controllers, which are automatically installed as dependencies via Helm:

- [NGINX Ingress Controller](https://github.com/kubernetes/ingress-nginx): Used for all internal routing, authentication, and traffic management within the Kubernetes cluster.
- [Cloudflare Tunnel Ingress Controller](https://github.com/STRRL/cloudflare-tunnel-ingress-controller): Used to securely connect your cluster to the Cloudflare network and automatically manage public DNS records.

Our controller acts as an orchestrator layer on top of these two, simplifying their combined usage.

## 🏛️ Architecture

This controller monitors Ingress resources. When it finds a resource with its specific ingressClassName, it automatically generates two distinct Ingress resources, each with a different role.

```
                       ┌──────────────────────────────┐
                       │   Ingress created by User    │
                       │ (class: integrated-ingress)  │
                       └─────────────┬────────────────┘
                                     │ 1. Detected by Controller
                                     ▼
                  ┌────────────────────────────────────┐
                  │   Integrated Ingress Controller    │
                  └────────────────┬───────────────────┘
                                   │ 2. Generates two Ingresses
           ┌───────────────────────┴───────────────────────┐
           ▼                                               ▼
┌──────────────────────┐                      ┌──────────────────────┐
│ Ingress for          │                      │ Ingress for          │
│ Cloudflare Tunnel    │                      │ NGINX                │
│ (class: cloudflare)  │                      │ (class: nginx)       │
└──────────┬───────────┘                      └──────────┬───────────┘
           │                                             │
           │ Creates DNS & Forwards traffic to NGINX     │ Auth & Routes to App
           ▼                                             ▼
┌──────────────────────┐                      ┌──────────────────────┐
│ Cloudflare Tunnel    │ ───────────────>     │ NGINX Ingress        │
│ Ingress Controller   │                      │ Controller           │
└──────────────────────┘                      └──────────────────────┘


```

## 🚀 Installation

This controller can be easily installed using Helm.

### Prerequisites

- A running Kubernetes cluster
- The helm command-line tool

### Installation Steps

#### 1. Add the Helm repository.

```bash
helm repo add integrated-ingress https://maeshinshin.github.io/integrated-ingress-controller
```

#### 2. Update the repository.

```bash
helm repo update
```

#### 3. Install the Helm chart.

This command also installs its dependencies, nginx-ingress-controller and cloudflare-tunnel-ingress-controller.
Replace the <...> placeholders with your own Cloudflare information.

```bash
helm upgrade --install --wait \
  -n integrated-ingress-controller --create-namespace \
  integrated-ingress \
  integrated-ingress/integrated-ingress \
  --set=cloudflaretunnel.cloudflare.apiToken="<cloudflare-api-token>" \
  --set=cloudflaretunnel.cloudflare.accountId="<cloudflare-account-id>" \
  --set=cloudflaretunnel.cloudflare.tunnelName="<your-favorite-tunnel-name>"
```

## Usage
After installation, you can expose a service by creating an Ingress resource with ingressClassName set to integrated-ingress.

### Example Configuration

```yaml
apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  name: my-app-ingress
  annotations:
    # Set Basic Authentication for NGINX
    nginx.ingress.kubernetes.io/auth-type: "basic"
    nginx.ingress.kubernetes.io/auth-secret: "my-basic-auth-secret"
spec:
  # Specify the IngressClass handled by this controller
  ingressClassName: integrated-ingress
  rules:
    - host: my-app.your-domain.com
      http:
        paths:
          - path: /
            pathType: Prefix
            backend:
              service:
                name: my-app-service
                port:
                  number: 80
```

When you apply this manifest, the controller will automatically generate the necessary configurations for both Cloudflare and NGINX.

## License

This project is licensed under the Apache License 2.0.
