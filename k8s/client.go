package k8s

import (
	"errors"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

var (
	errNoClustersInConfig  = errors.New("no clusters in config")
	errNoAuthInfosInConfig = errors.New("no auth info in config")
)

// CreateKubeClientFromConfig creates a new Kubernetes client from the given configuration.
// returns nil and the appropriate error if the client couldn't be created for any reason
func CreateKubeClientFromConfig(conf *KubeConfig) (*kubernetes.Clientset, error) {
	rcConf := new(rest.Config)
	if len(conf.Clusters) < 1 {
		return nil, errNoClustersInConfig
	}
	cluster := conf.Clusters[0].Cluster
	if len(conf.AuthInfos) < 1 {
		return nil, errNoAuthInfosInConfig
	}
	authInfo := conf.AuthInfos[0].AuthInfo

	rcConf.Host = cluster.Server
	rcConf.Username = authInfo.Username
	rcConf.Password = authInfo.Password
	rcConf.TLSClientConfig.CertData = []byte(authInfo.ClientCertificateData)
	rcConf.TLSClientConfig.KeyData = []byte(authInfo.ClientKeyData)
	rcConf.TLSClientConfig.CAData = []byte(cluster.CertificateAuthorityData)
	rcConf.BearerToken = authInfo.Token
	rcConf.UserAgent = "k8s-claimer"
	rcConf.Insecure = cluster.InsecureSkipTLSVerify

	return kubernetes.NewForConfig(rcConf)
}
