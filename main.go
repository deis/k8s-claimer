package main

import (
	"log"
	"net/http"

	"github.com/deis/k8s-claimer/gke"
	"github.com/deis/k8s-claimer/handlers"
	"github.com/deis/k8s-claimer/htp"
)

const (
	appName = "k8s-claimer"
)

func main() {
	serverConf, err := parseServerConfig(appName)
	if err != nil {
		log.Fatalf("Error getting server config (%s)", err)
	}
	gCloudConfFile, err := parseGoogleConfigFile(appName)
	if err != nil {
		log.Fatalf("Error getting google cloud config (%s)", err)
	}
	containerService, err := gke.GetContainerService(
		gCloudConfFile.ClientEmail,
		gke.PrivateKey(gCloudConfFile.PrivateKey),
	)
	if err != nil {
		log.Fatalf("Error creating GKE client (%s)", err)
	}

	mux := http.NewServeMux()
	leaseHandler := htp.MethodMux(map[htp.Method]http.Handler{
		htp.Post:   handlers.CreateLease(containerService),
		htp.Delete: handlers.DeleteLease(),
	})
	mux.Handle("/lease", leaseHandler)

	log.Printf("Running %s on %s", appName, serverConf.HostStr())
	http.ListenAndServe(serverConf.HostStr(), mux)
}
