package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"log"
	"net/http"

	"github.com/deis/k8s-claimer/config"
	"github.com/deis/k8s-claimer/handlers"
	"github.com/deis/k8s-claimer/htp"
	"github.com/deis/k8s-claimer/k8s"
	"github.com/deis/k8s-claimer/providers/gke"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

const (
	appName      = "k8s-claimer"
	authTokenKey = "Authorization"
)

var (
	errNilConfig = errors.New("nil config")
)

func kubeNamespacesFromConfig() func(*k8s.KubeConfig) (k8s.NamespaceListerDeleter, error) {
	return func(conf *k8s.KubeConfig) (k8s.NamespaceListerDeleter, error) {
		if conf == nil {
			return nil, errNilConfig
		}
		cl, err := k8s.CreateKubeClientFromConfig(conf)
		if err != nil {
			return nil, err
		}
		return cl.Namespaces(), nil
	}
}

func configureRoutesWithAuth(serveMux *http.ServeMux, createLeaseHandler http.Handler, deleteLeaseHandler http.Handler, authToken string) {
	createLeaseHandler = htp.MethodMux(map[htp.Method]http.Handler{htp.Post: createLeaseHandler})
	deleteLeaseHandler = htp.MethodMux(map[htp.Method]http.Handler{htp.Delete: deleteLeaseHandler})

	serveMux.Handle("/lease", handlers.WithAuth(authToken, authTokenKey, createLeaseHandler))
	serveMux.Handle("/lease/", handlers.WithAuth(authToken, authTokenKey, deleteLeaseHandler))
}

//CreateHealthzHandler returns an http.Handler
func CreateHealthzHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
}

func main() {
	serverConf, err := parseServerConfig(appName)
	if err != nil {
		log.Fatalf("Error getting server config (%s)", err)
	}
	serverConf.Print()
	gCloudConf, err := parseGoogleConfig(appName)
	if err != nil {
		log.Fatalf("Error getting google cloud config (%s) -- %+v", err, gCloudConf)
	}
	gCloudConfFile := new(config.GoogleCloudAccountFile)
	if err := json.NewDecoder(bytes.NewBuffer([]byte(gCloudConf.AccountFile))).Decode(gCloudConfFile); err != nil {
		log.Fatalf("Error parsing google cloud config (%s)\n %+v", err, gCloudConf.AccountFile)
	}
	containerService, err := gke.GetContainerService(
		gCloudConfFile.ClientEmail,
		gke.PrivateKey(gCloudConfFile.PrivateKey),
	)
	if err != nil {
		log.Fatalf("Error creating GKE client (%s)", err)
	}

	config, err := rest.InClusterConfig()
	if err != nil {
		log.Fatalf("Error creating Kubernetes client (%s)", err)
	}
	k8sClient, err := kubernetes.NewForConfig(config)
	if err != nil {
		log.Fatalf("Error creating Kubernetes client (%s)", err)
	}

	services := k8sClient.Services(serverConf.Namespace)
	gkeClusterLister := gke.NewGKEClusterLister(containerService)
	mux := http.NewServeMux()
	createLeaseHandler := handlers.CreateLease(
		gkeClusterLister,
		services,
		serverConf.ServiceName,
		gCloudConf.ProjectID,
		gCloudConf.Zone,
	)
	deleteLeaseHandler := handlers.DeleteLease(
		services,
		gkeClusterLister,
		serverConf.ServiceName,
		gCloudConf.ProjectID,
		gCloudConf.Zone,
		serverConf.ClearNamespaces,
		kubeNamespacesFromConfig(),
	)

	mux.Handle("/healthz", CreateHealthzHandler())

	configureRoutesWithAuth(mux, createLeaseHandler, deleteLeaseHandler, serverConf.AuthToken)

	log.Println("k8s claimer started!")
	http.ListenAndServe(serverConf.HostStr(), mux)
}
