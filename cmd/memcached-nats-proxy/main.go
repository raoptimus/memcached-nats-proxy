package main

import (
	"fmt"
	"github.com/raoptimus/memcached-nats-proxy/src/proxy"
	"os"

	log "github.com/sirupsen/logrus"
	"github.com/urfave/cli"
)

var (
	Version   string
	GitCommit string
)

const version = "1.1.0"

func init() {
	log.SetFormatter(&log.TextFormatter{})
}
func main() {
	app := cli.NewApp()
	app.Name = "NATS memcached proxy"
	app.Usage = "."
	app.Version = fmt.Sprintf("v%s.rev[%s]", Version, GitCommit)
	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:   "nats-url",
			EnvVar: "NATS_URL",
			Value:  "nats://127.0.0.1:4222",
			Usage:  "The url can contain username/password semantics. e.g. nats://username:pass@localhost:4222.",
		},
		cli.BoolFlag{
			Name:   "nats-publish-async",
			EnvVar: "NATS_PUBLISH_ASYNC",
			Usage:  "Publish message to the cluster asynchronously.",
		},
		cli.StringFlag{
			Name:   "log-level",
			EnvVar: "LOG_LEVEL",
			Value:  "info",
			Usage:  "",
		},
		cli.StringFlag{
			Name:   "server-addr",
			EnvVar: "SERVER_ADDR",
			Value:  ":11211",
			Usage:  "Listen address.",
		},
		cli.StringFlag{
			Name:   "metrics-addr",
			EnvVar: "METRICS_ADDR",
			Value:  ":1414",
			Usage:  "Prometheus metrics HTTP endpoint.",
		},
		cli.BoolFlag{
			Name:   "debug",
			EnvVar: "DEBUG",
			Usage:  "Enable debug mode.",
		},
	}
	app.Action = func(c *cli.Context) error {
		switch {
		case c.Bool("debug"):
			log.SetLevel(log.DebugLevel)
		default:
			if level, err := log.ParseLevel(c.String("log-level")); err == nil {
				log.SetLevel(level)
			}
		}
		p, err := proxy.New(version, proxy.Options{
			NatsURL:          c.String("nats-url"),
			NatsPublishAsync: c.Bool("nats-publish-async"),
			MetricsAddr:      c.String("metrics-addr"),
			ServerAddr:       c.String("server-addr"),
		})
		if err != nil {
			return err
		}
		return p.Listen()
	}
	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}
