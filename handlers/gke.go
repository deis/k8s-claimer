package handlers

import (
	"errors"
	"strings"

	"github.com/deis/k8s-claimer/clusters"
	"github.com/deis/k8s-claimer/leases"
	container "google.golang.org/api/container/v1"
	yaml "gopkg.in/yaml.v2"
	k8scmd "k8s.io/kubernetes/pkg/client/unversioned/clientcmd/api"
)

const (
	kubeconfigAPIVersion = "v1"
)

var (
	errUnusedGKEClusterNotFound = errors.New("all GKE clusters are in use")
)

// findUnusedGKECluster finds a GKE cluster that's not currently in use according to the
// annotations in svc. returns errUnusedGKEClusterNotFound if none is found
func findUnusedGKECluster(clusterMap *clusters.Map, leaseMap *leases.Map) (*container.Cluster, error) {
	clusterNames := clusterMap.Names()
	for _, clusterName := range clusterNames {
		cluster, _ := clusterMap.ClusterByName(clusterName)
		_, found := leaseMap.LeaseByClusterName(clusterName)
		if !found {
			return cluster, nil
		}
	}
	return nil, errUnusedGKEClusterNotFound
}

func createKubeConfigFromCluster(cluster *container.Cluster) ([]byte, error) {
	contextName := strings.ToLower(cluster.Name)
	authInfoName := contextName
	clusters := map[string]*k8scmd.Cluster{
		cluster.Name: &k8scmd.Cluster{
			Server: cluster.Endpoint,
			CertificateAuthorityData: []byte(cluster.MasterAuth.ClusterCaCertificate),
		},
	}
	contexts := map[string]*k8scmd.Context{
		contextName: &k8scmd.Context{
			Cluster:  cluster.Name,
			AuthInfo: authInfoName,
		},
	}
	authInfos := map[string]*k8scmd.AuthInfo{
		authInfoName: &k8scmd.AuthInfo{
			ClientCertificateData: []byte(cluster.MasterAuth.ClientCertificate),
			ClientKeyData:         []byte(cluster.MasterAuth.ClientKey),
			Username:              cluster.MasterAuth.Username,
			Password:              cluster.MasterAuth.Password,
		},
	}
	cfg := &k8scmd.Config{
		CurrentContext: contextName,
		APIVersion:     kubeconfigAPIVersion,
		Clusters:       clusters,
		Contexts:       contexts,
		AuthInfos:      authInfos,
	}
	cfgBytes, err := yaml.Marshal(cfg)
	if err != nil {
		return nil, err
	}
	return cfgBytes, nil
}
