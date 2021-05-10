package main

import (
	"context"
	"fmt"
	"net/http"
	"os"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/urfave/cli/v2"
	"gopkg.in/yaml.v2"

	"github.com/cfunkhouser/lb112x"
	"github.com/cfunkhouser/lb112x/export"
)

var (
	// Version of lb112xutil. Set at build time to something meaningful.
	Version = "development"

	versionMetric = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "lb112x_exporter_version",
		Help: "Version information about this binary",
		ConstLabels: map[string]string{
			"version": Version,
		},
	})

	defaultLB112XDeviceURL    = "http://192.168.5.1"
	defaultPromMetricsAddress = ":9112"
)

func parseFormatter(c *cli.Context) (formatter, error) {
	f := c.String("format")
	switch f {
	case "", "human":
		return human, nil
	}
	return human, fmt.Errorf("unsupported format %q", f)
}

func stat(c *cli.Context) error {
	url := c.String("url")
	password := c.String("password")
	if url == "" || password == "" {
		return cli.Exit("URL and password required", 1)
	}
	client := lb112x.New(url, password)
	ctx := context.Background()
	if err := client.Authenticate(ctx); err != nil {
		return cli.Exit(err, 1)
	}
	m, err := client.Poll(ctx)
	if err != nil {
		return cli.Exit(err, 1)
	}
	format, err := parseFormatter(c)
	if err != nil {
		return cli.Exit(err, 1)
	}
	format(os.Stdout, m)
	return nil
}

func loadConfig(c *cli.Context) (*export.Config, error) {
	f, err := os.Open(c.String("config"))
	if err != nil {
		return nil, err
	}
	var cfg export.Config
	if err := yaml.NewDecoder(f).Decode(&cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}

func serveExporter(c *cli.Context) error {
	r := prometheus.NewRegistry()
	if err := r.Register(versionMetric); err != nil {
		return err
	}
	cfg, err := loadConfig(c)
	if err != nil {
		return cli.Exit(err, 1)
	}
	exporter, err := export.New(cfg)
	if err != nil {
		return cli.Exit(err, 1)
	}

	versionMetric.Set(1.0)
	http.Handle("/metrics", promhttp.HandlerFor(r, promhttp.HandlerOpts{}))
	http.Handle("/scrape", exporter)
	return http.ListenAndServe(c.String("listen"), nil)
}

func main() {
	app := &cli.App{
		Name:    "lb112xutil",
		Usage:   "Utility for woring with Netgear LB112X LTE Modems",
		Version: Version,
		Commands: []*cli.Command{
			{
				Name: "status",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Required:   true,
						Value:      defaultLB112XDeviceURL,
						Name:       "url",
						Usage:      "URL at which the modem API is found.",
						HasBeenSet: true,
						EnvVars:    []string{"LB112X_URL"},
					},
					&cli.StringFlag{
						Required:    true,
						Name:        "password",
						Usage:       "Admin password for the modem web API.",
						DefaultText: "HIDDEN",
						EnvVars:     []string{"LB112X_ADMIN_PASSWORD"},
					},
				},
				Aliases: []string{"stat"},
				Usage:   "Display status of a LB112X device.",
				Action:  stat,
			},
			{
				Name: "export",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:     "config",
						Aliases:  []string{"f"},
						Usage:    "Configuration file for Prometheus Exporter",
						Required: true,
					},
					&cli.StringFlag{
						Name:       "listen",
						Usage:      "Local ip:port from which to serve Prometheus metrics.",
						Aliases:    []string{"L"},
						Value:      defaultPromMetricsAddress,
						HasBeenSet: true,
					},
				},
				Usage:  "Export LB112x device metrics to Prometheus. Blocks until killed.",
				Action: serveExporter,
			},
			{
				Name: "testcfg",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:     "config",
						Aliases:  []string{"f"},
						Usage:    "Configuration file for Prometheus Exporter",
						Required: true,
					},
				},
				Usage: "Test Prometheus exporter configuration.",
				Action: func(c *cli.Context) error {
					if _, err := loadConfig(c); err != nil {
						return err
					}
					fmt.Println("OK")
					return nil
				},
			},
		},
	}

	if err := app.Run(os.Args); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
