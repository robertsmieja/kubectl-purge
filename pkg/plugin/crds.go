package plugin

import (
	"fmt"
	"github.com/pkg/errors"
	apixv1client "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset/typed/apiextensions/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/fields"
	"sync"
)

func deleteClusterCrds(apixClient *apixv1client.ApiextensionsV1Client, errorCh chan<- error) {
	ctx, cancel := createCtx()
	waitGroup := sync.WaitGroup{}

	crds, err := apixClient.CustomResourceDefinitions().List(ctx, metav1.ListOptions{})
	if err != nil {
		errorCh <- errors.Wrap(err, "failed to list cluster-wide crds")
	}
	for _, crd := range crds.Items {
		waitGroup.Add(1)

		name := crd.Name
		go func() {
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

func deleteNamespaceCrds(apixClient *apixv1client.ApiextensionsV1Client, namespace string, errorCh chan<- error) {
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
		go func() {
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
