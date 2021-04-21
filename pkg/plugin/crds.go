package plugin

import (
	"fmt"
	"github.com/pkg/errors"
	apixv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	apixv1client "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset/typed/apiextensions/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/fields"
	dynamic "k8s.io/client-go/dynamic"
	"sync"
)

func deleteClusterCrds(apixClient *apixv1client.ApiextensionsV1Client, dynamicClient dynamic.Interface, errorCh chan<- error) {
	ctx, cancel := createCtx()
	waitGroup := sync.WaitGroup{}

	crds, err := apixClient.CustomResourceDefinitions().List(ctx, metav1.ListOptions{})
	if err != nil {
		errorCh <- errors.Wrap(err, "failed to list cluster-wide crds")
	}
	for _, crd := range crds.Items {
		waitGroup.Add(1)

		name := crd.Name
		crd := crd
		go func() {
			deleteCustomResources(&dynamicClient, crd, "", errorCh)
			err := apixClient.CustomResourceDefinitions().Delete(ctx, name, deletePolicy)
			if err != nil {
				errorCh <- errors.Wrap(err, fmt.Sprintf("failed to delete crd %s", name))
			}
			waitGroup.Done()
		}()
	}
	defer cancel()
	waitGroup.Wait()
}

func deleteNamespacedCrds(apixClient *apixv1client.ApiextensionsV1Client, dynamicClient dynamic.Interface, namespace string, errorCh chan<- error) {
	ctx, cancel := createCtx()
	waitGroup := sync.WaitGroup{}

	crds, err := apixClient.CustomResourceDefinitions().List(ctx, metav1.ListOptions{
		FieldSelector: fields.SelectorFromSet(fields.Set{"metadata.namespace": namespace}).String(),
	})
	if err != nil {
		errorCh <- errors.Wrap(err, fmt.Sprintf("failed to list namespaced crds in: %s", namespace))
	}
	for _, crd := range crds.Items {
		waitGroup.Add(1)
		name := crd.Name
		crd := crd

		go func() {
			deleteCustomResources(&dynamicClient, crd, namespace, errorCh)

			err = apixClient.CustomResourceDefinitions().Delete(ctx, name, deletePolicy)
			if err != nil {
				errorCh <- errors.Wrap(err, fmt.Sprintf("failed to delete crd %s", name))
			}
			waitGroup.Done()
		}()
	}
	defer cancel()
	waitGroup.Wait()
}

func deleteCustomResources(dynamicClient *dynamic.Interface, crd apixv1.CustomResourceDefinition, namespace string, errorCh chan<- error) {
	name := crd.Name
	ctx, _ := createCtx()

	crdGvr := crd.GroupVersionKind().GroupVersion().WithResource(name)
	crApi := (*dynamicClient).Resource(crdGvr)
	crWaitGroup := sync.WaitGroup{}

	customResources, err := crApi.List(ctx, metav1.ListOptions{})
	if err != nil {
		var errorMsg string
		if namespace != "" {
			errorMsg = fmt.Sprintf("failed to list CustomResources for crd: %s in namespace: %s", name, namespace)
		} else {
			errorMsg = fmt.Sprintf("failed to list CustomResources for crd: %s", name)
		}
		errorCh <- errors.Wrap(err, errorMsg)
	}

	for _, customResource := range customResources.Items {
		crWaitGroup.Add(1)
		customResource := customResource
		go func() {
			err := crApi.Delete(ctx, customResource.GetName(), deletePolicy)
			if err != nil {
				errorCh <- errors.Wrap(err, fmt.Sprintf("failed to delete CustomResource %s for crd: %s in namespace: %s", customResource.GetName(), name, namespace))
			}
			crWaitGroup.Done()
		}()
	}
	crWaitGroup.Wait()
}
