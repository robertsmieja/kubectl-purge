package plugin

import (
	"fmt"
	"github.com/robertsmieja/kubectl-purge/pkg/logger"
	"golang.org/x/net/context"
	"sync"
	"time"

	"github.com/pkg/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/client-go/kubernetes"
)

var gracePeriodSeconds = int64(0)
var fgPolicy = metav1.DeletePropagationForeground
var deletePolicy = metav1.DeleteOptions{
	GracePeriodSeconds: &gracePeriodSeconds,
	PropagationPolicy:  &fgPolicy,
}

func createCtx() (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), 5*time.Second)
}

func RunPlugin(configFlags *genericclioptions.ConfigFlags, log *logger.Logger, errorCh chan<- error) error {
	ctx, cancel := createCtx()

	config, err := configFlags.ToRESTConfig()
	if err != nil {
		return errors.Wrap(err, "failed to read kubeconfig")
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return errors.Wrap(err, "failed to create clientset")
	}

	namespaces, err := clientset.CoreV1().Namespaces().List(ctx, metav1.ListOptions{})
	if err != nil {
		return errors.Wrap(err, "failed to list namespaces")
	}

	waitGroup := sync.WaitGroup{}
	for _, namespace := range namespaces.Items {
		namespaceName := namespace.Name
		log.Info("Deleting namespace: %s", namespaceName)

		waitGroup.Add(1)
		go func() {
			deleteDeployments(clientset, namespaceName, errorCh)
			waitGroup.Done()
		}()

		waitGroup.Add(1)
		go func() {
			deleteDaemonSets(clientset, namespaceName, errorCh)
			waitGroup.Done()
		}()
	}

	defer cancel()
	waitGroup.Wait()
	return nil
}

func deleteDeployments(clientset *kubernetes.Clientset, namespace string, errorCh chan<- error) {
	ctx, cancel := createCtx()
	waitGroup := sync.WaitGroup{}

	deployments, err := clientset.AppsV1().Deployments(namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		errorCh <- errors.Wrap(err, "failed to list deployments")
	}
	for _, deployment := range deployments.Items {
		waitGroup.Add(1)

		deploymentName := deployment.Name
		go func() {
			err := clientset.AppsV1().Deployments(namespace).Delete(ctx, deploymentName, deletePolicy)
			if err != nil {
				errorCh <- errors.Wrap(err, fmt.Sprintf("failed to delete deployment %s", deploymentName))
			}
			waitGroup.Done()
		}()
	}

	defer cancel()
	waitGroup.Wait()
}

func deleteDaemonSets(clientset *kubernetes.Clientset, namespace string, errorCh chan<- error) {
	ctx, cancel := createCtx()
	waitGroup := sync.WaitGroup{}

	daemonSets, err := clientset.AppsV1().DaemonSets(namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		errorCh <- errors.Wrap(err, "failed to list daemonSets")
	}
	for _, daemonSet := range daemonSets.Items {
		waitGroup.Add(1)

		daemonSetName := daemonSet.Name
		go func() {
			err := clientset.AppsV1().DaemonSets(namespace).Delete(ctx, daemonSetName, deletePolicy)
			if err != nil {
				errorCh <- errors.Wrap(err, fmt.Sprintf("failed to delete daemonSet %s", daemonSetName))
			}
			waitGroup.Done()
		}()
	}
	defer cancel()
	waitGroup.Wait()
}
