package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/workqueue"
	"k8s.io/klog"
	"strings"
	"time"

	netv1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type Controller struct {
	indexer  cache.Indexer
	queue    workqueue.RateLimitingInterface
	informer cache.Controller
	Client *kubernetes.Clientset
}

func NewController(queue workqueue.RateLimitingInterface, indexer cache.Indexer, informer cache.Controller, client *kubernetes.Clientset) *Controller {
	return &Controller{
		informer: informer,
		indexer:  indexer,
		queue:    queue,
		Client: client,
	}
}

func (c *Controller) processNextItem() bool {
	// Wait until there is a new item in the working queue
	key, quit := c.queue.Get()
	if quit {
		return false
	}
	// Tell the queue that we are done with processing this key. This unblocks the key for other workers
	// This allows safe parallel processing because two pods with the same key are never processed in
	// parallel.
	defer c.queue.Done(key)

	// Invoke the method containing the business logic
	err := c.syncToStdout(key.(string))
	// Handle the error if something went wrong during the execution of the business logic
	c.handleErr(err, key)
	return true
}

// syncToStdout is the business logic of the controller. In this controller it simply prints
// information about the pod to stdout. In case an error happened, it has to simply return the error.
// The retry logic should not be part of the business logic.
func (c *Controller) syncToStdout(key string) error {
	labelValues := strings.Split(key, "/")
	obj, exists, err := c.indexer.GetByKey(key)
	if err != nil {
		klog.Errorf("Fetching object with key %s from store failed with %v", key, err)
		return err
	}

	// Define network policy shell
	np := &netv1.NetworkPolicy{}
	np.Name = strings.Join(labelValues,"-")
	np.Namespace = labelValues[0]

	if !exists {
		// If pod doesn't exist, delete it.  Have to see if pod info actually is returned if it is deleted
		fmt.Printf("%s does not exist anymore ... deleting associated network policy\n", key)
		err = c.Client.NetworkingV1().NetworkPolicies(np.Namespace).Delete(np.Name, &metav1.DeleteOptions{})
		if err != nil {
			return err
		}
	} else {
		// Create pod object from obj
		pod := obj.(*v1.Pod)
		// Loop through containers and find ports
		var ports []netv1.NetworkPolicyPort
		for _,container := range pod.Spec.Containers {
			// Loop through ports
			for _,port := range container.Ports {
				portNum := intstr.FromInt(int(port.ContainerPort))
				npp := netv1.NetworkPolicyPort{
					Protocol: &port.Protocol,
					Port: &portNum,
				}
				ports = append(ports, npp)
			}
		}
		ingressRule := netv1.NetworkPolicyIngressRule{
			Ports: ports,
		}
		egressRule := netv1.NetworkPolicyEgressRule{
			Ports: ports,
		}
		np.Spec.Ingress = []netv1.NetworkPolicyIngressRule{ingressRule}
		np.Spec.Egress = []netv1.NetworkPolicyEgressRule{egressRule}
		np.Spec.PodSelector = metav1.LabelSelector{
			MatchLabels: map[string]string{
				"autoNetPolicy" : np.Name,
			},
		}
		// Decide to create or update by trying to get the existing
		_, err = c.Client.NetworkingV1().NetworkPolicies(np.Namespace).Update(np)
		if err != nil {
			fmt.Println("Creating NetworkPolicy for pod:", pod.Name)
			np, err = c.Client.NetworkingV1().NetworkPolicies(np.Namespace).Create(np)
		} else {
			fmt.Println("Updating NetworkPolicy for pod:", pod.Name)
			np, err = c.Client.NetworkingV1().NetworkPolicies(np.Namespace).Update(np)
		}
		if err != nil {
			return err
		}
		// Put label on the pod if it doesn't exist
		if _, ok := pod.Labels["autoNetPolicy"]; !ok {
			fmt.Println("Adding label to pod ", pod.Name)
			// Build label metadata
			newLabel := map[string]map[string]map[string]string{
				"metadata" : map[string]map[string]string{
					"labels" : map[string]string{
						"autoNetPolicy" : np.Name,
					},
				},
			}
			data, err := json.Marshal(newLabel)
			if err != nil {
				return err
			}
			// Add a label to the pod
			pod, err = c.Client.CoreV1().Pods(pod.Namespace).Patch(pod.Name, types.MergePatchType, data, "")
			if err != nil {
				return err
			}
		}

	}
	return nil
}

