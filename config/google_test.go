package config

import (
	"path/filepath"
	"testing"

	"github.com/arschles/assert"
	"github.com/deis/k8s-claimer/testutil"
)

func TestGoogleCloudAccountInfo(t *testing.T) {
	fileLoc := filepath.Join(testutil.TestDataDir(), "google_account_info.json")
	f, err := GoogleCloudAccountInfo(fileLoc)
	assert.NoErr(t, err)
	assert.Equal(t, f.PrivateKeyID, "abc", "private key ID")
	assert.Equal(t, f.PrivateKey, "def", "private key")
	assert.Equal(t, f.ClientEmail, "aaron@deis.com", "client email")
	assert.Equal(t, f.ClientID, "aaronschlesinger", "client ID")
}
