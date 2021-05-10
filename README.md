# LB112X Utility

Library and command line utility for interacting with Netgear LB112X LTE Modems.

*Tested only with LB1120!*

## CLI Utility

View the usage by running the utility with no arguments.

```console
$ lb112xutil
NAME:
   lb112xutil - Utility for woring with Netgear LB112X LTE Modems

USAGE:
   lb112xutil [global options] command [command options] [arguments...]

VERSION:
   development

COMMANDS:
   status, stat  Display status of a LB112X device.
   export        Export LB112x device metrics to Prometheus. Blocks until killed.
   testcfg       Test Prometheus exporter configuration.
   help, h       Shows a list of commands or help for one command

GLOBAL OPTIONS:
   --help, -h     show help (default: false)
   --version, -v  print the version (default: false)
```

Use the `status` command to get the current status of the device. The admin
password can be provided as an environment variable to keep it secret.

```console
$ export LB112X_ADMIN_PASSWORD=ThisIsAPassword
$ lb112xutil status
Device
 Model:         LB1120
 IMEI:          123456789101112
Network
 Name:          AT&T
 IP:            10.180.2.2
 IPv6:          2600:0000:1111:2222:3333:4444:5555:6666
Signal
 Bars:          3
 RSSI:          -75
 Temperature:   30°C
 Temp Critical: false
```

## Prometheus Exporter

The CLI includes a Prometheus exporter for LB112x devices. Since each device
requires authentication credentials, the exporter must me configured ahead of
time with the devices it will export. This can be done with a YAML file which
might look like:

```yaml
---
global:
  timeout: 10s
devices:
  - url: http://192.168.5.1
    password: SuperSecretPassword
  - url: http://192.168.6.1
    password: CantGuessMe
    timeout: 15s
```

See [`export/config.go`](export/config.go) for the exact fields. The exporter
itself can be started using the `lb112xutil export` command. The target key for
scraping is the URL value for each device.

```console
$ lb112xutil export
NAME:
   lb112xutil export - Export LB112x device metrics to Prometheus. Blocks until killed.

USAGE:
   lb112xutil export [command options] [arguments...]

OPTIONS:
   --config value, -f value  Configuration file for Prometheus Exporter
   --listen value, -L value  Local ip:port from which to serve Prometheus metrics. (default: ":9112")
   --help, -h                show help (default: false)
```

Once running, Prometheus can be configured to scrape the LB112x devices with the
following static config, using relabels:

```yaml
scrape_configs:
  - job_name: lb112x
    metrics_path: /scrape
    static_configs:
      - targets:
          - http://192.168.5.1
          - http://192.168.6.1
    relabel_configs:
      - source_labels: [__address__]
        target_label: __param_target
      - source_labels: [__param_target]
        target_label: instance
      - target_label: __address__
        # The host:port of the exporter, started above.
        replacement: 127.0.0.1:9112
```