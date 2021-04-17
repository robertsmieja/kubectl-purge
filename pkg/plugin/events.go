package plugin

import (
	"fmt"
	"github.com/pkg/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"sync"
)

func deleteEvents(clientset *kubernetes.Clientset, namespace string, logCh chan<- string, errorCh chan<- error) {
	ctx, cancel := createCtx()
	waitGroup := sync.WaitGroup{}

	api := clientset.EventsV1().Events(namespace)

	events, err := api.List(ctx, metav1.ListOptions{})
	if err != nil {
		errorCh <- errors.Wrap(err, "failed to list events")
	}
	for _, event := range events.Items {
		waitGroup.Add(1)

		name := event.Name
		go func() {
			err := api.Delete(ctx, name, deletePolicy)
			if err != nil {
				errorCh <- errors.Wrap(err, fmt.Sprintf("failed to delete event %s", name))
			}
			waitGroup.Done()
		}()
	}
	defer cancel()
	waitGroup.Wait()
}
