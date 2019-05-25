package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"html/template"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/prometheus/client_golang/prometheus"
)

var (
	index = template.Must(template.New("index").Parse(
		`<!doctype html>
	 <title>RTL_433 Prometheus Exporter</title>
	 <h1>RTL_433 Prometheus Exporter</h1>
	 <a href="/metrics">Metrics</a>`))

	addr = flag.String("listen", ":9001", "Address to listen on")

	labels = []string{"model", "id", "channel"}

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
	ID int `json:"id"`
	// Channel sensor is transmitting on. Typically 1-3, controlled by a switch on the device
	Channel int `json:"channel"`
	// Battery status, typically "LOW" or "OK", case-insensitive.
	Battery string `json:"battery"`
	// Temperature in Celsius. Nil if not present in initial JSON.
	Temperature *float64 `json:"temperature_C"`
	// Humidity (0-100). Nil if not present in initial JSON.
	Humidity *int32 `json:"humidity"`
}

func main() {
	flag.Parse()

	prometheus.MustRegister(packetsReceived)
	prometheus.MustRegister(temperature)
	prometheus.MustRegister(humidity)
	prometheus.MustRegister(timestamp)
	prometheus.MustRegister(battery)

	go func() {
		http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "text/html; charset=utf-8")
			index.Execute(w, "")
		})
		http.Handle("/metrics", prometheus.Handler())
		if err := http.ListenAndServe(*addr, nil); err != nil {
			panic(err)
		}
	}()

	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		msg := Message{}
		if err := json.Unmarshal(scanner.Bytes(), &msg); err != nil {
			panic(err)
		}

		labels := []string{msg.Model, strconv.Itoa(msg.ID), strconv.Itoa(msg.Channel)}
		packetsReceived.WithLabelValues(labels...).Inc()
		timestamp.WithLabelValues(labels...).SetToCurrentTime()
		if temperature != nil {
			temperature.WithLabelValues(labels...).Set(*msg.Temperature)
		}
		if msg.Humidity != nil {
			humidity.WithLabelValues(labels...).Set(float64(*msg.Humidity) / 100)
		}
		switch {
		case strings.EqualFold(msg.Battery, "OK"):
			battery.WithLabelValues(labels...).Set(1)
		case strings.EqualFold(msg.Battery, "LOW"):
			battery.WithLabelValues(labels...).Set(0)
		}
	}

	if err := scanner.Err(); err != nil {
		panic(err)
	}
}
