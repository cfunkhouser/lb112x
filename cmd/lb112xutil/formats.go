package main

import (
	"fmt"
	"io"
	"os"
	"text/tabwriter"

	"github.com/cfunkhouser/lb112x"
)

type formatter func(io.Writer, *lb112x.APIModel)

func human(out io.Writer, m *lb112x.APIModel) {
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
}
