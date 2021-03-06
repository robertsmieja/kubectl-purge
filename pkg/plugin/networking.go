package plugin

import (
	"fmt"
	"github.com/pkg/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"sync"
)

func deleteIngresses(clientset *kubernetes.Clientset, namespace string, logCh chan<- string, errorCh chan<- error) {
	ctx, cancel := createCtx()
	defer cancel()
	waitGroup := sync.WaitGroup{}

	api := clientset.NetworkingV1().Ingresses(namespace)

	ingresses, err := api.List(ctx, metav1.ListOptions{})
	if err != nil {
		errorCh <- errors.Wrap(err, "failed to list ingresses")
		return
	}
	for _, ingress := range ingresses.Items {
		waitGroup.Add(1)

		name := ingress.Name
		go func() {
			defer waitGroup.Done()
			err := api.Delete(ctx, name, deletePolicy)
			if err != nil {
				errorCh <- errors.Wrap(err, fmt.Sprintf("failed to delete ingress %s", name))
			}
		}()
	}
	waitGroup.Wait()
}

func deleteNetworkPolicies(clientset *kubernetes.Clientset, namespace string, logCh chan<- string, errorCh chan<- error) {
	ctx, cancel := createCtx()
	defer cancel()
	waitGroup := sync.WaitGroup{}

	api := clientset.NetworkingV1().NetworkPolicies(namespace)

	networkPolicies, err := api.List(ctx, metav1.ListOptions{})
	if err != nil {
		errorCh <- errors.Wrap(err, "failed to list networkPolicies")
		return
	}
	for _, networkPolicy := range networkPolicies.Items {
		waitGroup.Add(1)

		name := networkPolicy.Name
		go func() {
			defer waitGroup.Done()
			err := api.Delete(ctx, name, deletePolicy)
			if err != nil {
				errorCh <- errors.Wrap(err, fmt.Sprintf("failed to delete networkPolicy %s", name))
			}
		}()
	}
	waitGroup.Wait()
}

// this should be fine, as there are no IngressClasses by default
func deleteIngressClasses(clientset *kubernetes.Clientset, logCh chan<- string, errorCh chan<- error) {
	ctx, cancel := createCtx()
	defer cancel()
	waitGroup := sync.WaitGroup{}

	api := clientset.NetworkingV1().IngressClasses()

	ingressClasses, err := api.List(ctx, metav1.ListOptions{})
	if err != nil {
		errorCh <- errors.Wrap(err, "failed to list ingressClasses")
		return
	}
	for _, ingressClass := range ingressClasses.Items {
		waitGroup.Add(1)

		name := ingressClass.Name
		go func() {
			defer waitGroup.Done()
			err := api.Delete(ctx, name, deletePolicy)
			if err != nil {
				errorCh <- errors.Wrap(err, fmt.Sprintf("failed to delete ingressClass %s", name))
			}
		}()
	}
	waitGroup.Wait()
}
