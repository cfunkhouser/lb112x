package export

import (
	"context"
	"fmt"
	"net/http"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	"github.com/cfunkhouser/lb112x"
)

type deviceMetrics struct {
	temperature   prometheus.Gauge
	tempCritical  prometheus.Gauge
	resetRequired prometheus.Gauge
	rssi          prometheus.Gauge
	numBars       prometheus.Gauge
	info          *prometheus.GaugeVec
	netInfo       *prometheus.GaugeVec
}

func (m *deviceMetrics) register(r prometheus.Registerer) error {
	if err := r.Register(m.temperature); err != nil {
		return err
	}
	if err := r.Register(m.tempCritical); err != nil {
		return err
	}
	if err := r.Register(m.resetRequired); err != nil {
		return err
	}
	if err := r.Register(m.rssi); err != nil {
		return err
	}
	if err := r.Register(m.numBars); err != nil {
		return err
	}
	if err := r.Register(m.netInfo); err != nil {
		return err
	}
	return nil
}

type deviceExporter struct {
	client   *lb112x.Client
	metrics  deviceMetrics
	registry *prometheus.Registry
}

func (e *deviceExporter) update(ctx context.Context) error {
	if err := e.client.Authenticate(ctx); err != nil {
		return err
	}
	details, err := e.client.Poll(ctx)
	if err != nil {
		return err
	}

	var criticality float64
	if details.Power.DeviceTempCritical {
		criticality = 1.0
	}

	e.metrics.temperature.Set(float64(details.General.Temperature))
	e.metrics.tempCritical.Set(criticality)
	// TODO(cfunkhouser): Parse resetRequired
	e.metrics.rssi.Set(float64(details.WWAN.SignalStrenth.RSSI))
	e.metrics.numBars.Set(float64(details.WWAN.SignalStrenth.Bars))
	e.metrics.info.With(prometheus.Labels{
		"model": details.General.Model,
		"imei":  details.General.IMEI,
		"hw":    details.General.HardwareVersion,
		"fw":    details.General.FirmwareVersion,
		"sw":    details.General.AppVersion,
	}).Set(1.0)
	e.metrics.netInfo.With(prometheus.Labels{
		"network": details.WWAN.RegisterNetworkDisplay,
		"ipv4":    details.WWAN.IP,
		"ipv6":    details.WWAN.IPv6,
	}).Set(1.0)

	return nil
}

func newDeviceExporter(client *lb112x.Client) (*deviceExporter, error) {
	de := &deviceExporter{
		client: client,
		metrics: deviceMetrics{
			temperature: prometheus.NewGauge(
				prometheus.GaugeOpts{
					Name: "lb112x_temperature",
					Help: "Temperature of the LB112x Device in degrees Celcius",
				},
			),
			tempCritical: prometheus.NewGauge(
				prometheus.GaugeOpts{
					Name: "lb112x_temperature_critical",
					Help: "Criticality of the LB112x Device's temperature.",
				},
			),
			resetRequired: prometheus.NewGauge(
				prometheus.GaugeOpts{
					Name: "lb112x_requires_reset",
					Help: "Whether the LB112x Device requires a reset.",
				},
			),
			rssi: prometheus.NewGauge(
				prometheus.GaugeOpts{
					Name: "lb112x_rssi",
					Help: "Strength in RSSI of the LB112x Device's cellular signal.",
				},
			),
			numBars: prometheus.NewGauge(
				prometheus.GaugeOpts{
					Name: "lb112x_num_bars",
					Help: "Strength in number of bars of the LB112x Device's cellular signal.",
				},
			),
			info: prometheus.NewGaugeVec(
				prometheus.GaugeOpts{
					Name: "lb112x_device_info",
					Help: "Information describing the LB112x Device.",
				},
				[]string{"model", "imei", "hw", "fw", "sw"},
			),
			netInfo: prometheus.NewGaugeVec(
				prometheus.GaugeOpts{
					Name: "lb112x_wwan_info",
					Help: "Information describing the LB112x Device's WWAN interface.",
				},
				[]string{"network", "ipv4", "ipv6"},
			),
		},

		registry: prometheus.NewRegistry(),
	}
	if err := de.metrics.register(de.registry); err != nil {
		return nil, err
	}
	return de, nil
}

type Handler struct {
	exporters map[string]*deviceExporter
}

var unconfiguredTargetTmpl = `Could not find target: %q

Was it configured?

This exporter cannot dynamically discover targets, since credentials are
required for each target. Add the target to the configuration file to
monitor it.
`

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	target := r.URL.Query().Get("target")
	if target == "" {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(w, "No target specified, not sure what you want.")
		return
	}

	de, has := h.exporters[target]
	if !has {
		w.WriteHeader(http.StatusNotFound)
		fmt.Fprintf(w, unconfiguredTargetTmpl, target)
		return
	}

	if err := de.update(r.Context()); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, "Failed polling LB112x device: %v", err)
		return
	}
	promhttp.HandlerFor(de.registry, promhttp.HandlerOpts{}).ServeHTTP(w, r)
}

func New(config *Config, opts ...lb112x.Option) (*Handler, error) {
	h := &Handler{
		exporters: make(map[string]*deviceExporter),
	}
	if ttl := config.Global.ClientTimeout; ttl != nil {
		opts = append(opts, lb112x.WithClientTimeout(*ttl))
	}
	for _, device := range config.Devices {
		dopts := opts[:]
		if ttl := config.Global.ClientTimeout; ttl != nil {
			dopts = append(dopts, lb112x.WithClientTimeout(*ttl))
		}
		de, err := newDeviceExporter(lb112x.New(device.URL, device.Password, dopts...))
		if err != nil {
			return nil, err
		}
		h.exporters[device.URL] = de
	}
	return h, nil
}
