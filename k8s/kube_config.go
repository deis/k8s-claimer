package k8s

import (
	"encoding/base64"
	"fmt"
	"log"
	"strings"

	container "google.golang.org/api/container/v1"

	"gopkg.in/yaml.v2"
	"k8s.io/client-go/pkg/runtime"
)

const (
	kubeconfigAPIVersion = "v1"
)

// KubeConfig holds the information needed to connect to remote kubernetes clusters as a given user
type KubeConfig struct {
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

// MarshalAndEncodeKubeConfig takes a kubeconfig and base64 encodes it
func MarshalAndEncodeKubeConfig(cfg *KubeConfig) (string, error) {
	y, err := yaml.Marshal(cfg)
	if err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(y), nil
}

// CreateKubeConfig will take a yaml kubeconfig and return a parse KubeConfig object
func CreateKubeConfig(k []byte) (*KubeConfig, error) {
	kubeConfig := KubeConfig{}
	err := yaml.Unmarshal(k, &kubeConfig)
	if err != nil {
		log.Printf("Error parsing yaml:%s\n", err)
		return nil, err
	}
	return &kubeConfig, nil
}

// CreateKubeConfigFromCluster generates a valid KubeConfig object from a leased cluster
func CreateKubeConfigFromCluster(c *container.Cluster) (*KubeConfig, error) {
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

	return &KubeConfig{
		CurrentContext: contextName,
		APIVersion:     kubeconfigAPIVersion,
		Clusters:       clusters,
		Contexts:       contexts,
		AuthInfos:      authInfos,
	}, nil
}
