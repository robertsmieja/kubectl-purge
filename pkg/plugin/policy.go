package plugin

import (
	"fmt"
	"github.com/pkg/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"sync"
)

func deletePodSecurityPolicies(clientset *kubernetes.Clientset, errorCh chan<- error) {
	ctx, cancel := createCtx()
	waitGroup := sync.WaitGroup{}

	// TODO remove after K8s 1.22+, as this is deprecated
	api := clientset.PolicyV1beta1().PodSecurityPolicies()

	podSecurityPolicies, err := api.List(ctx, metav1.ListOptions{})
	if err != nil {
		errorCh <- errors.Wrap(err, "failed to list podSecurityPolicies")
	}
	for _, podSecurityPolicy := range podSecurityPolicies.Items {
		waitGroup.Add(1)

		name := podSecurityPolicy.Name
		go func() {
			err := api.Delete(ctx, name, deletePolicy)
			if err != nil {
				errorCh <- errors.Wrap(err, fmt.Sprintf("failed to delete podSecurityPolicy %s", name))
			}
			waitGroup.Done()
		}()
	}
	defer cancel()
	waitGroup.Wait()
}

func deletePodDisruptionBudgets(clientset *kubernetes.Clientset, namespace string, errorCh chan<- error) {
	ctx, cancel := createCtx()
	waitGroup := sync.WaitGroup{}

	api := clientset.PolicyV1().PodDisruptionBudgets(namespace)

	podDisruptionBudgets, err := api.List(ctx, metav1.ListOptions{})
	if err != nil {
		errorCh <- errors.Wrap(err, "failed to list podDisruptionBudgets")
	}
	for _, podDisruptionBudget := range podDisruptionBudgets.Items {
		waitGroup.Add(1)

		name := podDisruptionBudget.Name
		go func() {
			err := api.Delete(ctx, name, deletePolicy)
			if err != nil {
				errorCh <- errors.Wrap(err, fmt.Sprintf("failed to delete podDisruptionBudget %s", name))
			}
			waitGroup.Done()
		}()
	}
	defer cancel()
	waitGroup.Wait()
}
