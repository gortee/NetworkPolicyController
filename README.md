# NetworkPolicyController


## How to use

```
$ ./NetworkPolicyController -h
Usage of ./NetworkPolicyController:
  -kubeconfig string
   absolute path to the kubeconfig file
  -master string
   master url
```

Example output
```
$ ./NetworkPolicyController  -kubeconfig ~/.kube/config
Starting NetworkPolicyController controller
Sync/Add/Update for Pod nsx-demo-7498cd4994-9jgc7
Sync/Add/Update for Pod nsx-demo-7498cd4994-9jgc7
Sync/Add/Update for Pod nsx-demo-7498cd4994-hqd8z
Sync/Add/Update for Pod nsx-demo-7498cd4994-jz7gv
Sync/Add/Update for Pod nsx-demo-7498cd4994-hqd8z
Sync/Add/Update for Pod nsx-demo-7498cd4994-jz7gv
Sync/Add/Update for Pod nsx-demo-7498cd4994-hqd8z
Sync/Add/Update for Pod nsx-demo-7498cd4994-jz7gv
Sync/Add/Update for Pod nsx-demo-7498cd4994-9jgc7
Sync/Add/Update for Pod nsx-demo-7498cd4994-jz7gv
Sync/Add/Update for Pod nsx-demo-7498cd4994-hqd8z
Sync/Add/Update for Pod nsx-demo-7498cd4994-9jgc7
```

## How to deploy on K8s

## MVP
*Controller that listens for create/update/delete of pod on any namespace

*Creation/update/delete of a 1:1 network policy for each POD

*Network Policy will implement a ingress control on the listening port (deny all on everything else)

*Network Policy will allow all exgress traffic from the pod


## Test Method
Simple Test method:

```
kubectl run bootcamp --image=gcr.io/google-samples/kubernetes-bootcamp:v1 --port=8080
```


