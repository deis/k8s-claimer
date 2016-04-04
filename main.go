package main

import (
	"log"
	"net/http"

	"github.com/deis/k8s-claimer/config"
	"github.com/deis/k8s-claimer/handlers"
	"github.com/deis/k8s-claimer/htp"
	"github.com/kelseyhightower/envconfig"
)

const (
	appName = "k8s-claimer"
)

func main() {
	conf := new(config.Server)
	err := envconfig.Process(appName, conf)
	if err != nil {
		log.Fatalf("Error getting config (%s)", err)
	}
	mux := http.NewServeMux()
	leaseHandler := htp.MethodMux(map[htp.Method]http.Handler{
		htp.Post:   handlers.CreateLease(),
		htp.Delete: handlers.DeleteLease(),
	})
	mux.Handle("/lease", leaseHandler)

	log.Printf("Running %s on %s", appName, conf.HostStr())
	http.ListenAndServe(conf.HostStr(), mux)
}
