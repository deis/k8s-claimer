package main

import (
	"github.com/deis/k8s-claimer/config"
	"github.com/kelseyhightower/envconfig"
)

func parseGoogleConfig(appName string) (*config.GoogleCloud, error) {
	conf := new(config.GoogleCloud)
	if err := envconfig.Process(appName, conf); err != nil {
		return nil, err
	}
	return conf, nil
}

func parseServerConfig(appName string) (*config.Server, error) {
	conf := new(config.Server)
	if err := envconfig.Process(appName, conf); err != nil {
		return nil, err
	}
	return conf, nil
}

