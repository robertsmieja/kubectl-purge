package plugin

import (
	"fmt"
	"github.com/pkg/errors"
	"github.com/robertsmieja/kubectl-purge/pkg/util"
	"golang.org/x/net/context"
	apixv1client "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset/typed/apiextensions/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"sync"
)

var gracePeriodSeconds = int64(0)
var fgPolicy = metav1.DeletePropagationForeground
var deletePolicy = metav1.DeleteOptions{
	GracePeriodSeconds: &gracePeriodSeconds,
	PropagationPolicy:  &fgPolicy,
}

var systemNamespaces = []string{"kube-public", "kube-node-lease", "kube-system"}

func createCtx() (context.Context, context.CancelFunc) {
	return context.WithCancel(context.Background())
}

func RunPlugin(configFlags *genericclioptions.ConfigFlags, logCh chan<- string, errorCh chan<- error) error {
	ctx, cancel := createCtx()
	defer cancel()

	config, err := configFlags.ToRESTConfig()
	if err != nil {
		return errors.Wrap(err, "failed to read kubeconfig")
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return errors.Wrap(err, "failed to create clientset")
	}

	apixClient, err := apixv1client.NewForConfig(config)
	if err != nil {
		return errors.Wrap(err, "failed to create apiextensions client")
	}

	dynamicClient, err := dynamic.NewForConfig(config)
	if err != nil {
		return errors.Wrap(err, "failed to create dynamic client")
	}

	namespaces, err := clientset.CoreV1().Namespaces().List(ctx, metav1.ListOptions{})
	if err != nil {
		return errors.Wrap(err, "failed to list namespaces")
	}

	// wait for all the goroutines per cluster
	clusterWaitGroup := sync.WaitGroup{}

	// wait for all the goroutines per namespace
	namespaceWaitGroup := sync.WaitGroup{}

	logCh <- "Deleting cluster CRDs"
	clusterWaitGroup.Add(1)
	go func() {
		deleteClusterCrds(apixClient, dynamicClient, errorCh)
		clusterWaitGroup.Done()
	}()

	clusterWaitGroup.Add(1)
	go func() {
		deleteClusterRoleBindings(clientset, logCh, errorCh)
		clusterWaitGroup.Done()
	}()

	clusterWaitGroup.Add(1)
	go func() {
		deleteClusterRoles(clientset, logCh, errorCh)
		clusterWaitGroup.Done()
	}()

	clusterWaitGroup.Add(1)
	go func() {
		deletePodSecurityPolicies(clientset, logCh, errorCh)
		clusterWaitGroup.Done()
	}()

	clusterWaitGroup.Add(1)
	go func() {
		deleteIngressClasses(clientset, logCh, errorCh)
		clusterWaitGroup.Done()
	}()

	for _, namespace := range namespaces.Items {
		namespaceName := namespace.Name

		if util.Contains(systemNamespaces, namespaceName) {
			logCh <- fmt.Sprintf("Skipping system namespace: %s", namespaceName)
			continue
		}

		logCh <- fmt.Sprintf("Deleting namespace: %s", namespaceName)

		namespaceWaitGroup.Add(1)
		go func() {
			deleteNamespacedCrds(apixClient, dynamicClient, namespaceName, errorCh)
			namespaceWaitGroup.Done()
		}()

		namespaceWaitGroup.Add(1)
		go func() {
			deletePersistentVolumeClaims(clientset, namespaceName, logCh, errorCh)
			namespaceWaitGroup.Done()
		}()

		namespaceWaitGroup.Add(1)
		go func() {
			deleteConfigMaps(clientset, namespaceName, logCh, errorCh)
			namespaceWaitGroup.Done()
		}()

		namespaceWaitGroup.Add(1)
		go func() {
			deleteEndpoints(clientset, namespaceName, logCh, errorCh)
			namespaceWaitGroup.Done()
		}()

		namespaceWaitGroup.Add(1)
		go func() {
			// RoleBindings should be deleted BEFORE Roles
			deleteRoleBindings(clientset, namespaceName, logCh, errorCh)
			deleteRoles(clientset, namespaceName, logCh, errorCh)
			namespaceWaitGroup.Done()
		}()

		namespaceWaitGroup.Add(1)
		go func() {
			deleteIngresses(clientset, namespaceName, logCh, errorCh)
			namespaceWaitGroup.Done()
		}()

		namespaceWaitGroup.Add(1)
		go func() {
			deleteNetworkPolicies(clientset, namespaceName, logCh, errorCh)
			namespaceWaitGroup.Done()
		}()

		namespaceWaitGroup.Add(1)
		go func() {
			// CronJobs may kick off Jobs, they should go 1st
			deleteCronJobs(clientset, namespaceName, logCh, errorCh)
			deleteJobs(clientset, namespaceName, logCh, errorCh)
			namespaceWaitGroup.Done()
		}()

		namespaceWaitGroup.Add(1)
		go func() {
			deleteDeployments(clientset, namespaceName, logCh, errorCh)
			namespaceWaitGroup.Done()
		}()

		namespaceWaitGroup.Add(1)
		go func() {
			deleteDaemonSets(clientset, namespaceName, logCh, errorCh)
			namespaceWaitGroup.Done()
		}()

		namespaceWaitGroup.Add(1)
		go func() {
			deleteStatefulSets(clientset, namespaceName, logCh, errorCh)
			namespaceWaitGroup.Done()
		}()

		namespaceWaitGroup.Add(1)
		go func() {
			deleteReplicaSets(clientset, namespaceName, logCh, errorCh)
			namespaceWaitGroup.Done()
		}()

		namespaceWaitGroup.Add(1)
		go func() {
			deleteSecrets(clientset, namespaceName, logCh, errorCh)
			namespaceWaitGroup.Done()
		}()

		namespaceWaitGroup.Add(1)
		go func() {
			deletePodDisruptionBudgets(clientset, namespaceName, logCh, errorCh)
			namespaceWaitGroup.Done()
		}()

		namespaceWaitGroup.Add(1)
		go func() {
			deleteEvents(clientset, namespaceName, logCh, errorCh)
			namespaceWaitGroup.Done()
		}()

		// cleanup the namespace after everything is done
		if namespaceName != "default" {
			clusterWaitGroup.Add(1)
			go func() {
				namespaceWaitGroup.Wait()
				if err := clientset.CoreV1().Namespaces().Delete(ctx, namespaceName, deletePolicy); err != nil {
					errorCh <- errors.Wrap(err, fmt.Sprintf("failed to delete namespace: %s", namespaceName))
				}
				clusterWaitGroup.Done()
			}()
		}
	}

	// delete PersistentVolumes after the namespaced PersistentVolumeClaims are deleted
	clusterWaitGroup.Add(1)
	go func() {
		deletePersistentVolumes(clientset, logCh, errorCh)
		clusterWaitGroup.Done()
	}()

	clusterWaitGroup.Wait()
	close(logCh)
	close(errorCh)
	return nil
}
