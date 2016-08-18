package handlers

import (
	"encoding/base64"
	"errors"
	"fmt"
	"math/rand"
	"regexp"
	"strings"

	"github.com/deis/k8s-claimer/clusters"
	"github.com/deis/k8s-claimer/leases"
	container "google.golang.org/api/container/v1"
	"gopkg.in/yaml.v2"
	"k8s.io/kubernetes/pkg/runtime"
)

const (
	kubeconfigAPIVersion = "v1"
)

var (
	errUnusedGKEClusterNotFound = errors.New("all GKE clusters are in use")
)

// Config holds the information needed to connect to remote kubernetes clusters as a given user
type Config struct {
	Kind           string          `yaml:"kind,omitempty"`
	APIVersion     string          `yaml:"apiVersion,omitempty"`
	Preferences    Preferences     `yaml:"preferences"`
	Clusters       []NamedCluster  `yaml:"clusters"`
	AuthInfos      []NamedAuthInfo `yaml:"users"`
	Contexts       []NamedContext  `yaml:"contexts"`
	CurrentContext string          `yaml:"current-context"`
}

// Preferences prefs
type Preferences struct {
	Colors     bool             `yaml:"colors,omitempty"`
	Extensions []NamedExtension `yaml:"extensions,omitempty"`
}

// Cluster contains information about how to communicate with a kubernetes cluster
type Cluster struct {
	Server                   string           `yaml:"server"`
	APIVersion               string           `yaml:"api-version,omitempty"`
	InsecureSkipTLSVerify    bool             `yaml:"insecure-skip-tls-verify,omitempty"`
	CertificateAuthority     string           `yaml:"certificate-authority,omitempty"`
	CertificateAuthorityData string           `yaml:"certificate-authority-data,omitempty"`
	Extensions               []NamedExtension `yaml:"extensions,omitempty"`
}

// AuthInfo contains information that describes identity information.  This is use to tell the kubernetes cluster who you are.
type AuthInfo struct {
	ClientCertificate     string              `yaml:"client-certificate,omitempty"`
	ClientCertificateData string              `yaml:"client-certificate-data,omitempty"`
	ClientKey             string              `yaml:"client-key,omitempty"`
	ClientKeyData         string              `yaml:"client-key-data,omitempty"`
	Token                 string              `yaml:"token,omitempty"`
	Impersonate           string              `yaml:"as,omitempty"`
	Username              string              `yaml:"username,omitempty"`
	Password              string              `yaml:"password,omitempty"`
	AuthProvider          *AuthProviderConfig `yaml:"auth-provider,omitempty"`
	Extensions            []NamedExtension    `yaml:"extensions,omitempty"`
}

// Context is a tuple of references to a cluster (how do I communicate with a kubernetes cluster), a user (how do I identify myself), and a namespace (what subset of resources do I want to work with)
type Context struct {
	Cluster    string           `yaml:"cluster"`
	AuthInfo   string           `yaml:"user"`
	Namespace  string           `yaml:"namespace,omitempty"`
	Extensions []NamedExtension `yaml:"extensions,omitempty"`
}

// NamedCluster relates nicknames to cluster information
type NamedCluster struct {
	Name    string  `yaml:"name"`
	Cluster Cluster `yaml:"cluster"`
}

// NamedContext relates nicknames to context information
type NamedContext struct {
	Name    string  `yaml:"name"`
	Context Context `yaml:"context"`
}

// NamedAuthInfo relates nicknames to auth information
type NamedAuthInfo struct {
	Name     string   `yaml:"name"`
	AuthInfo AuthInfo `yaml:"user"`
}

// NamedExtension relates nicknames to extension information
type NamedExtension struct {
	Name      string               `yaml:"name"`
	Extension runtime.RawExtension `yaml:"extension"`
}

// AuthProviderConfig holds the configuration for a specified auth provider.
type AuthProviderConfig struct {
	Name   string            `yaml:"name"`
	Config map[string]string `yaml:"config"`
}

// findUnusedGKECluster finds a GKE cluster that's not currently in use according to the
// annotations in svc. It will also attempt to match the clusterRegex passed in if possible.
// Returns errUnusedGKEClusterNotFound if none is found
func findUnusedGKECluster(clusterMap *clusters.Map, leaseMap *leases.Map, clusterRegex string) (*container.Cluster, error) {
	if clusterRegex != "" {
		return findUnusuedGKEClusterByName(clusterMap, leaseMap, clusterRegex)
	}
	return findRandomUnusuedGKECluster(clusterMap, leaseMap)
}

// findUnusuedGKEClusterByName attempts to find a unused GKE cluster that matches the regex passed in via the cli.
func findUnusuedGKEClusterByName(clusterMap *clusters.Map, leaseMap *leases.Map, clusterRegex string) (*container.Cluster, error) {
	regex, err := regexp.Compile(clusterRegex)
	if err != nil {
		return nil, err
	}
	for _, clusterName := range clusterMap.Names() {
		if regex.MatchString(clusterName) {
			cluster, _ := clusterMap.ClusterByName(clusterName)
			_, found := leaseMap.LeaseByClusterName(clusterName)
			if !found {
				return cluster, nil
			}
		}
	}
	return nil, errUnusedGKEClusterNotFound
}

// findUnusuedGKECluster attempts to find a random unused GKE cluster
func findRandomUnusuedGKECluster(clusterMap *clusters.Map, leaseMap *leases.Map) (*container.Cluster, error) {
	clusterNames := clusterMap.Names()
	for tries := 0; tries < 10; tries++ {
		if len(clusterNames) > 0 {
			clusterName := clusterNames[rand.Intn(len(clusterNames))]
			cluster, _ := clusterMap.ClusterByName(clusterName)
			_, found := leaseMap.LeaseByClusterName(clusterName)
			if !found {
				return cluster, nil
			}
		}
	}
	return nil, errUnusedGKEClusterNotFound
}

func createKubeConfigFromCluster(c *container.Cluster) (*Config, error) {
	contextName := strings.ToLower(c.Name)
	authInfoName := contextName

	var clusters []NamedCluster
	cluster := Cluster{
		Server: fmt.Sprintf("https://%s", c.Endpoint),
		CertificateAuthorityData: c.MasterAuth.ClusterCaCertificate,
	}
	namedCluster := NamedCluster{
		Name:    c.Name,
		Cluster: cluster,
	}
	clusters = append(clusters, namedCluster)

	var contexts []NamedContext
	context := Context{
		Cluster:  c.Name,
		AuthInfo: authInfoName,
	}
	namedContext := NamedContext{
		Name:    contextName,
		Context: context,
	}
	contexts = append(contexts, namedContext)

	var authInfos []NamedAuthInfo
	authInfo := AuthInfo{
		ClientCertificateData: c.MasterAuth.ClientCertificate,
		ClientKeyData:         c.MasterAuth.ClientKey,
		Username:              c.MasterAuth.Username,
		Password:              c.MasterAuth.Password,
	}
	namedAuthInfo := NamedAuthInfo{
		Name:     authInfoName,
		AuthInfo: authInfo,
	}
	authInfos = append(authInfos, namedAuthInfo)

	return &Config{
		CurrentContext: contextName,
		APIVersion:     kubeconfigAPIVersion,
		Clusters:       clusters,
		Contexts:       contexts,
		AuthInfos:      authInfos,
	}, nil
}

func marshalAndEncodeKubeConfig(cfg *Config) (string, error) {
	y, err := yaml.Marshal(cfg)
	if err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(y), nil
}
