package main

import (
	"os"

	"github.com/codegangsta/cli"
	"github.com/deis/k8s-claimer/cli/commands"
)

func main() {
	app := cli.NewApp()
	app.Name = "k8s-claimer"
	app.Usage = "This CLI can be used against a k8s-claimer server to acquire and release leases"
	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:  "server",
			Value: "",
			Usage: "The k8s-claimer server to talk to",
		},
	}
	app.Commands = []cli.Command{
		cli.Command{
			Name: "lease",
			Subcommands: []cli.Command{
				cli.Command{
					Name: "create",
					Usage: `Creates a new lease and returns 'export' statements to set the lease values as environment variables. Set the 'env-prefix' flag to prefix the environment variable names. If you pass that flag, a '_' character will separate the prefix with the rest of the environment variable name. Below are the basic environment variable names:

- IP - the IP address of the Kubernetes master server
- TOKEN - contains the lease token. Use this when you run 'k8s-claimer-cli lease delete'
- CLUSTER_NAME - contains the name of the cluster. For informational purposes only

The Kubeconfig file will be written to kubeconfig-file
`,
					Action: commands.CreateLease,
					Flags: []cli.Flag{
						cli.IntFlag{
							Name:  "duration",
							Value: 10,
							Usage: "The duration of the lease in seconds",
						},
						cli.StringFlag{
							Name:  "env-prefix",
							Value: "",
							Usage: "The prefix for all environment variables that this command sets",
						},
						cli.StringFlag{
							Name:  "kubeconfig-file",
							Value: "./kubeconfig.yaml",
							Usage: "The location of the resulting Kubeconfig file",
						},
					},
				},
				cli.Command{
					Name:   "delete",
					Action: commands.DeleteLease,
					Usage: `Releases a currently held lease. Pass the lease token as the first and only parameter to this command. For example:

k8s-claimer-cli lease delete $TOKEN
`,
				},
			},
		},
	}
	app.Run(os.Args)
}
