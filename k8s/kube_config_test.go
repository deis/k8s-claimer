package k8s

import (
	"encoding/base64"
	"fmt"
	"strings"
	"testing"

	container "google.golang.org/api/container/v1"

	yaml "gopkg.in/yaml.v2"

	"github.com/arschles/assert"
)

func TestCreateKubeClientFromConfig(t *testing.T) {
	conf := &KubeConfig{
		Kind:        "testKind",
		APIVersion:  "v1",
		Preferences: Preferences{Colors: true},
		Clusters: []NamedCluster{
			NamedCluster{
				Name: "test1",
				Cluster: Cluster{
					Server:                   "test.server.com",
					APIVersion:               "v1",
					InsecureSkipTLSVerify:    false,
					CertificateAuthorityData: ca,
				},
			},
		},
		AuthInfos: []NamedAuthInfo{
			NamedAuthInfo{
				Name: "test1",
				AuthInfo: AuthInfo{
					ClientCertificateData: pubKey,
					ClientKeyData:         privKey,
					Impersonate:           "impersonate1",
					Username:              "testUser",
					Password:              "testPass",
				},
			},
		},
	}
	cl, err := CreateKubeClientFromConfig(conf)
	assert.NoErr(t, err)
	assert.NotNil(t, cl, "kube client")
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
	k8sConfig, err := CreateKubeConfigFromCluster(cluster)
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

func TestMarshalAndEncodeKubeConfig(t *testing.T) {
	const namespace = "myns"
	const locationOfOrigin = "myloc"

	contextNames := []string{"ctx1", "ctx2", "ctx3"}
	clusterNames := []string{"cluster1", "cluster2", "cluster3"}
	authInfoNames := []string{"authInfo1", "authInfo2", "authInfo3"}

	var contexts []NamedContext
	for i, contextName := range contextNames {
		context := Context{
			Cluster:   clusterNames[i],
			AuthInfo:  authInfoNames[i],
			Namespace: namespace,
		}
		namedContext := NamedContext{
			Name:    contextName,
			Context: context,
		}
		contexts = append(contexts, namedContext)
	}

	var clusters []NamedCluster
	for _, clusterName := range clusterNames {
		cluster := Cluster{
			Server:                   clusterName + "/server",
			APIVersion:               kubeconfigAPIVersion,
			InsecureSkipTLSVerify:    false,
			CertificateAuthorityData: clusterName + "_cert_authority",
		}
		namedCluster := NamedCluster{
			Name:    clusterName,
			Cluster: cluster,
		}
		clusters = append(clusters, namedCluster)
	}

	var authInfos []NamedAuthInfo
	for _, authInfoName := range authInfoNames {
		authInfo := AuthInfo{
			ClientCertificateData: authInfoName + "_cert_data",
			ClientKeyData:         authInfoName + "_key_data",
			Username:              authInfoName + "_username",
			Password:              authInfoName + "_password",
			Token:                 authInfoName + "_bearer_token",
		}
		namedAuthInfo := NamedAuthInfo{
			Name:     authInfoName,
			AuthInfo: authInfo,
		}
		authInfos = append(authInfos, namedAuthInfo)
	}

	cfg := &KubeConfig{
		CurrentContext: "ctx1",
		Clusters:       clusters,
		Contexts:       contexts,
		AuthInfos:      authInfos,
	}

	yamlString, err := MarshalAndEncodeKubeConfig(cfg)
	assert.NoErr(t, err)
	decodedBytes, err := base64.StdEncoding.DecodeString(yamlString)
	assert.NoErr(t, err)
	decodedCfg := new(KubeConfig)
	assert.NoErr(t, yaml.Unmarshal(decodedBytes, decodedCfg))
	assert.Equal(t, decodedCfg.APIVersion, cfg.APIVersion, "API version")
	assert.Equal(t, decodedCfg.AuthInfos[0].AuthInfo.ClientCertificateData, cfg.AuthInfos[0].AuthInfo.ClientCertificateData,
		"authInfo1_cert_data")
}

func TestCreateKubeConfig(t *testing.T) {
	config := `
---
apiVersion: v1
clusters:
- cluster:
    certificate-authority-data: "cert-authority-data"
    server: https://my-cluster.fqdn
  name: "cluster-name"
contexts:
- context:
    cluster: "context-cluster"
    user: "context-user"
  name: "context-name"
current-context: "current-context"
kind: Config
users:
- name: "users"
  user:
    client-certificate-data: "client-cert-data"
    client-key-data: "client-key-data"
  `
	kubeConfig, err := CreateKubeConfig([]byte(config))
	assert.NoErr(t, err)
	assert.Equal(t, kubeConfig.Clusters[0].Name, "cluster-name", "Cluster Name")
}
