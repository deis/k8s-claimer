package handlers

import (
	"encoding/base64"
	"encoding/json"
	"strings"
	"testing"

	"github.com/arschles/assert"
	"github.com/deis/k8s-claimer/clusters"
	"github.com/deis/k8s-claimer/gke"
	"github.com/deis/k8s-claimer/leases"
	"github.com/deis/k8s-claimer/testutil"
	"github.com/pborman/uuid"
	container "google.golang.org/api/container/v1"
	k8scmd "k8s.io/kubernetes/pkg/client/unversioned/clientcmd/api"
)

const (
	projID = "test project"
	zone   = "test zone"
)

func TestFindUnusedGKECluster(t *testing.T) {
	clusterNames := testutil.GetClusterNames()

	// test when all clusters are leased
	fakeLister := &gke.FakeClusterLister{
		Resp: &container.ListClustersResponse{Clusters: nil},
		Err:  nil,
	}
	for _, clusterName := range clusterNames {
		fakeLister.Resp.Clusters = append(fakeLister.Resp.Clusters, &container.Cluster{Name: clusterName})
	}

	clusterMap, err := clusters.ParseMapFromGKE(fakeLister, projID, zone)
	assert.NoErr(t, err)

	rawAnnotations := testutil.GetRawAnnotations(
		clusterNames,
		leases.TimeFormat,
		testutil.DefaultTimeFunc,
		testutil.DefaultUUIDFunc,
	)
	leaseMap, err := leases.ParseMapFromAnnotations(rawAnnotations)
	assert.NoErr(t, err)
	unusedCluster, err := findUnusedGKECluster(clusterMap, leaseMap)
	assert.True(t, unusedCluster == nil, "unused cluster returned non-nil when all clusters were in use")
	assert.Err(t, errUnusedGKEClusterNotFound, err)

	// test when there is a cluster that's not leased
	var freedUUID uuid.UUID
	for uuidStr := range rawAnnotations {
		parsedUUID := uuid.Parse(uuidStr)
		assert.True(t, parsedUUID != nil, "uuid parsed from %s was invalid", uuidStr)
		freedUUID = parsedUUID
		break
	}

	freedLease, found := leaseMap.LeaseForUUID(freedUUID)
	assert.True(t, found, "lease for uuid %s not found", freedUUID)
	deleted := leaseMap.DeleteLease(freedUUID)
	assert.True(t, deleted, "lease for cluster %s was not deleted", freedLease.ClusterName)

	unusedCluster, err = findUnusedGKECluster(clusterMap, leaseMap)
	assert.NoErr(t, err)
	assert.Equal(t, unusedCluster.Name, freedLease.ClusterName, "free cluster name")
}

func TestCreateKubeConfigFromCluster(t *testing.T) {
	cluster := &container.Cluster{
		Name:     "my cluster",
		Endpoint: "https://my.k8s.endpoint.com",
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

	clusterConfig, clusterConfigFound := k8sConfig.Clusters[cluster.Name]
	assert.True(t, clusterConfigFound, "cluster %s not found in config", cluster.Name)
	assert.Equal(t, clusterConfig.Server, cluster.Endpoint, "cluster endpoint")
	assert.Equal(t, string(clusterConfig.CertificateAuthorityData), cluster.MasterAuth.ClusterCaCertificate, "cluster CA data")

	authInfoConfig, authInfoFound := k8sConfig.AuthInfos[strings.ToLower(cluster.Name)]
	assert.True(t, authInfoFound, "auth info for cluster %s not found in config", cluster.Name)
	assert.Equal(t, string(authInfoConfig.ClientCertificateData), cluster.MasterAuth.ClientCertificate, "client certificate")
	assert.Equal(t, string(authInfoConfig.ClientKeyData), cluster.MasterAuth.ClientKey, "client key")
	assert.Equal(t, authInfoConfig.Username, cluster.MasterAuth.Username, "username")
	assert.Equal(t, authInfoConfig.Password, cluster.MasterAuth.Password, "password")

	contextConfig, contextFound := k8sConfig.Contexts[strings.ToLower(cluster.Name)]
	assert.True(t, contextFound, "context for cluster %s not found in config", cluster.Name)
	_, clusterFoundFromContext := k8sConfig.Clusters[contextConfig.Cluster]
	assert.True(t, clusterFoundFromContext, "cluster not found from context.Cluster value %s", contextConfig.Cluster)
	_, authFoundFromContext := k8sConfig.AuthInfos[contextConfig.AuthInfo]
	assert.True(t, authFoundFromContext, "auth info not found from context.AuthInfo value %s", contextConfig.AuthInfo)
}

type contextClusterAndAuthInfo struct {
	contextName  string
	clusterName  string
	authInfoName string
}

func TestMarshalAndEncodeKubeConfig(t *testing.T) {
	const namespace = "myns"
	const locationOfOrigin = "myloc"
	cfg := &k8scmd.Config{
		APIVersion: kubeconfigAPIVersion,
		Clusters:   map[string]*k8scmd.Cluster{},
		AuthInfos:  map[string]*k8scmd.AuthInfo{},
		Contexts:   map[string]*k8scmd.Context{},
	}

	contextNames := []string{"ctx1", "ctx2", "ctx3"}
	clusterNames := []string{"cluster1", "cluster2", "cluster3"}
	authInfoNames := []string{"authInfo1", "authInfo2", "authInfo3"}
	for i, contextName := range contextNames {
		cfg.CurrentContext = contextName
		cfg.Contexts[contextName] = &k8scmd.Context{
			LocationOfOrigin: locationOfOrigin,
			Cluster:          clusterNames[i],
			AuthInfo:         authInfoNames[i],
			Namespace:        namespace,
		}
	}
	for _, clusterName := range clusterNames {
		cfg.Clusters[clusterName] = &k8scmd.Cluster{
			LocationOfOrigin:         locationOfOrigin,
			Server:                   clusterName + "/server",
			APIVersion:               kubeconfigAPIVersion,
			InsecureSkipTLSVerify:    false,
			CertificateAuthorityData: []byte(clusterName + "_cert_authority"),
		}
	}
	for _, authInfoName := range authInfoNames {
		cfg.AuthInfos[authInfoName] = &k8scmd.AuthInfo{
			LocationOfOrigin:      locationOfOrigin,
			ClientCertificateData: []byte(authInfoName + "_cert_data"),
			ClientKeyData:         []byte(authInfoName + "_key_data"),
			Token:                 authInfoName + "_bearer_token",
			Username:              authInfoName + "_username",
			Password:              authInfoName + "_password",
		}
	}
	str, err := marshalAndEncodeKubeConfig(cfg)
	assert.NoErr(t, err)
	decodedBytes, err := base64.StdEncoding.DecodeString(str)
	assert.NoErr(t, err)
	decodedCfg := new(k8scmd.Config)
	assert.NoErr(t, json.Unmarshal(decodedBytes, decodedCfg))
	assert.Equal(t, decodedCfg.APIVersion, cfg.APIVersion, "API version")

}
