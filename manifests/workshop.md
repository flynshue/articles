# Workshop<!-- omit in toc -->
This is a hands on workshop to get the Data Infrastructure team started with understanding the basics of OCP.
- [Prerequisites](#prerequisites)
- [Deploying the Demo App](#deploying-the-demo-app)
  - [Scaling up more Replicas](#scaling-up-more-replicas)
  - [Spinning up a test pod and ResourceQuotas](#spinning-up-a-test-pod-and-resourcequotas)
  - [Communicating with the app](#communicating-with-the-app)
  - [Liveness and Readiness](#liveness-and-readiness)
  - [Working with app configuration](#working-with-app-configuration)
- [Next Steps](#next-steps)
## Prerequisites 
* OC CLI tools
* `demo-app-${USER}` namespace created in the sbx-app1 cluster using atlas-namespace helm-chart

## Deploying the Demo App
First log into sbx-app1 cluster

```bash
oc login -u $(whoami) ${SBX_APP1_URL}
```
**Note: use your okta creds**

```bash
cd manifests

oc project demo-app-${USER}

oc apply -f deployment.yaml
```

Let's take a look at the resources that were created from that manifest
```bash
[flynshue@flynshue-dell-latitude-7490 manifests]$ oc get deploy
NAME       READY     UP-TO-DATE   AVAILABLE   AGE
articles   1/1       1            1           2m5s

[flynshue@flynshue-dell-latitude-7490 manifests]$ oc get pods
NAME                        READY     STATUS    RESTARTS   AGE
articles-86c6478cd5-jpbbd   1/1       Running   0          119s

[flynshue@flynshue-dell-latitude-7490 manifests]$ oc get svc
NAME       TYPE        CLUSTER-IP       EXTERNAL-IP   PORT(S)    AGE
articles   ClusterIP   172.24.195.249   <none>        8080/TCP   2m8s

[flynshue@flynshue-dell-latitude-7490 manifests]$ oc get route
NAME       HOST/PORT                                                                    PATH      SERVICES   PORT      TERMINATION     WILDCARD
articles   articles-demo-app-flynshue.apps.sbx-app1.... ... 1 more             articles   http      edge/Redirect   None
```

### Scaling up more Replicas
Here's a few ways to update replicas

1. Edit the deployment spec using cli
> "The edit command allows you to directly edit any API resource you can retrieve via the command line tools. It will open
the editor defined by your OC _EDITOR, or EDITOR environment variables, or fall back to 'vi' for Linux or 'notepad' for
Windows"

```bash
oc edit deploy/articles
...
spec:
  progressDeadlineSeconds: 600
  replicas: 1 # <-- Change this
  revisionHistoryLimit: 10
  selector:
    matchLabels:
      app: articles
...
```
2. Edit the manifest file and re-apply
```bash
vim deployment.yaml
...
spec:
  progressDeadlineSeconds: 600
  replicas: 1 # <-- Change this
  revisionHistoryLimit: 10
  selector:
    matchLabels:
      app: articles
...

oc apply -f deployment.yaml
```

3. Use the `oc scale` command
```bash
oc scale deploy articles --replicas=3
```

Verify the deployment replicas
```bash
$ oc get deploy
NAME       READY     UP-TO-DATE   AVAILABLE   AGE
articles   3/3       3            3           15m

$ oc get pods
NAME                        READY     STATUS    RESTARTS   AGE
articles-86c6478cd5-79v58   1/1       Running   0          98s
articles-86c6478cd5-hrdc6   1/1       Running   0          98s
articles-86c6478cd5-jpbbd   1/1       Running   0          17m
```

### Spinning up a test pod and ResourceQuotas
Let's spin up another pod in your namespace so test the articles app

**Note: The following command will spin up a single pod, not managed by deployment.  The pod will be removed once you exit out of the terminal**
```bash
$ oc run ubuntu --rm -it --image=public.ecr.aws/lts/ubuntu:latest --command /bin/bash
Error from server (Forbidden): pods "ubuntu" is forbidden: exceeded quota: resourcequota, requested: limits.cpu=1, used: limits.cpu=3, limited: limits.cpu=3
```

Looks like we can't spin up another pod due to exceeded quota

When you created the namespaces using the atlas-namespace chart, it created `ResourceQuota` and `LimitRange` resources within the namespace.

ResourceQuota sets aggregate quota restrictions enforced per namespace
```bash
$ oc describe quota
Name:                      resourcequota
Namespace:                 demo-app-flynshue
Resource                   Used   Hard
--------                   ----   ----
configmaps                 2      50
limits.cpu                 3      3
limits.memory              3Gi    6Gi
openshift.io/imagestreams  0      20
persistentvolumeclaims     0      6
pods                       3      10
replicationcontrollers     0      50
requests.cpu               300m   2
requests.memory            300Mi  4Gi
requests.storage           0      10Gi
secrets                    5      5
services                   1      10
```

Looking at the resourcequota, we used up all of our limits.cpu.

What are the pods using for resources?
```bash
$ oc describe pods
Containers:
  articles:
    Container ID:   cri-o://7ba0120071511c3eadb227d9b72c2b32740c34825e5263b75a7ec34fb0a25ffc
    Image:          public.ecr.aws/flynshue/articles:v0.2.2
    Image ID:       public.ecr.aws/flynshue/articles@sha256:e65d46fbccf9e2a4b6a4dc32ea19d620d01cfaebc783edb72cc3d7c13a284c71
    Port:           <none>
    Host Port:      <none>
    State:          Running
      Started:      Tue, 25 Oct 2022 14:43:04 -0400
    Ready:          True
    Restart Count:  0
    Limits:
      cpu:     1
      memory:  1Gi
    Requests:
      cpu:        100m
      memory:     100Mi
```

If you look at the deployment.yaml manifest, we didn't define resources.
```bash
    spec:
      containers:
      - image: public.ecr.aws/flynshue/articles:v0.2.2
        name: articles
        resources: {}
```

So how are the pods getting the resources set?

Let's take a look at the limitrange
```bash
$ oc describe limitrange
Name:       limitrange
Namespace:  demo-app-flynshue
Type        Resource  Min  Max  Default Request  Default Limit  Max Limit/Request Ratio
----        --------  ---  ---  ---------------  -------------  -----------------------
Container   memory    -    -    100Mi            1Gi            -
Container   cpu       -    -    100m             1              -
```

Pods that do not have resources defined in their spec will use the resources defined in the limitrange

Let's update the resources on the articles deploy so that we can spin up a test pod.
```bash
vim deployment.yaml
...
    spec:
      containers:
      - image: public.ecr.aws/flynshue/articles:v0.2.2
        name: articles
        resources:
          requests:
            cpu: "100m"
            memory: "100Mi"
          limits:
            cpu: "500m"
            memory: "500Mi"
...
```

Update the articles deploy
```bash
oc apply -f deployment.yaml
```

The new pods will spin up and use the resources defined in our deployment manifest
```bash
oc describe pods
....
Containers:
  articles:
    Container ID:   cri-o://b7f4325953016bd714aab4350b5b8cb599c624d8ab991cc0ab5ed110efd84698
    Image:          public.ecr.aws/flynshue/articles:v0.2.2
    Image ID:       public.ecr.aws/flynshue/articles@sha256:e65d46fbccf9e2a4b6a4dc32ea19d620d01cfaebc783edb72cc3d7c13a284c71
    Port:           <none>
    Host Port:      <none>
    State:          Running
      Started:      Tue, 25 Oct 2022 15:23:16 -0400
    Ready:          True
    Restart Count:  0
    Limits:
      cpu:     500m
      memory:  500Mi
    Requests:
      cpu:        100m
      memory:     100Mi
...
```

Let's take look at the resourcequota too.
```bash
$ oc describe quota
Name:                      resourcequota
Namespace:                 demo-app-flynshue
Resource                   Used    Hard
--------                   ----    ----
configmaps                 2       50
limits.cpu                 1500m   3
limits.memory              1500Mi  6Gi
openshift.io/imagestreams  0       20
persistentvolumeclaims     0       6
pods                       3       10
replicationcontrollers     0       50
requests.cpu               300m    2
requests.memory            300Mi   4Gi
requests.storage           0       10Gi
secrets                    5       5
services                   1       10
```

We should be able to spin up our test pod now
```bash
$ oc run ubuntu --rm -it --image=public.ecr.aws/lts/ubuntu:latest --command /bin/bash
If you don't see a command prompt, try pressing enter.
root@ubuntu:/# 
```

We'll need to add some tools to the container
```
root@ubuntu:/# apt update -y

root@ubuntu:/# apt install wget curl -y
```

### Communicating with the app
So we have some pods running this articles app, but how do we talk to it?

Open up another terminal window
```bash
$ oc get pods -o wide
NAME                        READY     STATUS    RESTARTS   AGE       IP             NODE                          NOMINATED NODE   READINESS GATES
articles-59d7bf5655-cc759   1/1       Running   0          8m23s     172.23.4.94    sbx-app1-5cplr-worker-lx5jt   <none>           <none>
articles-59d7bf5655-ftf8x   1/1       Running   0          8m23s     172.20.5.124   sbx-app1-5cplr-worker-wplvn   <none>           <none>
articles-59d7bf5655-jn2vq   1/1       Running   0          12m       172.21.5.216   sbx-app1-5cplr-worker-4q2tt   <none>           <none>
ubuntu                      1/1       Running   0          3m52s     172.21.5.219   sbx-app1-5cplr-worker-4q2tt   <none>           <none>
```

The app is set to listen on port 5000, so we can use that port and one of the pod ip address to talk to one of the replicas from our test ubuntu pod.

Go back to the terminal running the ubuntu test container and send a request to the articles app.
```
root@ubuntu:/# curl 172.23.4.94:5000/articles
{"articles":[{"id":3,"title":"Article Title 3","desc":"Article Description 3","content":"Article Content 3"},{"id":1,"title":"Article Title 1","desc":"Article Description 1","content":"Article Content 1"},{"id":2,"title":"Article Title 2","desc":"Article Description 2","content":"Article Content 2"}]}
```

We have 3 replicas for the articles app, so talking directly to each pod isn't ideal.

Let's take a look at the articles service
```bash
oc get svc articles -o yaml
...
  ports:
  - name: http
    port: 8080
    protocol: TCP
    targetPort: 5000
  selector:
    app: articles
```

We can use the articles service on port 8080/TCP to target pods behind the service on port 5000/TCP.  It's basically a load balancer for the pods behind it.

```bash
$ oc get endpoints
NAME       ENDPOINTS                                              AGE
articles   172.20.5.124:5000,172.21.5.216:5000,172.23.4.94:5000   63m
```

When the service was created, it automatically creates endpoints behind it.  But how does the service know which pods to put in behind the service?

k8s uses selector labels to grab the pod ip addresses that will be used for the endpoints
```bash
  selector:
    app: articles
```

Let's take that same selector from the service and query the pods
```bash
$ oc get pods -l app=articles -o wide
NAME                        READY     STATUS    RESTARTS   AGE       IP             NODE                          NOMINATED NODE   READINESS GATES
articles-59d7bf5655-cc759   1/1       Running   0          21m       172.23.4.94    sbx-app1-5cplr-worker-lx5jt   <none>           <none>
articles-59d7bf5655-ftf8x   1/1       Running   0          21m       172.20.5.124   sbx-app1-5cplr-worker-wplvn   <none>           <none>
articles-59d7bf5655-jn2vq   1/1       Running   0          26m       172.21.5.216   sbx-app1-5cplr-worker-4q2tt   <none>           <none>
```

Notice that pod ip addresses match what's behind the articles endpoints

Open a third terminal window
```bash
$ oc get endpoints -w
NAME       ENDPOINTS                                              AGE
articles   172.20.5.124:5000,172.21.5.216:5000,172.23.4.94:5000   1h
```

From the second terminal window, lets delete one of the articles pods.
```bash
$ oc delete pod articles-59d7bf5655-cc759
pod "articles-59d7bf5655-cc759" deleted
```

Since our deployment is set to always have 3 replicas, another articles pod will spin up to replace the deleted one

You'll also notice from the third window that endpoint ip addresses gets updated
```bash
$ oc get endpoints -w
NAME       ENDPOINTS                                              AGE
articles   172.20.5.124:5000,172.21.5.216:5000,172.23.4.94:5000   1h
articles   172.20.5.124:5000,172.21.5.216:5000   1h
articles   172.20.5.124:5000,172.21.5.216:5000,172.23.4.95:5000   1h
```

Okay so how do we use the service?


Go back to your terminal running the ubuntu test pod
```
root@ubuntu:/# curl articles.demo-app-flynshue.svc.cluster.local:8080/articles
```

The service name convention is `SERVICE_NAME.NAMESPACE.svc.cluster.local:PORT`

The `svc.cluster.local` is optional
```
root@ubuntu:/# curl articles.demo-app-flynshue:8080/articles
```

Services allow you to talk to pods within the cluster.  We'll need use routes if we want use a client outside of the cluster to communicate with the articles app.

```bash
oc get routes articles -o yaml
...
spec:
  host: articles-demo-app-flynshue.apps.sbx-app1.lab1...
  port:
    targetPort: http
  tls:
    insecureEdgeTerminationPolicy: Redirect
    termination: edge
  to:
    kind: Service
    name: articles
    weight: 100
...
```

From another terminal, anything but the terminal running ubuntu test pod
```bash
$ ARTICLES_APP=$(oc get route articles -o jsonpath='{.spec.host}')

$ curl https://$ARTICLES_APP/articles 
{"articles":[{"id":3,"title":"Article Title 3","desc":"Article Description 3","content":"Article Content 3"},{"id":1,"title":"Article Title 1","desc":"Article Description 1","content":"Article Content 1"},{"id":2,"title":"Article Title 2","desc":"Article Description 2","content":"Article Content 2"}]}
```

### Liveness and Readiness
So far we've seen how we can scale our app in k8s and maintain a set number of replicas using the deployment manifest.  We also saw how we can leverage services as load balancer for the pods without manual intervention.

Next, we'll see how the liveness and readiness can be used to save an engineer from being paged out in the middle of the night.

Let's start with adding a liveness check to the deployment.
```bash
vim deployment.yaml
...
    spec:
      containers:
      - image: public.ecr.aws/flynshue/articles:v0.2.2
        name: articles
        resources:
          requests:
            cpu: "100m"
            memory: "100Mi"
          limits:
            cpu: "500m"
            memory: "500Mi"
        liveness:
          httpGet:
            path: /health
            port: 5000
...
```

For this example, this we're going to use a http check, but this could also be a command to check to see if a process is running.  The `/health` doesn't exist in the articles app and that is by design for this demo.

Now that we have the liveness check added, let update it.
```bash
$ oc apply -f deployment.yaml
```

Watch the new pods that were deployed and you'll see that they will start to restart
```
$ oc get pods -w
NAME                        READY     STATUS    RESTARTS   AGE
articles-5587979798-52hqg   1/1       Running   0          26s
articles-5587979798-8pc6t   1/1       Running   0          30s
articles-5587979798-tzcgm   1/1       Running   1          33s
articles-5587979798-8pc6t   1/1       Running   1         31s

```

Let's take a look at the namespace events for one of the pods restarting
```bash
$ oc get events | grep articles-5587979798-8pc6t
3s          Warning   Unhealthy                     pod/articles-5587979798-8pc6t    Liveness probe failed: HTTP probe failed with statuscode: 404
3s          Normal    Killing                       pod/articles-5587979798-8pc6t    Container articles failed liveness probe, will be restarted
```

The pods are restarting because it's failing the Liveness probe.

Imagine if we were running the app on VMs and it started failing the health check.  NOC would page out the on call engineer and they'd SSH into the VM and restart the app to fix issue and go back to asleep.

With k8s, if you set up proper liveness checks, k8s can detect when it's failing the health check and will automatically restart container in the pod, without waking up an engineer.

Let's fix the livenessProbe, to make the deployment happy.
```bash
vim deployment.yaml
...
        livenessProbe:
          httpGet:
            path: /healthz
            port: 5000
...

oc apply -f deployment.yaml
```

Now that the deployment is happy, let's add a readinessProbe that will fail.
```bash
vim deployment.yaml
...
        livenessProbe:
          httpGet:
            path: /healthz
            port: 5000
        readinessProbe:
          httpGet:
            path: /health
            port: 5000:
...

oc apply -f deployment.yaml
```

You'll see that there's pod that's not in ready and stopping the deployment from rolling out additional new pods
```bash
$ oc get pods
NAME                        READY     STATUS    RESTARTS   AGE
articles-5bf55c55c5-6sn5z   0/1       Running   0          42s # <-- New pod with failing readiness check
articles-7d7cb68f74-7cgkn   1/1       Running   0          3m23s
articles-7d7cb68f74-nm8bn   1/1       Running   0          3m18s
articles-7d7cb68f74-q8cpz   1/1       Running   0          3m20
```

I'm going to scale down the old pods using the replicasets. I won't go deep into this topic right now but deployments --> replicasets --> pods.  Think of replicasets as versioned deployments.

```bash
$ oc get rs
NAME                  DESIRED   CURRENT   READY     AGE
articles-5587979798   0         0         0         23m
articles-59d7bf5655   0         0         0         85m
articles-5bf55c55c5   1         1         0         9m16s # <--new version
articles-756597d5b7   0         0         0         26m
articles-7d7cb68f74   3         3         3         11m # <-- old version that we need to delete
articles-86c6478cd5   0         0         0         125m
```

```bash
$ oc scale rs/articles-7d7cb68f74 --replicas=0
replicaset.apps/articles-7d7cb68f74 scaled

$ oc get rs
NAME                  DESIRED   CURRENT   READY     AGE
articles-5587979798   0         0         0         24m
articles-59d7bf5655   0         0         0         86m
articles-5bf55c55c5   3         3         0         10m
articles-756597d5b7   0         0         0         27m
articles-7d7cb68f74   0         0         0         13m
articles-86c6478cd5   0         0         0         127m

$ oc get pods
NAME                        READY     STATUS    RESTARTS   AGE
articles-5bf55c55c5-6sn5z   0/1       Running   0          10m
articles-5bf55c55c5-mbm2z   0/1       Running   0          37s
articles-5bf55c55c5-tcw8w   0/1       Running   0          37s
```

Now we'll see 3 pods from the new deployment with the failing readinessProbe, but they're not restarting.

When a pod fails it's readinessProbe, k8s will remove the pods from service endpoints
```bash
$ oc describe endpoints
Name:         articles
Namespace:    demo-app-flynshue
Labels:       app=articles
Annotations:  <none>
Subsets:
  Addresses:          <none>
  NotReadyAddresses:  172.20.5.129,172.21.5.228,172.23.4.101
  Ports:
    Name  Port  Protocol
    ----  ----  --------
    http  5000  TCP
```

This useful in situations where that app itself is running but relies on connectivity to another service to fully operate, so we don't want to start sending traffic to the app until it's really ready.

Let's fix the readinessProbe
```bash
vim deployment.yaml
...
        livenessProbe:
          httpGet:
            path: /healthz
            port: 5000
        readinessProbe:
          httpGet:
            path: /healthz
            port: 5000:
...

oc apply -f deployment.yaml

$ oc describe endpoints
Name:         articles
Namespace:    demo-app-flynshue
Labels:       app=articles
Annotations:  <none>
Subsets:
  Addresses:          172.20.5.130,172.21.5.231,172.23.4.102
  NotReadyAddresses:  <none>
  Ports:
    Name  Port  Protocol
    ----  ----  --------
    http  5000  TCP

```

### Working with app configuration
What if our app needs to be configured?  We can use configmaps to add configuration to our apps.

For this demo we'll have to update the articles image version
```vim deployment.yaml
...
spec:
      containers:
      - image: public.ecr.aws/flynshue/articles:v0.4.0
        name: articles
...
```

Go back to your terminal that was running the ubuntu test pod.  Mine died so the pod was removed and I'll need to create a new pod, but this time I'll keep the pod running in the background.

```bash
oc run ubuntu --image=public.ecr.aws/lts/ubuntu:latest --command -- /usr/bin/tail '-f'
```

Now that it's running the background, I can use `oc rsh` to basically shell into the pod.
```
root@ubuntu:/# curl articles.demo-app-flynshue:8080 
Articles API Homepage

```

With this version of the articles app, it can display an additional message on the homepage using an environment variable `GREETING`.

We can create a configmap with these values and reference them in the deployment.

```bash
$ oc create configmap articles-env --from-literal=greeting=HelloWorld 
configmap/articles-env created

$ oc get cm articles-env -o yaml
apiVersion: v1
data:
  greeting: HelloWorld
kind: ConfigMap
```

Now let's update the deployment to reference this configmap.
```bash
vim deployment.yaml
...
        readinessProbe:
          httpGet:
            path: /healthz
            port: 5000
        env:
          - name: GREETING
            valueFrom:
              configMapKeyRef:
                name: articles-env
                key: greeting
...

oc apply -f deployment.yaml
```

Now rsh into your ubuntu pod and hit the articles home page
```
root@ubuntu:/# curl articles.demo-app-flynshue:8080
Articles API Homepage
HelloWorld
```

## Next Steps
This was just scratching the surface on what k8s can do.

Please checkout [Kubernetes Tutorials](https://kubernetes.io/docs/tutorials/) for more tutorials.