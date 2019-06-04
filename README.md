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

``` 
$ cd k8s/
$ kubectl apply -f ./
```

## MVP
-Controller that listens for create/update/delete of pod on any namespace

-Creation/delete of a 1:1 network policy for each POD

-Network Policy will implement a ingress control on the listening port (deny all on everything else)

-Network Policy will allow all egress traffic from the pod

-Create network policy label

## Flow of controller
POD Create

1. Generate name:namespace-podname key:autonetpol label assign to POD

2. Create network policy from the port name

3. Assign label in 1. to network policy from port name


POD Delete

1. Remove network policy from label name


## Test Method
Simple Test method:

```
kubectl run bootcamp --image=gcr.io/google-samples/kubernetes-bootcamp:v1 --port=8080
kubectl get pod bootcamp -o wide --show-labels
kubectl get netpol <netpol name> --show-labels
kubectl descrribe netpol <netpol name>
```

Test scale the deployment to ensure the pod based rules work on a scale operation:

    kubectl scale deployments/bootcamp --replicas=4


Test that rules do not break with a service of LoadBalancer


    kubectl expose deployment/bootcamp --type=LoadBalancer --port 8080


