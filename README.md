# LB112X Utility

Library and command line utility for interacting with Netgear LB112X LTE Modems.

Note: Tested only with LB1120.

## CLI Utility

View the usage by running the utility with no arguments.

```console
$ lb112xutil
NAME:
   lb112xutil - Utility for woring with Netgear LB112X LTE Modems

USAGE:
   lb112xutil [global options] command [command options] [arguments...]

COMMANDS:
   status, stat  Display status of a LB112X device.
   help, h       Shows a list of commands or help for one command

GLOBAL OPTIONS:
   --url value       URL at which the modem API is found. (default: "http://192.168.5.1") [$LB112X_URL]
   --password value  Admin password for the modem web API. (default: HIDDEN) [$LB112X_ADMIN_PASSWORD]
   --help, -h        show help (default: false)
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