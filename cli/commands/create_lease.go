package commands

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/codegangsta/cli"
	"github.com/deis/k8s-claimer/api"
	"github.com/deis/k8s-claimer/htp"
)

const (
	ipEnvVarName          = "IP"
	tokenEnvVarName       = "TOKEN"
	clusterNameEnvVarName = "CLUSTER_NAME"
)

// CreateLease is a cli.Command action for creating a lease
func CreateLease(c *cli.Context) error {
	// inspect env for auth env var
	authToken := os.Getenv("AUTH_TOKEN")
	if authToken == "" {
		log.Fatalf("An authorization token is required in the form of an env var AUTH_TOKEN")
		return errMissingAuthToken
	}
	server := c.GlobalString("server")
	if server == "" {
		log.Fatalf("Server missing")
		return errMissingServer
	}
	durationSec := c.Int("duration")
	if durationSec <= 0 {
		log.Fatalf("Invalid duration %d", durationSec)
	}
	envPrefix := c.String("env-prefix")
	kcfgFile := c.String("kubeconfig-file")
	if len(kcfgFile) < 1 {
		log.Fatalf("Missing kubeconfig-file")
		return errMissingKubeConfigFile
	}

	fd, err := os.Create(kcfgFile)
	if err != nil {
		log.Fatalf("Error opening %s (%s)", kcfgFile, err)
		return err
	}
	defer fd.Close()

	endpt := newEndpoint(htp.Post, server, "lease")
	reqBuf := new(bytes.Buffer)
	req := api.CreateLeaseReq{MaxTimeSec: durationSec}
	if encErr := json.NewEncoder(reqBuf).Encode(req); encErr != nil {
		log.Fatalf("Error encoding request body (%s)", err)
		return err
	}
	res, err := endpt.executeReq(getHTTPClient(), reqBuf, authToken)
	if err != nil {
		log.Fatalf("Error executing %s (%s)", endpt, err)
		return err
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		log.Fatalf("Error executing %s (status code %d)", endpt, res.StatusCode)
		return errInvalidStatusCode{endpoint: endpt.String(), code: res.StatusCode}
	}

	decodedRes, err := api.DecodeCreateLeaseResp(res.Body)
	if err != nil {
		log.Fatalf("Error decoding response (%s)", err)
		return err
	}

	kcfg, err := decodedRes.DecodeKubeConfig()
	if err != nil {
		log.Fatalf("Error decoding kubeconfig (%s)", err)
		return err
	}
	fmt.Println(exportVar(envPrefix, ipEnvVarName, decodedRes.IP))
	fmt.Println(exportVar(envPrefix, tokenEnvVarName, decodedRes.Token))
	fmt.Println(exportVar(envPrefix, clusterNameEnvVarName, decodedRes.ClusterName))

	if _, err := io.Copy(fd, bytes.NewBuffer(kcfg)); err != nil {
		log.Fatalf("Error writing new Kubeconfig file to %s (%s)", kcfgFile, err)
		return err
	}
	return nil
}

func exportVar(prefix, envVarName, val string) string {
	if prefix != "" {
		envVarName = fmt.Sprintf("%s_%s", prefix, envVarName)
	}
	val = strings.Replace(val, `"`, `\"`, -1)
	return fmt.Sprintf(`export %s="%s"`, envVarName, val)
}
