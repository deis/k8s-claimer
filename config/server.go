package config

import (
	"fmt"
	"log"
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

// Print will render the current server configuration
func (s Server) Print() {
	log.Println("Server Configuration:")
	log.Printf("\tListening:%s:%v\n", s.BindHost, s.BindPort)
	log.Printf("\tNamespace:%s\n", s.Namespace)
	log.Printf("\tService Name:%s\n", s.ServiceName)
	log.Printf("\tAuth Token:%s\n", s.AuthToken)
	log.Printf("\tClear Namespaces?:%v\n", s.ClearNamespaces)
}
