package commands

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/codegangsta/cli"
	"github.com/deis/k8s-claimer/htp"
)

// DeleteLease is a cli.Command action for deleting a lease
func DeleteLease(c *cli.Context) {
	server := c.GlobalString("server")
	if server == "" {
		log.Fatalf("Server missing")
	}
	if len(c.Args()) < 1 {
		log.Fatalf("Lease token missing")
	}
	leaseToken := c.Args()[0]
	endpt := newEndpoint(htp.Delete, server, "lease/"+leaseToken)
	resp, err := endpt.executeReq(getHTTPClient(), nil)
	if err != nil {
		log.Fatalf("Error executing %s (%s)", endpt, err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		bodyBytes, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			bodyBytes = nil
		}
		log.Fatalf("Error deleting. Code: %d, Body: %s", resp.StatusCode, string(bodyBytes))
	}
	fmt.Println("Deleted lease", leaseToken)
}
