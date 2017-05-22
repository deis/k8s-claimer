package main

import (
	"bytes"
	"encoding/json"

	"github.com/deis/k8s-claimer/config"
	"github.com/kelseyhightower/envconfig"
)

func parseAzureConfig(appName string) (*config.Azure, error) {
	conf := new(config.Azure)
	if err := envconfig.Process(appName, conf); err != nil {
		return nil, err
	}
	return conf, nil
}

func parseGoogleConfig(appName string) (*config.Google, error) {
	conf := new(config.Google)
	if err := envconfig.Process(appName, conf); err != nil {
		return nil, err
	}

	gCloudConfFile := new(config.AccountFile)
	if err := json.NewDecoder(bytes.NewBuffer([]byte(conf.AccountFileJSON))).Decode(gCloudConfFile); err != nil {
		return nil, err
	}
	conf.AccountFile = *gCloudConfFile
	return conf, nil
}

func parseServerConfig(appName string) (*config.Server, error) {
	conf := new(config.Server)
	if err := envconfig.Process(appName, conf); err != nil {
		return nil, err
	}
	return conf, nil
}
