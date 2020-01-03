package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"html/template"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"strings"

	"github.com/prometheus/client_golang/prometheus"
)

var (
	index = template.Must(template.New("index").Parse(
		`<!doctype html>
	 <title>RTL_433 Prometheus Exporter</title>
	 <h1>RTL_433 Prometheus Exporter</h1>
	 <a href="/metrics">Metrics</a>
	 <p>
	 <pre>{{.}}</pre>
	 `))

	addr       = flag.String("listen", ":9001", "Address to listen on")
	subprocess = flag.String("subprocess", "rtl_433 -F json", "What command to run to get rtl_433 radio packets")
	matchers   = make(locationMatchers)

	labels = []string{"model", "id", "channel", "location"}

	packetsReceived = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "rtl_433_packets_received",
			Help: "Packets (temperature messages) received.",
		},
		labels,
	)
	temperature = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "rtl_433_temperature_celsius",
			Help: "Temperature in Celsius",
		},
		labels,
	)
	humidity = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "rtl_433_humidity",
			Help: "Relative Humidity (0-1.0)",
		},
		labels,
	)
	timestamp = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "rtl_433_timestamp_seconds",
			Help: "Timestamp we received the message (Unix seconds)",
		},
		labels,
	)
	battery = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "rtl_433_battery",
			Help: "Battery high (1) or low (0).",
		},
		labels,
	)
)

// Message is a single sensor observation: a single line of JSON input from $ rtl_433 -F json
type Message struct {
	// ISO 8601 Datetime e.g. "2019-05-23 20:41:45"
	Time string `json:"time"`
	// Sensor Model
	Model string `json:"model"`
	// Sensor ID. May be random per-boot, or saved into device memory.
	RawID interface{} `json:"id"`
	// Channel sensor is transmitting on. Typically 1-3, controlled by a switch on the device
	// Either an int or string
	RawChannel interface{} `json:"channel"`
	// Battery status, typically "LOW" or "OK" or "", case-insensitive.
	Battery string `json:"battery"`
	// Alternative battery key. 1 or 0 or nil (not present)
	BatteryOK *int `json:"battery_ok"`
	// Yet another alternative battery key. 1 for low battery, 0 for high battery, nil (not present)
	BatteryLow *int `json:"battery_low"`
	// Temperature in Celsius. Nil if not present in initial JSON.
	Temperature *float64 `json:"temperature_C"`
	// Humidity (0-100). Nil if not present in initial JSON.
	Humidity *int32 `json:"humidity"`
}

type locationMatcher struct {
	Model   string
	ID 		string
}

type locationMatchers map[locationMatcher]string

func (lms locationMatchers) String() string {
	out := []string{}
	for matcher, location := range lms {
		out = append(out, matcher.Model+","+matcher.ID+","+location)
	}
	return strings.Join(out, ";")
}

func (lms locationMatchers) Set(v string) error {
	matchers := strings.Split(v, ";")
	for _, m := range matchers {
		f := strings.Split(m, ",")
		lms[locationMatcher{Model: f[0], ID: f[1]}] = f[2]
	}
	return nil
}

// Channel returns a string representation of the channel
// Some sensors output numbered channels, some output string channels.
// We have to handle both.
func (m *Message) Channel() (string, error) {
	if s, ok := m.RawChannel.(string); ok {
		return s, nil
	}
	if f, ok := m.RawChannel.(float64); ok {
		return fmt.Sprintf("%d", int(f)), nil
	}
	return "", fmt.Errorf("Could not parse JSON, bad channel (expected float or string), got: %v", m.RawChannel)
}

func (m *Message) ID() (string, error) {
	if s, ok := m.RawID.(string); ok {
		return s, nil
	}
	if f, ok := m.RawID.(float64); ok {
		return fmt.Sprintf("%d", int(f)), nil
	}
	return "", fmt.Errorf("Could not parse JSON, bad ID (expected float or string), got: %v", m.RawID)
}

func run(r io.Reader) error {
	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		msg := Message{}
		line := scanner.Bytes()
		if err := json.Unmarshal(line, &msg); err != nil {
			log.Fatal(err)
		}

		channel, err := msg.Channel()
		if err != nil {
			log.Fatal(err)
		}

		id, err := msg.ID()
		if err != nil {
			log.Fatal(err)
		}

		location := matchers[locationMatcher{Model: msg.Model, ID: id}]

		labels := []string{msg.Model, id, channel, location}
		packetsReceived.WithLabelValues(labels...).Inc()
		timestamp.WithLabelValues(labels...).SetToCurrentTime()
		if temperature != nil {
			temperature.WithLabelValues(labels...).Set(*msg.Temperature)
		}
		if msg.Humidity != nil {
			humidity.WithLabelValues(labels...).Set(float64(*msg.Humidity) / 100)
		}
		if msg.Battery != "" {
			switch {
			case strings.EqualFold(msg.Battery, "OK"):
				battery.WithLabelValues(labels...).Set(1)
			case strings.EqualFold(msg.Battery, "LOW"):
				battery.WithLabelValues(labels...).Set(0)
			}
		} else if msg.BatteryOK != nil {
			battery.WithLabelValues(labels...).Set(float64(*msg.BatteryOK))
		} else if msg.BatteryLow != nil {
			battery.WithLabelValues(labels...).Set(float64(1 - *msg.BatteryLow))
		}
	}
	return scanner.Err()
}

func main() {
	flag.Var(&matchers, "location_matchers", "Acurite-Tower,13732,Kitchen;...")
	flag.Parse()
	log.Print("Flag config: " + matchers.String())
	prometheus.MustRegister(packetsReceived, temperature, humidity, timestamp, battery)

	go func() {
		http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "text/html; charset=utf-8")
			index.Execute(w, matchers)
		})
		http.Handle("/metrics", prometheus.Handler())
		if err := http.ListenAndServe(*addr, nil); err != nil {
			log.Fatal(err)
		}
	}()

	cmd := exec.Command("/bin/bash", "-c", *subprocess)
	// If we don't tell the subprocess stderr to be our stderr, we get no logs on failure.
	cmd.Stderr = os.Stderr
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		log.Fatal(err)
	}
	if err := cmd.Start(); err != nil {
		log.Fatal(err)
	}
	scannerErr := run(stdout)
	// Wait first, then check scanner.Err, because Wait's error messages are better.
	if err := cmd.Wait(); err != nil {
		log.Fatal(err)
	}
	if scannerErr != nil {
		log.Fatal(err)
	}
}
