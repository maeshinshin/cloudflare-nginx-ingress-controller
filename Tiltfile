load('ext://deployment', 'deployment_create')
load('ext://restart_process', 'docker_build_with_restart')

IMG='controller:latest'

def init():
  def helmInit():
    def add_ingress_nginx_repo():
      return 'helm repo add ingress-nginx https://kubernetes.github.io/ingress-nginx'
    def add_strrl_dev_repo():
      return 'helm repo add strrl.dev https://helm.strrl.dev'
    def helm_repo_update():
      return 'helm repo update'
    return add_ingress_nginx_repo() + ' && ' + add_strrl_dev_repo() + ' && ' + helm_repo_update()

  def helminstall():
    def install_nginx_ingress():
      return 'helm upgrade --install ingress-nginx ingress-nginx/ingress-nginx --namespace nginx-ingress --create-namespace'
    def install_cloudflare_tunnel_ingress():
      -           return 'helm upgrade --install --wait -n cloudflare-tunnel-ingress-controller --create-namespace cloudflare-tunnel-ingress-controller strrl.dev/cloudflare-tunnel-ingress-controller --set=cloudflare.apiToken="",cloudflare.accountId="",cloudflare.tunnelName=""
    local_resource('install-nginx-ingress',install_nginx_ingress())
    local_resource('install-cloudflare-tunnel-ingress',install_cloudflare_tunnel_ingress())

  local_resource('helm_init', helmInit())
  helminstall()

def rebuild():
  def build():
    return 'docker build -t controller:latest .'

  def dockerload():
    return 'kind load docker-image controller:latest'
  local_resource('build',build())
  local_resource('load',dockerload())

init()

k8s_yaml(kustomize('./config/dev'))

rebuild()

deployment_create(
  name='test-nginx',
  image='nginx:latest',
  namespace='default',
  ports='80'
)



