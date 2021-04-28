package main

import (
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/urfave/cli/v2"

	"github.com/cfunkhouser/lb112x"
)

func main() {
	app := &cli.App{
		Name:  "lb112xutil",
		Usage: "Utility for woring with Netgear LB112X LTE Modems",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Required:   true,
				Value:      "http://192.168.5.1",
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
		Commands: []*cli.Command{
			{
				Name:    "status",
				Aliases: []string{"stat"},
				Usage:   "Display status of a LB112X device.",
				Action: func(c *cli.Context) error {
					url := c.String("url")
					password := c.String("password")
					if url == "" || password == "" {
						return cli.Exit("URL and password required", 1)
					}
					client := lb112x.New(url, password)
					if err := client.Authenticate(); err != nil {
						return cli.Exit(err, 1)
					}
					m, err := client.Poll()
					if err != nil {
						return cli.Exit(err, 1)
					}
					w := tabwriter.NewWriter(os.Stdout, 0, 0, 1, ' ', tabwriter.DiscardEmptyColumns)
					fmt.Fprintln(w, "Device\t")
					fmt.Fprintf(w, " Model:\t%v\n", m.General.Model)
					fmt.Fprintf(w, " IMEI:\t%v\n", m.General.IMEI)
					fmt.Fprintln(w, "Network\t")
					fmt.Fprintf(w, " Name:\t%v\n", m.WWAN.RegisterNetworkDisplay)
					fmt.Fprintf(w, " IP:\t%v\n", m.WWAN.IP)
					fmt.Fprintf(w, " IPv6:\t%v\n", m.WWAN.IPv6)
					fmt.Fprintln(w, "Signal\t")
					fmt.Fprintf(w, " Bars:\t%v\n", m.WWAN.SignalStrenth.Bars)
					fmt.Fprintf(w, " RSSI:\t%v\n", m.WWAN.SignalStrenth.RSSI)
					fmt.Fprintf(w, " Temperature:\t%v\u00b0C\n", m.General.Temperature)
					fmt.Fprintf(w, " Temp Critical:\t%v\n", m.Power.DeviceTempCritical)
					w.Flush()
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
