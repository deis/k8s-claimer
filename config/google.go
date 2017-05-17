package config

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
)

// Google contains the Google cloud related configuration, including credentials and
// project info
type Google struct {
	AccountFileJSON string `envconfig:"GOOGLE_CLOUD_ACCOUNT_FILE"`
	ProjectID       string `envconfig:"GOOGLE_CLOUD_PROJECT_ID"`
	Zone            string `envconfig:"GOOGLE_CLOUD_ZONE" default:"-"`
	AccountFile     AccountFile
}

// AccountFile represents the structure of the account file JSON file.
// this struct was adapted from the Terraform project
// (https://github.com/hashicorp/terraform/blob/master/builtin/providers/google/config.go)
type AccountFile struct {
	// PrivateKeyID is the private key ID given in the account file
	PrivateKeyID string `json:"private_key_id"`
	// PrivateKey is the private key given in the account file
	PrivateKey string `json:"private_key"`
	// ClientEmail is the client email given in the account file
	ClientEmail string `json:"client_email"`
	// ClientID is the client ID given in the account file
	ClientID string `json:"client_id"`
}

// AccountInfo parses the file at fileLoc into a GoogleCloudAccountFile and returns
// it. Returns nil and an appropriate error if any error occurred in reading or decoding.
func AccountInfo(fileBase64 string) (*AccountFile, error) {
	decodedBytes, err := base64.StdEncoding.DecodeString(fileBase64)
	if err != nil {
		return nil, err
	}
	ret := new(AccountFile)
	if err := json.NewDecoder(bytes.NewBuffer(decodedBytes)).Decode(ret); err != nil {
		return nil, err
	}
	return ret, nil
}

//ValidConfig will return true if there are values set for each Property of the Google config object
func (g *Google) ValidConfig() bool {
	return g.AccountFileJSON != "" && g.ProjectID != "" && g.Zone != ""
}
