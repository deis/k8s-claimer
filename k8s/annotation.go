package k8s

import (
	"github.com/deis/k8s-claimer/leases"
	"k8s.io/client-go/pkg/api/v1"
)

// SaveAnnotations will publish the current lease map back to the k8s annotation
func SaveAnnotations(services ServiceUpdater, svc *v1.Service, leaseMap *leases.Map) error {
	annos, err := leaseMap.ToAnnotations()
	if err != nil {
		return err
	}
	svc.Annotations = annos
	if _, err := services.Update(svc); err != nil {
		return err
	}
	return nil
}
