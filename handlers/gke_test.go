package handlers

import (
	"encoding/base64"
	"fmt"
	"strings"
	"testing"

	"github.com/arschles/assert"
	"github.com/deis/k8s-claimer/clusters"
	"github.com/deis/k8s-claimer/gke"
	"github.com/deis/k8s-claimer/k8s"
	"github.com/deis/k8s-claimer/leases"
	"github.com/deis/k8s-claimer/testutil"
	container "google.golang.org/api/container/v1"
	"gopkg.in/yaml.v2"
)

const (
	projID = "test project"
	zone   = "test zone"
)

func TestFindUnusedGKEClusterByName(t *testing.T) {
	leaseableClusters := testutil.GetClusters()

	// test when all clusters are leased
	fakeLister := &gke.FakeClusterLister{
		Resp: &container.ListClustersResponse{Clusters: leaseableClusters},
		Err:  nil,
	}

	clusterMap, err := clusters.ParseMapFromGKE(fakeLister, projID, zone)
	assert.NoErr(t, err)
	leaseMap, err := leases.ParseMapFromAnnotations(nil)

	unusedCluster, err := findUnusedGKECluster(clusterMap, leaseMap, "getClusterByName", "")
	assert.NoErr(t, err)
	assert.Equal(t, unusedCluster.Name, "getClusterByName", "free cluster name")
}

func TestFindUnusedGKEClusterByVersion(t *testing.T) {
	leaseableClusters := testutil.GetClusters()

	// test when all clusters are leased
	fakeLister := &gke.FakeClusterLister{
		Resp: &container.ListClustersResponse{Clusters: leaseableClusters},
		Err:  nil,
	}

	clusterMap, err := clusters.ParseMapFromGKE(fakeLister, projID, zone)
	assert.NoErr(t, err)
	leaseMap, err := leases.ParseMapFromAnnotations(nil)

	unusedCluster, err := findUnusedGKECluster(clusterMap, leaseMap, "", "1.1.1")
	assert.NoErr(t, err)
	assert.Equal(t, unusedCluster.Name, "getClusterByVersion", "free cluster name")
	assert.Equal(t, unusedCluster.CurrentNodeVersion, "1.1.1", "free cluster version")
}

func TestFindRandomUnusedGKECluster(t *testing.T) {
	leaseableClusters := testutil.GetClusters()

	// test when all clusters are leased
	fakeLister := &gke.FakeClusterLister{
		Resp: &container.ListClustersResponse{Clusters: leaseableClusters},
		Err:  nil,
	}

	clusterMap, err := clusters.ParseMapFromGKE(fakeLister, projID, zone)
	assert.NoErr(t, err)
	leaseMap, err := leases.ParseMapFromAnnotations(nil)

	unusedCluster, err := findUnusedGKECluster(clusterMap, leaseMap, "", "")
	assert.NoErr(t, err)
	assert.NotNil(t, unusedCluster, "free cluster name")
}

