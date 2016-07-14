package handlers

import (
	"errors"
	"time"

	"github.com/deis/k8s-claimer/k8s"
	"github.com/deis/k8s-claimer/leases"
	"k8s.io/kubernetes/pkg/api"
	"k8s.io/kubernetes/pkg/client/restclient"
	kcl "k8s.io/kubernetes/pkg/client/unversioned"
)

var (
	errNoExpiredLeases     = errors.New("no expired leases exist")
	errNoClustersInConfig  = errors.New("no clusters in config")
	errNoAuthInfosInConfig = errors.New("no auth info in config")
)

// findExpiredLease searches in the leases in the svc annotations and returns the cluster name of
// the first expired lease it finds. If none found, returns an empty string and errNoExpiredLeases
func findExpiredLease(leaseMap *leases.Map) (*leases.UUIDAndLease, error) {
	now := time.Now()
	uuids, err := leaseMap.UUIDs()
	if err != nil {
		return nil, err
	}
	for _, u := range uuids {
		lease, _ := leaseMap.LeaseForUUID(u)
		exprTime, err := lease.ExpirationTime()
		if err != nil {
			return nil, err
		}
		if now.After(exprTime) {
			return leases.NewUUIDAndLease(u, lease), nil
		}
	}
	return nil, errNoExpiredLeases
}

func saveAnnotations(services k8s.ServiceUpdater, svc *api.Service, leaseMap *leases.Map) error {
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

// CreateKubeClientFromConfig creates a new Kubernetes client from the given configuration.
// returns nil and the appropriate error if the client couldn't be created for any reason
func CreateKubeClientFromConfig(conf *Config) (*kcl.Client, error) {
	rcConf := new(restclient.Config)
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

	return kcl.New(rcConf)
}
