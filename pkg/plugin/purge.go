package plugin

import (
	"github.com/pkg/errors"
	"github.com/robertsmieja/kubectl-purge/pkg/logger"
	"github.com/robertsmieja/kubectl-purge/pkg/util"
	"golang.org/x/net/context"
	apixv1client "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset/typed/apiextensions/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/cli-runtime/pkg/genericclioptions"
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
	//return context.WithTimeout(context.Background(), 5*time.Second)
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

	apixClient, err := apixv1client.NewForConfig(config)
	if err != nil {
		return errors.Wrap(err, "failed to create apiextensions client")
	}

	namespaces, err := clientset.CoreV1().Namespaces().List(ctx, metav1.ListOptions{})
	if err != nil {
		return errors.Wrap(err, "failed to list namespaces")
	}

	// wait for all the goroutines per cluster
	clusterWaitGroup := sync.WaitGroup{}

	// wait for all the goroutines per namespace
	namespaceWaitGroup := sync.WaitGroup{}

	log.Info("Deleting cluster CRDs")
	clusterWaitGroup.Add(1)
	go func() {
		deleteClusterCrds(apixClient, errorCh)
		clusterWaitGroup.Done()
	}()

	clusterWaitGroup.Add(1)
	go func() {
		deleteClusterRoleBindings(clientset, errorCh)
		clusterWaitGroup.Done()
	}()

	clusterWaitGroup.Add(1)
	go func() {
		deleteClusterRoles(clientset, errorCh)
		clusterWaitGroup.Done()
	}()

	clusterWaitGroup.Add(1)
	go func() {
		deletePodSecurityPolicies(clientset, errorCh)
		clusterWaitGroup.Done()
	}()

	clusterWaitGroup.Add(1)
	go func() {
		deleteIngressClasses(clientset, errorCh)
		clusterWaitGroup.Done()
	}()

	for _, namespace := range namespaces.Items {
		namespaceName := namespace.Name

		if util.Contains(systemNamespaces, namespaceName) {
			log.Info("Skipping system namespace: %s", namespaceName)
			continue
		}

		log.Info("Deleting namespace: %s", namespaceName)

		namespaceWaitGroup.Add(1)
		go func() {
			deleteNamespaceCrds(apixClient, namespaceName, errorCh)
			namespaceWaitGroup.Done()
		}()

		namespaceWaitGroup.Add(1)
		go func() {
			deletePersistentVolumeClaims(clientset, namespaceName, errorCh)
			namespaceWaitGroup.Done()
		}()

		namespaceWaitGroup.Add(1)
		go func() {
			deleteConfigMaps(clientset, namespaceName, errorCh)
			namespaceWaitGroup.Done()
		}()

		namespaceWaitGroup.Add(1)
		go func() {
			deleteEndpoints(clientset, namespaceName, errorCh)
			namespaceWaitGroup.Done()
		}()

		namespaceWaitGroup.Add(1)
		go func() {
			// RoleBindings should be deleted BEFORE Roles
			deleteRoleBindings(clientset, namespaceName, errorCh)
			deleteRoles(clientset, namespaceName, errorCh)
			namespaceWaitGroup.Done()
		}()

		namespaceWaitGroup.Add(1)
		go func() {
			deleteIngresses(clientset, namespaceName, errorCh)
			namespaceWaitGroup.Done()
		}()

		namespaceWaitGroup.Add(1)
		go func() {
			deleteNetworkPolicies(clientset, namespaceName, errorCh)
			namespaceWaitGroup.Done()
		}()

		namespaceWaitGroup.Add(1)
		go func() {
			// CronJobs may kick off Jobs, they should go 1st
			deleteCronJobs(clientset, namespaceName, errorCh)
			deleteJobs(clientset, namespaceName, errorCh)
			namespaceWaitGroup.Done()
		}()

		namespaceWaitGroup.Add(1)
		go func() {
			deleteDeployments(clientset, namespaceName, errorCh)
			namespaceWaitGroup.Done()
		}()

		namespaceWaitGroup.Add(1)
		go func() {
			deleteDaemonSets(clientset, namespaceName, errorCh)
			namespaceWaitGroup.Done()
		}()

		namespaceWaitGroup.Add(1)
		go func() {
			deleteStatefulSets(clientset, namespaceName, errorCh)
			namespaceWaitGroup.Done()
		}()

		namespaceWaitGroup.Add(1)
		go func() {
			deleteReplicaSets(clientset, namespaceName, errorCh)
			namespaceWaitGroup.Done()
		}()

		namespaceWaitGroup.Add(1)
		go func() {
			deleteSecrets(clientset, namespaceName, errorCh)
			namespaceWaitGroup.Done()
		}()

		namespaceWaitGroup.Add(1)
		go func() {
			deletePodDisruptionBudgets(clientset, namespaceName, errorCh)
			namespaceWaitGroup.Done()
		}()

		namespaceWaitGroup.Add(1)
		go func() {
			deleteEvents(clientset, namespaceName, errorCh)
			namespaceWaitGroup.Done()
		}()

		// cleanup the namespace after everything is done
		clusterWaitGroup.Add(1)
		go func() {
			namespaceWaitGroup.Wait()
			if err := clientset.CoreV1().Namespaces().Delete(ctx, namespaceName, deletePolicy); err != nil {
				errorCh <- err
			}
			clusterWaitGroup.Done()
		}()
	}

	// delete PersistentVolumes after the namespaced PersistentVolumeClaims are deleted
	clusterWaitGroup.Add(1)
	go func() {
		deletePersistentVolumes(clientset, errorCh)
		clusterWaitGroup.Done()
	}()

	defer cancel()
	clusterWaitGroup.Wait()
	return nil
}