func TestCreateKubeConfigFromCluster(t *testing.T) {
	cluster := &container.Cluster{
		Name:     "my cluster",
		Endpoint: "my.k8s.endpoint.com",
		MasterAuth: &container.MasterAuth{
			ClientCertificate:    "test client cert",
			ClientKey:            "test client key",
			ClusterCaCertificate: "test cluster CA",
			Password:             "test password",
			Username:             "test username",
		},
	}
	k8sConfig, err := createKubeConfigFromCluster(cluster)
	assert.NoErr(t, err)
	assert.Equal(t, k8sConfig.APIVersion, "v1", "api version")
	assert.Equal(t, k8sConfig.CurrentContext, strings.ToLower(cluster.Name), "context")

	namedCluster := k8sConfig.Clusters[0]
	assert.Equal(t, namedCluster.Cluster.Server, fmt.Sprintf("https://%s", cluster.Endpoint), "cluster endpoint")
	assert.Equal(t, string(namedCluster.Cluster.CertificateAuthorityData), cluster.MasterAuth.ClusterCaCertificate, "cluster CA data")

	namedAuthInfo := k8sConfig.AuthInfos[0]
	assert.Equal(t, string(namedAuthInfo.AuthInfo.ClientCertificateData), cluster.MasterAuth.ClientCertificate, "client certificate")
	assert.Equal(t, string(namedAuthInfo.AuthInfo.ClientKeyData), cluster.MasterAuth.ClientKey, "client key")
	assert.Equal(t, namedAuthInfo.AuthInfo.Username, cluster.MasterAuth.Username, "username")
	assert.Equal(t, namedAuthInfo.AuthInfo.Password, cluster.MasterAuth.Password, "password")

	contextConfig := k8sConfig.Contexts[0]
	assert.Equal(t, contextConfig.Context.Cluster, cluster.Name, "cluster name")
	assert.Equal(t, contextConfig.Context.AuthInfo, cluster.Name, "auth info name")
}

type contextClusterAndAuthInfo struct {
	contextName  string
	clusterName  string
	authInfoName string
}

func TestMarshalAndEncodeKubeConfig(t *testing.T) {
	const namespace = "myns"
	const locationOfOrigin = "myloc"

	contextNames := []string{"ctx1", "ctx2", "ctx3"}
	clusterNames := []string{"cluster1", "cluster2", "cluster3"}
	authInfoNames := []string{"authInfo1", "authInfo2", "authInfo3"}

	var contexts []k8s.NamedContext
	for i, contextName := range contextNames {
		context := k8s.Context{
			Cluster:   clusterNames[i],
			AuthInfo:  authInfoNames[i],
			Namespace: namespace,
		}
		namedContext := k8s.NamedContext{
			Name:    contextName,
			Context: context,
		}
		contexts = append(contexts, namedContext)
	}

	var clusters []k8s.NamedCluster
	for _, clusterName := range clusterNames {
		cluster := k8s.Cluster{
			Server:                   clusterName + "/server",
			APIVersion:               kubeconfigAPIVersion,
			InsecureSkipTLSVerify:    false,
			CertificateAuthorityData: clusterName + "_cert_authority",
		}
		namedCluster := k8s.NamedCluster{
			Name:    clusterName,
			Cluster: cluster,
		}
		clusters = append(clusters, namedCluster)
	}

	var authInfos []k8s.NamedAuthInfo
	for _, authInfoName := range authInfoNames {
		authInfo := k8s.AuthInfo{
			ClientCertificateData: authInfoName + "_cert_data",
			ClientKeyData:         authInfoName + "_key_data",
			Username:              authInfoName + "_username",
			Password:              authInfoName + "_password",
			Token:                 authInfoName + "_bearer_token",
		}
		namedAuthInfo := k8s.NamedAuthInfo{
			Name:     authInfoName,
			AuthInfo: authInfo,
		}
		authInfos = append(authInfos, namedAuthInfo)
	}

	cfg := &k8s.KubeConfig{
		CurrentContext: "ctx1",
		Clusters:       clusters,
		Contexts:       contexts,
		AuthInfos:      authInfos,
	}

	yamlString, err := marshalAndEncodeKubeConfig(cfg)
	assert.NoErr(t, err)
	decodedBytes, err := base64.StdEncoding.DecodeString(yamlString)
	assert.NoErr(t, err)
	decodedCfg := new(k8s.KubeConfig)
	assert.NoErr(t, yaml.Unmarshal(decodedBytes, decodedCfg))
	assert.Equal(t, decodedCfg.APIVersion, cfg.APIVersion, "API version")
	assert.Equal(t, decodedCfg.AuthInfos[0].AuthInfo.ClientCertificateData, cfg.AuthInfos[0].AuthInfo.ClientCertificateData,
		"authInfo1_cert_data")
}
