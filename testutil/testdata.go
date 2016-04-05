package testutil

import (
	"log"
	"os"
	"path/filepath"
)

var goPath string

func init() {
	goPath = os.Getenv("GOPATH")
	if goPath == "" {
		log.Fatalf("GOPATH not set")
	}
}

// TestDataDir returns the fully qualified path to the testdata/ directory
func TestDataDir() string {
	return filepath.Join(goPath, "src", "github.com", "deis", "k8s-claimer", "testdata")
}
