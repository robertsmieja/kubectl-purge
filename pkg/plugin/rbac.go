package plugin

import (
	"fmt"
	"github.com/pkg/errors"
	"github.com/robertsmieja/kubectl-purge/pkg/util"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"sync"
)

func deleteRoles(clientset *kubernetes.Clientset, namespace string, logCh chan<- string, errorCh chan<- error) {
	ctx, cancel := createCtx()
	defer cancel()
	waitGroup := sync.WaitGroup{}

	api := clientset.RbacV1().Roles(namespace)

	roles, err := api.List(ctx, metav1.ListOptions{})
	if err != nil {
		errorCh <- errors.Wrap(err, "failed to list roles")
		return
	}
	for _, role := range roles.Items {
		waitGroup.Add(1)

		name := role.Name
		go func() {
			defer waitGroup.Done()
			err := api.Delete(ctx, name, deletePolicy)
			if err != nil {
				errorCh <- errors.Wrap(err, fmt.Sprintf("failed to delete role %s", name))
			}
		}()
	}
	waitGroup.Wait()
}

func deleteRoleBindings(clientset *kubernetes.Clientset, namespace string, logCh chan<- string, errorCh chan<- error) {
	ctx, cancel := createCtx()
	defer cancel()
	waitGroup := sync.WaitGroup{}

	api := clientset.RbacV1().RoleBindings(namespace)

	roleBindings, err := api.List(ctx, metav1.ListOptions{})
	if err != nil {
		errorCh <- errors.Wrap(err, "failed to list roleBindings")
		return
	}
	for _, roleBinding := range roleBindings.Items {
		waitGroup.Add(1)

		name := roleBinding.Name
		go func() {
			defer waitGroup.Done()
			err := api.Delete(ctx, name, deletePolicy)
			if err != nil {
				errorCh <- errors.Wrap(err, fmt.Sprintf("failed to delete roleBinding %s", name))
			}
		}()
	}
	waitGroup.Wait()
}

var defaultClusterRoles = []string{
	"admin",
	"cluster-admin",
	"edit",
	"view",
	"vpnkit-controller", // docker desktop
}

var defaultClusterRolePrefixes = []string{
	"kubeadm:",
	"microk8s", // microk8s
	"system:",
}

func deleteClusterRoles(clientset *kubernetes.Clientset, logCh chan<- string, errorCh chan<- error) {
	ctx, cancel := createCtx()
	defer cancel()
	waitGroup := sync.WaitGroup{}

	api := clientset.RbacV1().ClusterRoles()

	clusterRoles, err := api.List(ctx, metav1.ListOptions{})
	if err != nil {
		errorCh <- errors.Wrap(err, "failed to list clusterRoles")
		return
	}
	for _, clusterRole := range clusterRoles.Items {
		name := clusterRole.Name

		if util.Contains(defaultClusterRoles, name) || util.StartsWithAny(defaultClusterRolePrefixes, name) {
			logCh <- fmt.Sprintf("Skipping ClusterRole: %s", name)
			continue
		}

		waitGroup.Add(1)
		go func() {
			defer waitGroup.Done()
			if err := api.Delete(ctx, name, deletePolicy); err != nil {
				errorCh <- errors.Wrap(err, fmt.Sprintf("failed to delete clusterRole %s", name))
			}
		}()
	}
	waitGroup.Wait()
}

var defaultClusterRoleBindings = []string{
	"cluster-admin",
	"docker-for-desktop-binding", // docker desktop
	"vpnkit-controller",          // docker desktop
}

var defaultClusterRoleBindingPrefixes = []string{
	"kubeadm:",
	"microk8s", // microk8s
	"system:",
}

func deleteClusterRoleBindings(clientset *kubernetes.Clientset, logCh chan<- string, errorCh chan<- error) {
	ctx, cancel := createCtx()
	defer cancel()
	waitGroup := sync.WaitGroup{}

	api := clientset.RbacV1().ClusterRoleBindings()

	clusterRoleBindings, err := api.List(ctx, metav1.ListOptions{})
	if err != nil {
		errorCh <- errors.Wrap(err, "failed to list clusterRoleBindings")
		return
	}
	for _, clusterRoleBinding := range clusterRoleBindings.Items {
		name := clusterRoleBinding.Name

		if util.Contains(defaultClusterRoleBindings, name) || util.StartsWithAny(defaultClusterRoleBindingPrefixes, name) {
			logCh <- fmt.Sprintf("Skipping ClusterRole: %s", name)
			continue
		}

		waitGroup.Add(1)
		go func() {
			defer waitGroup.Done()
			err := api.Delete(ctx, name, deletePolicy)
			if err != nil {
				errorCh <- errors.Wrap(err, fmt.Sprintf("failed to delete clusterRoleBinding %s", name))
			}
		}()
	}
	waitGroup.Wait()
}
