package main

import (
	"os"

	"github.com/Sirupsen/logrus"
	"github.com/codegangsta/cli"
	"github.com/rancher/kubernetes-agent/config"
	"github.com/rancher/kubernetes-agent/kubernetesclient"
	"github.com/rancher/kubernetes-agent/kubernetesevents"
	"github.com/rancher/servicealias/handlers"
)

var VERSION = "v0.0.0-dev"

func main() {
	app := cli.NewApp()
	app.Name = "service-aliaser"
	app.Usage = "Start the Service aliaser"
	app.Action = launch

	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:   "kubernetes-url",
			Value:  "http://10.43.0.1:6443",
			Usage:  "URL for kubernetes API",
			EnvVar: "KUBERNETES_URL",
		},
		cli.StringSliceFlag{
			Name:  "watch-kind",
			Value: &cli.StringSlice{"servicealiases"},
			Usage: "Which k8s kinds to watch to update service alias",
		},
		cli.StringFlag{
			Name:   "cattle-access-key",
			Usage:  "Cattle API Access Key",
			EnvVar: "CATTLE_ACCESS_KEY",
		},
		cli.StringFlag{
			Name:   "cattle-secret-key",
			Usage:  "Cattle API Secret Key",
			EnvVar: "CATTLE_SECRET_KEY",
		},
	}

	app.Run(os.Args)
}

func launch(c *cli.Context) {
	conf := config.Conf(c)

	resultChan := make(chan error)

	if err := kubernetesclient.Init(); err != nil {
		logrus.Fatal(err)
	}
	client := kubernetesclient.NewClient(conf.KubernetesURL, true)
	hs := []kubernetesevents.Handler{}

	logrus.Info("Watching changes for kinds: ", c.StringSlice("watch-kind"))
	for _, kind := range c.StringSlice("watch-kind") {
		hs = append(hs, handlers.NewHandler(client, kind))
	}

	go func(rc chan error) {
		err := kubernetesevents.ConnectToEventStream(hs, conf)
		logrus.Errorf("Kubernetes stream listener exited with error: %s", err)
		rc <- err
	}(resultChan)

	<-resultChan
	logrus.Info("Exiting.")
}
