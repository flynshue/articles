# articles
Demo golang api

## Running container
```bash
podman pull public.ecr.aws/flynshue/articles:v0.1.1
```
```bash
podman run -d -p 8080:5000 public.ecr.aws/flynshue/articles:v0.1.1
```

## Deploying to OCP using raw manifests
This uses ocp routes, which is why it's ocp specific.
```bash
oc apply -f deployment.yaml
```

## Deploying to k8s or ocp using helm chart
```bash
helm install -n <namespace> <helm release> charts/articles
```
Here's an example of deploying helm release articles to namespace test-flynshue
```bash
helm install -n test-flynshue articles charts/articles
```

To install with ocp route
```bash
helm install -n test-flynshue articles charts/articles \
--set route.enabled=true
```

To install with ingress
<!-- Will probably need to add values file here to make things easier -->
Create values file for chart. Here's an example using ingress-nginx and cert-manager
```bash
vim articles-values.yaml
ingress:
  enabled: true
  annotations:
    kubernetes.io/ingress.class: "nginx"
    cert-manager.io/cluster-issuer: "letsencrypt-staging"
  hosts:
    - host: articles.apps.seadogslab.com
      paths:
        - "/"
  tls:
   - secretName: articles-ingress-cert
     hosts:
       - articles.apps.seadogslab.com
```


```bash
helm install -n test-flynshue articles charts/articles \
--values=articles-values.yaml
```

To update an existing helm release
```bash
helm upgrade -n <namespace> <helm release> charts/articles
```

## Deploying the helm chart using argocd gitops
```bash
argocd app create articles \
--repo https://github.com/flynshuePersonal/articles.git \
--path charts/articles-chart --dest-namespace flynshue \
--dest-server https://kubernetes.default.svc
```

