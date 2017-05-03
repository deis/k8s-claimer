package config

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
)

// GoogleCloud contains the Google cloud related configuration, including credentials and
// project info
type GoogleCloud struct {
	AccountFile string `envconfig:"GOOGLE_CLOUD_ACCOUNT_FILE" required:"true"`
	ProjectID   string `envconfig:"GOOGLE_CLOUD_PROJECT_ID" required:"true"`
	Zone        string `envconfig:"GOOGLE_CLOUD_ZONE" default:"-"`
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
func GoogleCloudAccountInfo(fileBase64 string) (*GoogleCloudAccountFile, error) {
	decodedBytes, err := base64.StdEncoding.DecodeString(fileBase64)
	if err != nil {
		return nil, err
	}
	ret := new(GoogleCloudAccountFile)
	if err := json.NewDecoder(bytes.NewBuffer(decodedBytes)).Decode(ret); err != nil {
		return nil, err
	}
	return ret, nil
}
