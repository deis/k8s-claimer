package handlers

import (
	"fmt"

	"github.com/deis/k8s-claimer/k8s"
	"k8s.io/client-go/1.4/pkg/api"
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

// deleteNamespaces deletes all namespaces listed under all labels in namespaces.List, except for
// the namespaces in skip or skipDeleteNamespaces
func deleteNamespaces(namespaces k8s.NamespaceListerDeleter, skip map[string]struct{}) error {
	namespacesList, err := namespaces.List(api.ListOptions{})
	if err != nil {
		return errListNamespaces{origErr: err}
	}
	// TODO: delete concurrently https://github.com/deis/k8s-claimer/issues/49
	var errs []error
	for _, namespace := range namespacesList.Items {
		_, inSkip := skip[namespace.Name]
		_, isDefault := skipDeleteNamespaces[namespace.Name]
		if !isDefault && !inSkip {
			if err := namespaces.Delete(namespace.Name, &api.DeleteOptions{}); err != nil {
				errs = append(errs, err)
			}
		}
	}
	if len(errs) > 0 {
		return errDeleteNamespaces{origErrs: errs}
	}
	return nil
}
