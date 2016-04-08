package config

import (
	"encoding/base64"
	"io/ioutil"
	"path/filepath"
	"testing"

	"github.com/arschles/assert"
	"github.com/deis/k8s-claimer/testutil"
)

func TestGoogleCloudAccountInfo(t *testing.T) {
	fileLoc := filepath.Join(testutil.TestDataDir(), "google_account_info.json")
	fileBytes, err := ioutil.ReadFile(fileLoc)
	assert.NoErr(t, err)
	encoded := base64.StdEncoding.EncodeToString(fileBytes)
	f, err := GoogleCloudAccountInfo(encoded)
	assert.NoErr(t, err)
	assert.Equal(t, f.PrivateKeyID, "abc", "private key ID")
	assert.Equal(t, f.PrivateKey, "def", "private key")
	assert.Equal(t, f.ClientEmail, "aaron@deis.com", "client email")
	assert.Equal(t, f.ClientID, "aaronschlesinger", "client ID")
}
