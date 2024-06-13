package main

import (
	"flag"
	"fmt"
	"time"

	sampleclient "crds-controller/pkg/generated/clientset/versioned"
	sampleinformer "crds-controller/pkg/generated/informers/externalversions"

	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

func main() {
	kubeconfig := flag.String("kubeconfig", "config", "location to kubeconfig file")
	fmt.Println(kubeconfig)

	config, err := clientcmd.BuildConfigFromFlags("", *kubeconfig)
	if err != nil {
		fmt.Printf("error %s building config\n", err.Error())

	}

	kubeClient, err := kubernetes.NewForConfig(config)
	if err != nil {
		fmt.Printf("error %s, creating clientset\n", err.Error())
	}

	crdClient, err := sampleclient.NewForConfig(config)
	if err != nil {
		fmt.Printf("error %s, creating sample clientset\n", err.Error())
	}

	ch := make(chan struct{})

	kubeInformer := informers.NewSharedInformerFactory(kubeClient, 10*time.Minute)
	crdInformer := sampleinformer.NewSharedInformerFactory(crdClient, 10*time.Minute)

	c := newController(kubeClient, crdClient, kubeInformer.Apps().V1().Deployments(), crdInformer.Sample().V1alpha1().Foos())
	if err != nil {
		fmt.Printf("getting informer factory %s\n", err.Error())
	}

	kubeInformer.Start(ch)
	crdInformer.Start(ch)
	c.run(ch)

}
