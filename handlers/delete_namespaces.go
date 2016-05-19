package handlers

import (
	"fmt"

	"github.com/deis/k8s-claimer/k8s"
	k8sapi "k8s.io/kubernetes/pkg/api"
	labels "k8s.io/kubernetes/pkg/labels"
)

type errListNamespaces struct {
	origErr error
}

func (e errListNamespaces) Error() string {
	return fmt.Sprintf("error listing namespaces (%s)", e.origErr)
}

type errDeleteNamespaces struct {
	origErrs []error
}

func (e errDeleteNamespaces) Error() string {
	return fmt.Sprintf("error deleting namespaces (%+v)", e.origErrs)
}

func deleteNamespaces(namespaces k8s.NamespaceListerDeleter, skip map[string]struct{}) error {
	namespacesList, err := namespaces.List(k8sapi.ListOptions{LabelSelector: labels.Everything()})
	if err != nil {
		return errListNamespaces{origErr: err}
	}
	// TODO: delete concurrently https://github.com/deis/k8s-claimer/issues/49
	var errs []error
	for _, namespace := range namespacesList.Items {
		if namespace.Name != "kube-system" && namespace.Name != "default" {
			if err := namespaces.Delete(namespace.Name); err != nil {
				errs = append(errs, err)
			}
		}
	}
	if len(errs) > 0 {
		return errDeleteNamespaces{origErrs: errs}
	}
	return nil
}
