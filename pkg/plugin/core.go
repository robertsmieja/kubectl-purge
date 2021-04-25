package plugin

import (
	"fmt"
	"github.com/pkg/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"sync"
)

func deleteConfigMaps(clientset *kubernetes.Clientset, namespace string, logCh chan<- string, errorCh chan<- error) {
	ctx, cancel := createCtx()
	defer cancel()
	waitGroup := sync.WaitGroup{}

	api := clientset.CoreV1().ConfigMaps(namespace)

	configMaps, err := api.List(ctx, metav1.ListOptions{})
	if err != nil {
		errorCh <- errors.Wrap(err, "failed to list configMaps")
		return
	}
	for _, configMap := range configMaps.Items {
		waitGroup.Add(1)

		name := configMap.Name
		go func() {
			defer waitGroup.Done()
			err := api.Delete(ctx, name, deletePolicy)
			if err != nil {
				errorCh <- errors.Wrap(err, fmt.Sprintf("failed to delete configMap %s", name))
			}
		}()
	}
	waitGroup.Wait()
}

func deleteEndpoints(clientset *kubernetes.Clientset, namespace string, logCh chan<- string, errorCh chan<- error) {
	ctx, cancel := createCtx()
	defer cancel()
	waitGroup := sync.WaitGroup{}

	api := clientset.CoreV1().Endpoints(namespace)

	endpoints, err := api.List(ctx, metav1.ListOptions{})
	if err != nil {
		errorCh <- errors.Wrap(err, "failed to list endpoints")
		return
	}
	for _, endpoint := range endpoints.Items {
		waitGroup.Add(1)

		name := endpoint.Name
		go func() {
			defer waitGroup.Done()
			err := api.Delete(ctx, name, deletePolicy)
			if err != nil {
				errorCh <- errors.Wrap(err, fmt.Sprintf("failed to delete endpoint %s", name))
			}
		}()
	}
	waitGroup.Wait()
}

func deletePersistentVolumeClaims(clientset *kubernetes.Clientset, namespace string, logCh chan<- string, errorCh chan<- error) {
	ctx, cancel := createCtx()
	defer cancel()
	waitGroup := sync.WaitGroup{}

	api := clientset.CoreV1().PersistentVolumeClaims(namespace)

	persistentVolumeClaims, err := api.List(ctx, metav1.ListOptions{})
	if err != nil {
		errorCh <- errors.Wrap(err, "failed to list persistentVolumeClaims")
		return
	}
	for _, persistentVolumeClaim := range persistentVolumeClaims.Items {
		waitGroup.Add(1)

		name := persistentVolumeClaim.Name
		go func() {
			defer waitGroup.Done()
			err := api.Delete(ctx, name, deletePolicy)
			if err != nil {
				errorCh <- errors.Wrap(err, fmt.Sprintf("failed to delete persistentVolumeClaim %s", name))
			}
		}()
	}
	waitGroup.Wait()
}

func deletePersistentVolumes(clientset *kubernetes.Clientset, logCh chan<- string, errorCh chan<- error) {
	ctx, cancel := createCtx()
	defer cancel()
	waitGroup := sync.WaitGroup{}

	api := clientset.CoreV1().PersistentVolumes()

	persistentVolumes, err := api.List(ctx, metav1.ListOptions{})
	if err != nil {
		errorCh <- errors.Wrap(err, "failed to list persistentVolumes")
		return
	}
	for _, persistentVolume := range persistentVolumes.Items {
		waitGroup.Add(1)

		name := persistentVolume.Name
		go func() {
			defer waitGroup.Done()
			err := api.Delete(ctx, name, deletePolicy)
			if err != nil {
				errorCh <- errors.Wrap(err, fmt.Sprintf("failed to delete persistentVolume %s", name))
			}
		}()
	}
	waitGroup.Wait()
}

func deleteSecrets(clientset *kubernetes.Clientset, namespace string, logCh chan<- string, errorCh chan<- error) {
	ctx, cancel := createCtx()
	defer cancel()
	waitGroup := sync.WaitGroup{}

	api := clientset.CoreV1().Secrets(namespace)

	secrets, err := api.List(ctx, metav1.ListOptions{})
	if err != nil {
		errorCh <- errors.Wrap(err, "failed to list secrets")
		return
	}
	for _, secret := range secrets.Items {
		waitGroup.Add(1)

		name := secret.Name
		go func() {
			defer waitGroup.Done()
			err := api.Delete(ctx, name, deletePolicy)
			if err != nil {
				errorCh <- errors.Wrap(err, fmt.Sprintf("failed to delete secret %s", name))
			}
		}()
	}
	waitGroup.Wait()
}
