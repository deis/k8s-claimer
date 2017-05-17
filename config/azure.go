package config

// Azure contains the necessary configuration to talk to the Azure Cloud API
type Azure struct {
	ClientID       string `envconfig:"AZURE_CLIENT_ID"`
	ClientSecret   string `envconfig:"AZURE_CLIENT_SECRET"`
	TenantID       string `envconfig:"AZURE_TENANT_ID"`
	SubscriptionID string `envconfig:"AZURE_SUBSCRIPTION_ID"`
}

//ValidConfig will return true if there are values set for each Property of the Azure config object
func (a *Azure) ValidConfig() bool {
	return a.SubscriptionID != "" && a.ClientID != "" && a.ClientSecret != "" && a.TenantID != ""
}
