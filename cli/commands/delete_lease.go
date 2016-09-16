package commands

import (
	"fmt"
	"log"
	"os"

	"github.com/codegangsta/cli"
	"github.com/deis/k8s-claimer/client"
)

// DeleteLease is a cli.Command action for deleting a lease
func DeleteLease(c *cli.Context) error {
	// inspect env for auth env var
	authToken := os.Getenv("AUTH_TOKEN")
	if authToken == "" {
		log.Fatalf("An authorization token is required in the form of an env var AUTH_TOKEN")
	}
	server := c.GlobalString("server")
	if server == "" {
		log.Fatalf("Server missing")
	}
	if len(c.Args()) < 1 {
		log.Fatalf("Lease token missing")
	}
	leaseToken := c.Args()[0]

	if err := client.DeleteLease(server, authToken, leaseToken); err != nil {
		log.Fatalf("Error deleting lease: %s", err)
	}

	fmt.Println("Deleted lease", leaseToken)

	return nil
}
