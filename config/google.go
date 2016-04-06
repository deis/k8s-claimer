package config

import (
	"encoding/json"
	"os"
)

// GoogleCloud contains the Google cloud related configuration, including credentials and
// project info
type GoogleCloud struct {
	AccountFileLocation string `envconfig:"GOOGLE_CLOUD_ACCOUNT_FILE_LOCATION" required:"true"`
	ProjectID           string `envconfig:"GOOGLE_CLOUD_PROJECT_ID" required:"true"`
	Zone                string `envconfig:"GOOGLE_CLOUD_ZONE" required:"true"`
}

// GoogleCloudAccountFile represents the structure of the account file JSON file.
// this struct was adapted from the Terraform project
// (https://github.com/hashicorp/terraform/blob/master/builtin/providers/google/config.go)
type GoogleCloudAccountFile struct {
	// PrivateKeyID is the private key ID given in the account file
	PrivateKeyID string `json:"private_key_id"`
	// PrivateKey is the private key given in the account file
	PrivateKey string `json:"private_key"`
	// ClientEmail is the client email given in the account file
	ClientEmail string `json:"client_email"`
	// ClientID is the client ID given in the account file
	ClientID string `json:"client_id"`
}

// GoogleCloudAccountInfo parses the file at fileLoc into a GoogleCloudAccountFile and returns
// it. Returns nil and an appropriate error if any error occurred in reading or decoding.
func GoogleCloudAccountInfo(fileLoc string) (*GoogleCloudAccountFile, error) {
	fd, err := os.Open(fileLoc)
	if err != nil {
		return nil, err
	}
	defer fd.Close()
	ret := new(GoogleCloudAccountFile)
	if err := json.NewDecoder(fd).Decode(ret); err != nil {
		return nil, err
	}
	return ret, nil
}
