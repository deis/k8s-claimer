package handlers

import (
	"encoding/base64"
	"errors"
	"fmt"
	"math/rand"
	"regexp"
	"strings"

	"github.com/deis/k8s-claimer/clusters"
	"github.com/deis/k8s-claimer/k8s"
	"github.com/deis/k8s-claimer/leases"
	container "google.golang.org/api/container/v1"
	"gopkg.in/yaml.v2"
)

const (
	kubeconfigAPIVersion = "v1"
)

var (
	errUnusedGKEClusterNotFound = errors.New("all GKE clusters are in use")
)

// findUnusedGKECluster finds a GKE cluster that's not currently in use according to the
// annotations in svc. It will also attempt to match the clusterRegex passed in if possible.
// Returns errUnusedGKEClusterNotFound if none is found
func findUnusedGKECluster(clusterMap *clusters.Map, leaseMap *leases.Map, clusterRegex string, clusterVersion string) (*container.Cluster, error) {
	if clusterRegex != "" {
		return findUnusuedGKEClusterByName(clusterMap, leaseMap, clusterRegex)
	} else if clusterVersion != "" {
		return findUnusedGKEClusterByVersion(clusterMap, leaseMap, clusterVersion)
	} else {
		return findRandomUnusuedGKECluster(clusterMap, leaseMap)
	}
}

// findUnusuedGKEClusterByName attempts to find a unused GKE cluster that matches the regex passed in via the cli.
func findUnusuedGKEClusterByName(clusterMap *clusters.Map, leaseMap *leases.Map, clusterRegex string) (*container.Cluster, error) {
	regex, err := regexp.Compile(clusterRegex)
	if err != nil {
		return nil, err
	}
	for _, clusterName := range clusterMap.Names() {
		if regex.MatchString(clusterName) {
			cluster, err := checkLease(clusterMap, leaseMap, clusterName)
			if err == nil {
				return cluster, nil
			}
		}
	}
	return nil, errUnusedGKEClusterNotFound
}

// findUnusedGKEClusterByVersion attempts to find an unused GKE cluster that matches the version passed in via the CLI.
func findUnusedGKEClusterByVersion(clusterMap *clusters.Map, leaseMap *leases.Map, clusterVersion string) (*container.Cluster, error) {
	clusterNames := clusterMap.ClusterNamesByVersion(clusterVersion)
	if len(clusterNames) > 0 {
		for tries := 0; tries < 10; tries++ {
			clusterName := clusterNames[rand.Intn(len(clusterNames))]
			cluster, err := checkLease(clusterMap, leaseMap, clusterName)
			if err == nil {
				return cluster, nil
			}
		}
	}
	return nil, errUnusedGKEClusterNotFound
}

// findUnusuedGKECluster attempts to find a random unused GKE cluster
func findRandomUnusuedGKECluster(clusterMap *clusters.Map, leaseMap *leases.Map) (*container.Cluster, error) {
	clusterNames := clusterMap.Names()
	if len(clusterNames) > 0 {
		for tries := 0; tries < 10; tries++ {
			clusterName := clusterNames[rand.Intn(len(clusterNames))]
			cluster, err := checkLease(clusterMap, leaseMap, clusterName)
			if err == nil {
				return cluster, nil
			}
		}
	}
	return nil, errUnusedGKEClusterNotFound
}

// checkLease takes a clusterName and determines if there is an available lease
func checkLease(clusterMap *clusters.Map, leaseMap *leases.Map, clusterName string) (*container.Cluster, error) {
	cluster, _ := clusterMap.ClusterByName(clusterName)
	_, isLeased := leaseMap.LeaseByClusterName(clusterName)
	if !isLeased {
		return cluster, nil
	}
	return nil, errUnusedGKEClusterNotFound
}

func createKubeConfigFromCluster(c *container.Cluster) (*k8s.KubeConfig, error) {
	contextName := strings.ToLower(c.Name)
	authInfoName := contextName

	var clusters []k8s.NamedCluster
	cluster := k8s.Cluster{
		Server: fmt.Sprintf("https://%s", c.Endpoint),
		CertificateAuthorityData: c.MasterAuth.ClusterCaCertificate,
	}
	namedCluster := k8s.NamedCluster{
		Name:    c.Name,
		Cluster: cluster,
	}
	clusters = append(clusters, namedCluster)

	var contexts []k8s.NamedContext
	context := k8s.Context{
		Cluster:  c.Name,
		AuthInfo: authInfoName,
	}
	namedContext := k8s.NamedContext{
		Name:    contextName,
		Context: context,
	}
	contexts = append(contexts, namedContext)

	var authInfos []k8s.NamedAuthInfo
	authInfo := k8s.AuthInfo{
		ClientCertificateData: c.MasterAuth.ClientCertificate,
		ClientKeyData:         c.MasterAuth.ClientKey,
		Username:              c.MasterAuth.Username,
		Password:              c.MasterAuth.Password,
	}
	namedAuthInfo := k8s.NamedAuthInfo{
		Name:     authInfoName,
		AuthInfo: authInfo,
	}
	authInfos = append(authInfos, namedAuthInfo)

	return &k8s.KubeConfig{
		CurrentContext: contextName,
		APIVersion:     kubeconfigAPIVersion,
		Clusters:       clusters,
		Contexts:       contexts,
		AuthInfos:      authInfos,
	}, nil
}

func marshalAndEncodeKubeConfig(cfg *k8s.KubeConfig) (string, error) {
	y, err := yaml.Marshal(cfg)
	if err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(y), nil
}
