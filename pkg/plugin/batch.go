package plugin

import (
	"fmt"
	"github.com/pkg/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"sync"
)

func deleteCronJobs(clientset *kubernetes.Clientset, namespace string, logCh chan<- string, errorCh chan<- error) {
	ctx, cancel := createCtx()
	defer cancel()
	waitGroup := sync.WaitGroup{}

	api := clientset.BatchV1().CronJobs(namespace)

	cronJobs, err := api.List(ctx, metav1.ListOptions{})
	if err != nil {
		errorCh <- errors.Wrap(err, "failed to list cronJobs")
		return
	}
	for _, cronJob := range cronJobs.Items {
		waitGroup.Add(1)

		name := cronJob.Name
		go func() {
			defer waitGroup.Done()
			err := api.Delete(ctx, name, deletePolicy)
			if err != nil {
				errorCh <- errors.Wrap(err, fmt.Sprintf("failed to delete cronJob %s", name))
			}
		}()
	}
	waitGroup.Wait()
}

func deleteJobs(clientset *kubernetes.Clientset, namespace string, logCh chan<- string, errorCh chan<- error) {
	ctx, cancel := createCtx()
	defer cancel()
	waitGroup := sync.WaitGroup{}

	api := clientset.BatchV1().Jobs(namespace)

	jobs, err := api.List(ctx, metav1.ListOptions{})
	if err != nil {
		errorCh <- errors.Wrap(err, "failed to list jobs")
		return
	}
	for _, job := range jobs.Items {
		waitGroup.Add(1)

		name := job.Name
		go func() {
			defer waitGroup.Done()
			err := api.Delete(ctx, name, deletePolicy)
			if err != nil {
				errorCh <- errors.Wrap(err, fmt.Sprintf("failed to delete job %s", name))
			}
		}()
	}
	defer cancel()
	waitGroup.Wait()
}
