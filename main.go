package main

import (
	"log"
	"net/http"

	"github.com/deis/k8s-claimer/config"
	"github.com/deis/k8s-claimer/gke"
	"github.com/deis/k8s-claimer/handlers"
	"github.com/deis/k8s-claimer/htp"
	kcl "k8s.io/kubernetes/pkg/client/unversioned"
)

const (
	appName      = "k8s-claimer"
	authTokenKey = "Authorization"
)

func configureRoutesWithAuth(serveMux *http.ServeMux, createLeaseHandler http.Handler, deleteLeaseHandler http.Handler, authToken string) {
	createLeaseHandler = htp.MethodMux(map[htp.Method]http.Handler{htp.Post: createLeaseHandler})
	deleteLeaseHandler = htp.MethodMux(map[htp.Method]http.Handler{htp.Delete: deleteLeaseHandler})
	serveMux.Handle("/lease", handlers.WithAuth(authToken, authTokenKey, createLeaseHandler))
	serveMux.Handle("/lease/", handlers.WithAuth(authToken, authTokenKey, deleteLeaseHandler))
}

func main() {
	serverConf, err := parseServerConfig(appName)
	if err != nil {
		log.Fatalf("Error getting server config (%s)", err)
	}
	gCloudConf, err := parseGoogleConfig(appName)
	if err != nil {
		log.Fatalf("Error getting google cloud config (%s)", err)
	}
	gCloudConfFile, err := config.GoogleCloudAccountInfo(gCloudConf.AccountFileBase64)
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
	createLeaseHandler := handlers.CreateLease(
		gke.NewGKEClusterLister(containerService),
		services,
		serverConf.ServiceName,
		gCloudConf.ProjectID,
		gCloudConf.Zone,
	)
	deleteLeaseHandler := handlers.DeleteLease(services, serverConf.ServiceName, k8sClient.Namespaces())
	configureRoutesWithAuth(mux, createLeaseHandler, deleteLeaseHandler, serverConf.AuthToken)

	log.Printf("Running %s on %s", appName, serverConf.HostStr())
	http.ListenAndServe(serverConf.HostStr(), mux)
}
