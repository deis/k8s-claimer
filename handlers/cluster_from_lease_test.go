package handlers

import (
	"testing"
	"time"

	"github.com/arschles/assert"
	"github.com/deis/k8s-claimer/gke"
	"github.com/deis/k8s-claimer/leases"
	container "google.golang.org/api/container/v1"
)

var (
	cluster1 = container.Cluster{Name: "cluster1"}
)

func TestGetClusterFromLease(t *testing.T) {
	clusterLister := gke.FakeClusterLister{
		Resp: &container.ListClustersResponse{
			Clusters: []*container.Cluster{&cluster1},
		},
	}
	lease := leases.NewLease(cluster1.Name, time.Now().Add(1*time.Hour))
	cluster, err := getClusterFromLease(lease, clusterLister, projID, zone)
	assert.NoErr(t, err)
	assert.Equal(t, cluster.Name, cluster1.Name, "cluster name")
}
