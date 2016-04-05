package main

import (
	"log"
	"net/http"

	"github.com/deis/k8s-claimer/gke"
	"github.com/deis/k8s-claimer/handlers"
	"github.com/deis/k8s-claimer/htp"
	kcl "k8s.io/kubernetes/pkg/client/unversioned"
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

	k8sClient, err := kcl.NewInCluster()
	if err != nil {
		log.Fatalf("Error creating Kubernetes client (%s)", err)
	}

	services := k8sClient.Services(serverConf.Namespace)
	mux := http.NewServeMux()
	leaseHandler := htp.MethodMux(map[htp.Method]http.Handler{
		htp.Post:   handlers.CreateLease(containerService, services),
		htp.Delete: handlers.DeleteLease(services),
	})
	mux.Handle("/lease", leaseHandler)

	log.Printf("Running %s on %s", appName, serverConf.HostStr())
	http.ListenAndServe(serverConf.HostStr(), mux)
}
