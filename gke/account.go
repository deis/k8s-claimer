package gke

import (
	"net/http"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/jwt"
	container "google.golang.org/api/container/v1"
)

const (
	// ContainerScope is the Oauth2 scope required for interaction with the GKE API
	ContainerScope = "https://www.googleapis.com/auth/cloud-platform"
	// TokenURL is the Oauth2 token exchange URL for Google accounts
	TokenURL = "https://accounts.google.com/o/oauth2/token"
)

// PrivateKey is the type for a JWT private key
type PrivateKey string

// String is the fmt.Stringer interface implementation
func (p PrivateKey) String() string {
	return string(p)
}

// Bytes is a convenience function for []byte(p.String())
func (p PrivateKey) Bytes() []byte {
	return []byte(p.String())
}

func getJWTConf(email string, pk PrivateKey) *jwt.Config {
	return &jwt.Config{
		Email:      email,
		PrivateKey: pk.Bytes(),
		Scopes:     []string{ContainerScope},
		TokenURL:   TokenURL,
	}
}

func getOAuthClient(conf *jwt.Config) *http.Client {
	return conf.Client(oauth2.NoContext)
}

// GetContainerService creates a GKE client by creating an OAuth2 capable HTTP client from the
// given JWT credentials, then creating a new container client with that HTTP client
func GetContainerService(email string, pk PrivateKey) (*container.Service, error) {
	conf := getJWTConf(email, pk)
	cl := getOAuthClient(conf)
	return container.New(cl)
}
