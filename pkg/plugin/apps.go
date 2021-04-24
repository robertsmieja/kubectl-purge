package plugin

import (
	"fmt"
	"github.com/pkg/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"sync"
)

func deleteDeployments(clientset *kubernetes.Clientset, namespace string, logCh chan<- string, errorCh chan<- error) {
	ctx, cancel := createCtx()
	defer cancel()
	waitGroup := sync.WaitGroup{}

	api := clientset.AppsV1().Deployments(namespace)
	deployments, err := api.List(ctx, metav1.ListOptions{})
	if err != nil {
		errorCh <- errors.Wrap(err, "failed to list deployments")
	}
	for _, deployment := range deployments.Items {
		waitGroup.Add(1)

		deploymentName := deployment.Name
		go func() {
			if err := api.Delete(ctx, deploymentName, deletePolicy); err != nil {
				errorCh <- errors.Wrap(err, fmt.Sprintf("failed to delete deployment %s", deploymentName))
			}
			waitGroup.Done()
		}()
	}

	waitGroup.Wait()
}

func deleteDaemonSets(clientset *kubernetes.Clientset, namespace string, logCh chan<- string, errorCh chan<- error) {
	ctx, cancel := createCtx()
	defer cancel()
	waitGroup := sync.WaitGroup{}

	api := clientset.AppsV1().DaemonSets(namespace)
	daemonSets, err := api.List(ctx, metav1.ListOptions{})
	if err != nil {
		errorCh <- errors.Wrap(err, "failed to list daemonSets")
	}
	for _, daemonSet := range daemonSets.Items {
		waitGroup.Add(1)

		daemonSetName := daemonSet.Name
		go func() {
			err := api.Delete(ctx, daemonSetName, deletePolicy)
			if err != nil {
				errorCh <- errors.Wrap(err, fmt.Sprintf("failed to delete daemonSet %s", daemonSetName))
			}
			waitGroup.Done()
		}()
	}
	waitGroup.Wait()
}

func deleteStatefulSets(clientset *kubernetes.Clientset, namespace string, logCh chan<- string, errorCh chan<- error) {
	ctx, cancel := createCtx()
	defer cancel()
	waitGroup := sync.WaitGroup{}

	api := clientset.AppsV1().StatefulSets(namespace)

	statefulSets, err := api.List(ctx, metav1.ListOptions{})
	if err != nil {
		errorCh <- errors.Wrap(err, "failed to list statefulSets")
	}
	for _, statefulSet := range statefulSets.Items {
		waitGroup.Add(1)

		name := statefulSet.Name
		go func() {
			err := api.Delete(ctx, name, deletePolicy)
			if err != nil {
				errorCh <- errors.Wrap(err, fmt.Sprintf("failed to delete statefulSet %s", name))
			}
			waitGroup.Done()
		}()
	}
	waitGroup.Wait()
}

func deleteReplicaSets(clientset *kubernetes.Clientset, namespace string, logCh chan<- string, errorCh chan<- error) {
	ctx, cancel := createCtx()
	defer cancel()
	waitGroup := sync.WaitGroup{}

	api := clientset.AppsV1().ReplicaSets(namespace)

	replicaSets, err := api.List(ctx, metav1.ListOptions{})
	if err != nil {
		errorCh <- errors.Wrap(err, "failed to list replicaSets")
	}
	for _, replicaSet := range replicaSets.Items {
		waitGroup.Add(1)

		name := replicaSet.Name
		go func() {
			err := api.Delete(ctx, name, deletePolicy)
			if err != nil {
				errorCh <- errors.Wrap(err, fmt.Sprintf("failed to delete replicaSet %s", name))
			}
			waitGroup.Done()
		}()
	}
	waitGroup.Wait()
}