// handleErr checks if an error happened and makes sure we will retry later.
func (c *Controller) handleErr(err error, key interface{}) {
	if err == nil {
		// Forget about the #AddRateLimited history of the key on every successful synchronization.
		// This ensures that future processing of updates for this key is not delayed because of
		// an outdated error history.
		c.queue.Forget(key)
		return
	}

	// This controller retries 5 times if something goes wrong. After that, it stops trying.
	if c.queue.NumRequeues(key) < 5 {
		klog.Infof("Error syncing pod %v: %v", key, err)

		// Re-enqueue the key rate limited. Based on the rate limiter on the
		// queue and the re-enqueue history, the key will be processed later again.
		c.queue.AddRateLimited(key)
		return
	}

	c.queue.Forget(key)
	// Report to an external entity that, even after several retries, we could not successfully process this key
	runtime.HandleError(err)
	klog.Infof("Dropping pod %q out of the queue: %v", key, err)
}

func (c *Controller) Run(threadiness int, stopCh chan struct{}) {
	defer runtime.HandleCrash()

	// Let the workers stop when we are done
	defer c.queue.ShutDown()
	klog.Info("Starting Pod controller")

	go c.informer.Run(stopCh)

	// Wait for all involved caches to be synced, before processing items from the queue is started
	if !cache.WaitForCacheSync(stopCh, c.informer.HasSynced) {
		runtime.HandleError(fmt.Errorf("Timed out waiting for caches to sync"))
		return
	}

	for i := 0; i < threadiness; i++ {
		go wait.Until(c.runWorker, time.Second, stopCh)
	}

	<-stopCh
	klog.Info("Stopping Pod controller")
}

func (c *Controller) runWorker() {
	for c.processNextItem() {
	}
}

func main() {
	var kubeconfig string
	var master string

	flag.StringVar(&kubeconfig, "kubeconfig", "", "absolute path to the kubeconfig file")
	flag.StringVar(&master, "master", "", "master url")
	flag.Parse()

	// creates the connection
	config, err := clientcmd.BuildConfigFromFlags(master, kubeconfig)
	if err != nil {
		klog.Fatal(err)
	}

	// creates the clientset
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		klog.Fatal(err)
	}

	// create the pod watcher
	podListWatcher := cache.NewListWatchFromClient(clientset.CoreV1().RESTClient(), "pods", "netpolicy-test", fields.Everything())

	// create the workqueue
	queue := workqueue.NewRateLimitingQueue(workqueue.DefaultControllerRateLimiter())

	// Bind the workqueue to a cache with the help of an informer. This way we make sure that
	// whenever the cache is updated, the pod key is added to the workqueue.
	// Note that when we finally process the item from the workqueue, we might see a newer version
	// of the Pod than the version which was responsible for triggering the update.
	indexer, informer := cache.NewIndexerInformer(podListWatcher, &v1.Pod{}, 0, cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			key, err := cache.MetaNamespaceKeyFunc(obj)
			if err == nil {
				queue.Add(key)
			}
		},
		UpdateFunc: func(old interface{}, new interface{}) {
			key, err := cache.MetaNamespaceKeyFunc(new)
			if err == nil {
				queue.Add(key)
			}
		},
		DeleteFunc: func(obj interface{}) {
			// IndexerInformer uses a delta queue, therefore for deletes we have to use this
			// key function.
			key, err := cache.DeletionHandlingMetaNamespaceKeyFunc(obj)
			if err == nil {
				queue.Add(key)
			}
		},
	}, cache.Indexers{})

	controller := NewController(queue, indexer, informer, clientset)

	// We can now warm up the cache for initial synchronization.
	// Let's suppose that we knew about a pod "mypod" on our last run, therefore add it to the cache.
	// If this pod is not there anymore, the controller will be notified about the removal after the
	// cache has synchronized.
	// indexer.Add(&v1.Pod{
	// 	ObjectMeta: meta_v1.ObjectMeta{
	// 		Name:      "mypod",
	// 		Namespace: v1.NamespaceDefault,
	// 	},
	// })
	fmt.Printf("Starting NetworkPolicyController controller\n")
	// Now let's start the controller
	stop := make(chan struct{})
	defer close(stop)
	go controller.Run(1, stop)

	// Wait forever
	select {}
}
