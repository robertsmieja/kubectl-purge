package plugin

import (
	"fmt"
	"github.com/pkg/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"sync"
)

func deleteRoles(clientset *kubernetes.Clientset, namespace string, errorCh chan<- error) {
	ctx, cancel := createCtx()
	waitGroup := sync.WaitGroup{}

	api := clientset.RbacV1().Roles(namespace)

	roles, err := api.List(ctx, metav1.ListOptions{})
	if err != nil {
		errorCh <- errors.Wrap(err, "failed to list roles")
	}
	for _, role := range roles.Items {
		waitGroup.Add(1)

		name := role.Name
		go func() {
			err := api.Delete(ctx, name, deletePolicy)
			if err != nil {
				errorCh <- errors.Wrap(err, fmt.Sprintf("failed to delete role %s", name))
			}
			waitGroup.Done()
		}()
	}
	defer cancel()
	waitGroup.Wait()
}

func deleteRoleBindings(clientset *kubernetes.Clientset, namespace string, errorCh chan<- error) {
	ctx, cancel := createCtx()
	waitGroup := sync.WaitGroup{}

	api := clientset.RbacV1().RoleBindings(namespace)

	roleBindings, err := api.List(ctx, metav1.ListOptions{})
	if err != nil {
		errorCh <- errors.Wrap(err, "failed to list roleBindings")
	}
	for _, roleBinding := range roleBindings.Items {
		waitGroup.Add(1)

		name := roleBinding.Name
		go func() {
			err := api.Delete(ctx, name, deletePolicy)
			if err != nil {
				errorCh <- errors.Wrap(err, fmt.Sprintf("failed to delete roleBinding %s", name))
			}
			waitGroup.Done()
		}()
	}
	defer cancel()
	waitGroup.Wait()
}

func deleteClusterRoles(clientset *kubernetes.Clientset, errorCh chan<- error) {
	ctx, cancel := createCtx()
	waitGroup := sync.WaitGroup{}

	api := clientset.RbacV1().ClusterRoles()

	clusterRoles, err := api.List(ctx, metav1.ListOptions{})
	if err != nil {
		errorCh <- errors.Wrap(err, "failed to list clusterRoles")
	}
	for _, clusterRole := range clusterRoles.Items {
		waitGroup.Add(1)

		name := clusterRole.Name
		go func() {
			err := api.Delete(ctx, name, deletePolicy)
			if err != nil {
				errorCh <- errors.Wrap(err, fmt.Sprintf("failed to delete clusterRole %s", name))
			}
			waitGroup.Done()
		}()
	}
	defer cancel()
	waitGroup.Wait()
}

func deleteClusterRoleBindings(clientset *kubernetes.Clientset, errorCh chan<- error) {
	ctx, cancel := createCtx()
	waitGroup := sync.WaitGroup{}

	api := clientset.RbacV1().ClusterRoleBindings()

	clusterRoleBindings, err := api.List(ctx, metav1.ListOptions{})
	if err != nil {
		errorCh <- errors.Wrap(err, "failed to list clusterRoleBindings")
	}
	for _, clusterRoleBinding := range clusterRoleBindings.Items {
		waitGroup.Add(1)

		name := clusterRoleBinding.Name
		go func() {
			err := api.Delete(ctx, name, deletePolicy)
			if err != nil {
				errorCh <- errors.Wrap(err, fmt.Sprintf("failed to delete clusterRoleBinding %s", name))
			}
			waitGroup.Done()
		}()
	}
	defer cancel()
	waitGroup.Wait()
}
