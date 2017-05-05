package azure

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/Azure/go-autorest/autorest"
	"github.com/Azure/go-autorest/autorest/adal"
	"github.com/Azure/go-autorest/autorest/azure"
)

const (
	credentialsPath = "/.azure/credentials.json"
)

// ToJSON returns the passed item as a pretty-printed JSON string. If any JSON error occurs,
// it returns the empty string.
func ToJSON(v interface{}) (string, error) {
	j, err := json.MarshalIndent(v, "", "  ")
	return string(j), err
}

// NewBearerAuthorizer creates a new BearerAuthorizer using values of the passed credentials map.
func NewBearerAuthorizer(c map[string]string, scope string) (*autorest.BearerAuthorizer, error) {
	oauthConfig, err := adal.NewOAuthConfig(azure.PublicCloud.ActiveDirectoryEndpoint, c["AZURE_TENANT_ID"])
	if err != nil {
		log.Printf("Error trying to create new OAuth Config:%s\n", err)
		return nil, err
	}
	token, err := adal.NewServicePrincipalToken(*oauthConfig, c["AZURE_CLIENT_ID"], c["AZURE_CLIENT_SECRET"], scope)
	if err != nil {
		log.Printf("Error trying to create New Service Principal Token:%s", err)
		return nil, err
	}
	return autorest.NewBearerAuthorizer(token), nil
}

func ensureValueStrings(mapOfInterface map[string]interface{}) map[string]string {
	mapOfStrings := make(map[string]string)
	for key, value := range mapOfInterface {
		mapOfStrings[key] = ensureValueString(value)
	}
	return mapOfStrings
}

func ensureValueString(value interface{}) string {
	if value == nil {
		return ""
	}
	switch v := value.(type) {
	case string:
		return v
	default:
		return fmt.Sprintf("%v", v)
	}
}
