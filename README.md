# NetworkPolicyController


## How to use

```
$ ./autonetpol -h
Usage of ./autonetpol:
  --kubeconfig string
   absolute path to the kubeconfig file
  --master string
   master url
  --namespaces
   comma separated list of namespaces to watch/process 
```

Example output
```
$ ./autonetpol --master=https://10.2.2.2:8443 --kubeconfig=/root/.kube/config --namespaces=acme-air,netpolicy-test
Starting NetworkPolicyController controller
Waiting for pods
Starting NetworkPolicyController controller for namespace:  acme-air
Starting NetworkPolicyController controller for namespace:  netpolicy-test
Updating NetworkPolicy for:  acme-air-acme-web-c6bbf95d5-wlrwf
Updating NetworkPolicy for:  netpolicy-test-acme-web-5c678788b7-pfjmq
Updating NetworkPolicy for:  netpolicy-test-nginx
Updating NetworkPolicy for:  acme-air-mongodb-0
Updating NetworkPolicy for:  netpolicy-test-mongodb-0
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

Test that pod with multiple-ports will get networkpolicy for the ports.

``` 
cat <<EOF | kubectl apply -f -
apiVersion: v1
kind: Pod
metadata:
  name:	nginx-multiple-ports
  labels:
    app: nginx-multple-ports
spec:
  containers:
  - image: nginx
    name: nginx
    ports:
    - containerPort: 80
      name: http
      protocol: TCP
    - containerPort: 443
      name: https
      protocol: TCP
EOF
```

Test a pod with multiple containers each with their own port.

``` 
cat <<EOF | kubectl apply -f -
kind: Pod
metadata:
  name: mc
  labels:
    app: mc
spec:
  containers:
  - name: webapp
    image: gcr.io/google-samples/kubernetes-bootcamp:v1
    ports:
    - containerPort: 8080
  - name: nginx
    image: nginx:alpine
    ports:
    - containerPort: 80
EOF
```

