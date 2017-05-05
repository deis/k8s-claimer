package config

// Azure contains the necessary configuration to talk to the Azure Cloud API
type Azure struct {
	ClientID       string `envconfig:"AZURE_CLIENT_ID"`
	ClientSecret   string `envconfig:"AZURE_CLIENT_SECRET"`
	TenantID       string `envconfig:"AZURE_TENANT_ID"`
	SubscriptionID string `envconfig:"AZURE_SUBSCRIPTION_ID"`
}

// ToMap returns an azure config object as a map[string]string
func (a Azure) ToMap() map[string]string {
	return map[string]string{
		"AZURE_CLIENT_ID":       a.ClientID,
		"AZURE_CLIENT_SECRET":   a.ClientSecret,
		"AZURE_SUBSCRIPTION_ID": a.SubscriptionID,
		"AZURE_TENANT_ID":       a.TenantID}
}
