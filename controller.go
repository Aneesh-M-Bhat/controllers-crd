package main

import (
	"context"
	"fmt"
	"time"

	samplev1alpha1 "crds-controller/pkg/apis/sample/v1alpha1"
	sampleclient "crds-controller/pkg/generated/clientset/versioned"
	"crds-controller/pkg/generated/clientset/versioned/scheme"
	samplescheme "crds-controller/pkg/generated/clientset/versioned/scheme"
	sampleinformer "crds-controller/pkg/generated/informers/externalversions/sample/v1alpha1"
	samplelister "crds-controller/pkg/generated/listers/sample/v1alpha1"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"
	appsinformers "k8s.io/client-go/informers/apps/v1"
	"k8s.io/client-go/kubernetes"
	appslisters "k8s.io/client-go/listers/apps/v1"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/util/workqueue"
)

type controller struct {
	clientset    kubernetes.Interface
	crdclientset sampleclient.Interface

	depLister      appslisters.DeploymentLister
	depCacheSynced cache.InformerSynced
	foosLister     samplelister.FooLister
	fooCacheSynced cache.InformerSynced

	queue workqueue.RateLimitingInterface
}

func newController(clientset kubernetes.Interface, crdclientset sampleclient.Interface, depInformer appsinformers.DeploymentInformer, crdInformer sampleinformer.FooInformer) *controller {
	fmt.Println("Add to Scheme")
	samplescheme.AddToScheme(scheme.Scheme)

	c := &controller{
		crdclientset:   crdclientset,
		foosLister:     crdInformer.Lister(),
		fooCacheSynced: crdInformer.Informer().HasSynced,

		clientset:      clientset,
		depLister:      depInformer.Lister(),
		depCacheSynced: depInformer.Informer().HasSynced,

		queue: workqueue.NewNamedRateLimitingQueue(workqueue.DefaultControllerRateLimiter(), "crds-controller"),
	}

	crdInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc:    c.handleAdd,
		UpdateFunc: c.handleUpdate,
	})

	depInformer.Informer().AddEventHandler(
		cache.ResourceEventHandlerFuncs{
			AddFunc: c.handleDep,
			UpdateFunc: func(old, new interface{}) {
				newDep := new.(*appsv1.Deployment)
				oldDep := old.(*appsv1.Deployment)
				if newDep.ResourceVersion == oldDep.ResourceVersion {
					return
				}
				c.handleDep(new)
			},
			DeleteFunc: c.handleDep,
		},
	)

	return c
}

func (c *controller) run(ch <-chan struct{}) {
	fmt.Println("starting controller")
	if !cache.WaitForCacheSync(ch, c.depCacheSynced, c.fooCacheSynced) {
		fmt.Print("waiting for cache to by synced\n")
	}

	go wait.Until(c.worker, 1*time.Second, ch)

	<-ch
}

func (c *controller) worker() {
	for c.processItem() {

	}
}

func (c *controller) processItem() bool {
	fmt.Println("Process Item")
	item, shutdown := c.queue.Get()
	if shutdown {
		return false
	}
	var key string
	var ok bool
	var err error

	defer c.queue.Done(item)
	if key, ok = item.(string); !ok {
		c.queue.Forget(item)
		fmt.Printf("Expected string in queue")
		return true
	}

	if err = c.syncHandler(key); err != nil {
		c.queue.AddRateLimited(item)
		fmt.Printf("error syncing requeueing %s\n", err.Error())
		return true
	}
	c.queue.Forget(item)
	return true
}

func (c *controller) syncHandler(key string) error {
	fmt.Println("controller syncing")
	namespace, name, err := cache.SplitMetaNamespaceKey(key)
	if err != nil {
		fmt.Printf("Invalid resource key %s", err)
		return nil
	}

	foo, err := c.foosLister.Foos(namespace).Get(name)
	if err != nil {
		fmt.Printf("error in fetching foo : %s", err)
		if errors.IsNotFound(err) {
			return nil
		}
		return err
	}

	depName := foo.ObjectMeta.Name
	if depName == "" {
		fmt.Printf("Specify resource name")
		return nil
	}

	dep, err := c.depLister.Deployments(foo.Namespace).Get(depName)
	if errors.IsNotFound(err) {
		dep, err = c.clientset.AppsV1().Deployments(foo.Namespace).Create(context.TODO(), newDeployment(foo), metav1.CreateOptions{})
	}

	if err != nil {
		return err
	}

	if !metav1.IsControlledBy(dep, foo) {
		return fmt.Errorf("Resource exists but not connected")
	}

	fmt.Println("Counts & Replicas")
	fmt.Println(*foo.Spec.Count)
	fmt.Println(*dep.Spec.Replicas)
	if foo.Spec.Count != nil && *foo.Spec.Count != *dep.Spec.Replicas {
		_, err = c.clientset.AppsV1().Deployments(foo.Namespace).Update(context.TODO(), newDeployment(foo), metav1.UpdateOptions{})
	}

	if err != nil {
		return err
	}
	return nil
}

func depLabels(dep appsv1.Deployment) map[string]string {
	return dep.Spec.Template.Labels
}

func (c *controller) handleAdd(obj interface{}) {
	fmt.Println("Add")
	c.enqueue(obj)
}

func (c *controller) handleUpdate(old, obj interface{}) {
	fmt.Println("Update")
	c.enqueue(obj)
}

func (c *controller) handleDel(obj interface{}) {
	fmt.Println("Delete")
	c.enqueue(obj)
}

func (c *controller) enqueue(obj interface{}) {
	key, err := cache.MetaNamespaceKeyFunc(obj)
	if err != nil {
		fmt.Printf("getting key from cache %s\n", err.Error())
		return
	}
	c.queue.Add(key)
}

func (c *controller) handleDep(obj interface{}) {
	var object metav1.Object
	var ok bool

	if object, ok = obj.(metav1.Object); !ok {
		tombstone, ok := obj.(cache.DeletedFinalStateUnknown)
		if !ok {
			fmt.Println("error decoding object, invalid type")
			return
		}
		object, ok = tombstone.Obj.(metav1.Object)
		if !ok {
			fmt.Println("error decoding object tombstone, invalid type")
			return
		}
	}

	if ownerRef := metav1.GetControllerOf(object); ownerRef != nil {
		if ownerRef.Kind != "Foo" {
			panic("Kind is not equal to Foo")
		}

		foo, err := c.foosLister.Foos(object.GetNamespace()).Get(ownerRef.Name)
		if err != nil {
			panic(err)
		}

		c.enqueue(foo)
	}
}

func newDeployment(foo *samplev1alpha1.Foo) *appsv1.Deployment {
	fmt.Println("Controller: Creating new deployment")
	labels := map[string]string{
		"app":        "nginx",
		"controller": foo.Name,
	}
	return &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      foo.ObjectMeta.Name,
			Namespace: foo.Namespace,
			OwnerReferences: []metav1.OwnerReference{
				*metav1.NewControllerRef(foo, samplev1alpha1.SchemeGroupVersion.WithKind("Foo")),
			},
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: foo.Spec.Count,
			Selector: &metav1.LabelSelector{
				MatchLabels: labels,
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: labels,
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:    "nginx",
							Image:   "nginx:latest",
							Command: []string{"/bin/bash"},
							Args:    []string{"-c", "echo " + foo.Spec.Message + " && sleep 3600"},
						},
					},
				},
			},
		},
	}
}
