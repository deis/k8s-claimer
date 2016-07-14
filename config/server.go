package config

import (
	"fmt"
)

// Server represents the envconfig-compatible server configuration
type Server struct {
	BindHost        string `envconfig:"BIND_HOST" default:"0.0.0.0"`
	BindPort        int    `envconfig:"BIND_PORT" default:"8080"`
	Namespace       string `envconfig:"NAMESPACE" default:"k8s-claimer"`
	ServiceName     string `envconfig:"SERVICE_NAME" default:"k8s-claimer"`
	AuthToken       string `envconfig:"AUTH_TOKEN" required:"true"`
	ClearNamespaces bool   `envconfig:"CLEAR_NAMESPACES" default:"false"`
}

// HostStr returns the full host string for the server, based on s.BindHost and s.BindPort
func (s Server) HostStr() string {
	return fmt.Sprintf("%s:%d", s.BindHost, s.BindPort)
}
