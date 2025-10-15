package main

import (
	_ "embed"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

//go:embed rootPage.html
var rootPageHtml []byte

var errInvalidResponse = errors.New("unexpected response from the server")

type climateValues struct {
	TempC float64 `json:"tempC"`
	TempF float64 `json:"tempF"`
	RH    float64 `json:"rh"`
}

func (tv *climateValues) getTempValues(client *http.Client, url string) error {
	response, err := client.Get(url)
	if err != nil {
		log.Println(err)
		return err
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		err := fmt.Errorf(
			"%w: invalid status code: %s",
			errInvalidResponse,
			response.Status,
		)
		log.Println(err)
		return err
	}

	if err := json.NewDecoder(response.Body).Decode(tv); err != nil {
		log.Println(err)
		return err
	}
	return nil
}

// metrics type is used to cache data from the Pico W
// it is a concurrency-safe model to access data in case more than one instance of Prometheus
// requests it simultaneously
// Caches the data for two seconds, preventing overloading the Pico W in case the exporter receives
// many concurrent requests
type metrics struct {
	results      *climateValues
	up           float64 // I'm guessing this is a float64 and not a bool b/c of some Prometheus reason
	expire       time.Time
	sync.RWMutex // embedded!
}

func (m *metrics) getMetrics(client *http.Client, url string) *metrics {
	m.Lock()
	defer m.Unlock()

	if time.Now().Before(m.expire) {
		return m
	}

	m.up = 1
	if err := m.results.getTempValues(client, url); err != nil {
		m.up = 0
		m.results.TempC = 0
		m.results.TempF = 0
	}

	m.expire = time.Now().Add(2 * time.Second)
	return m
}

func (m *metrics) tempC() float64 {
	m.RLock()
	defer m.RUnlock()

	return m.results.TempC

}

func (m *metrics) tempF() float64 {
	m.RLock()
	defer m.RUnlock()

	return m.results.TempF
}

func (m *metrics) RH() float64 {
	m.RLock()
	defer m.RUnlock()

	return m.results.RH
}

func (m *metrics) status() float64 {
	m.RLock()
	defer m.RUnlock()
	return m.up
}

func newMux(url string, picoName string) http.Handler {
	mux := http.NewServeMux()

	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	m := &metrics{
		results: &climateValues{},
	}

	promauto.NewGaugeFunc(
		prometheus.GaugeOpts{
			Name:        fmt.Sprintf("%s_pico_temperature", picoName),
			Help:        "BME280 Sensor Temperature.",
			ConstLabels: prometheus.Labels{"unit": "celsius"},
		},
		func() float64 {
			return m.getMetrics(client, url).tempC()
		})

	promauto.NewGaugeFunc(
		prometheus.GaugeOpts{
			Name:        fmt.Sprintf("%s_pico_temperature", picoName),
			Help:        "BME280 Sensor Temperature.",
			ConstLabels: prometheus.Labels{"unit": "fahrenheit"}},
		func() float64 {
			return m.getMetrics(client, url).tempF()
		},
	)

	promauto.NewGaugeFunc(
		prometheus.GaugeOpts{
			Name:        fmt.Sprintf("%s_pico_relative_humidity", picoName),
			Help:        "BME280 Sensor Relative Humidity.",
			ConstLabels: prometheus.Labels{"unit": "percent"}},
		func() float64 {
			return m.getMetrics(client, url).RH()
		},
	)

	promauto.NewGaugeFunc(
		prometheus.GaugeOpts{
			Name: fmt.Sprintf("%s_pico_up", picoName),
			Help: "Pico Sensor Server Status.",
		},
		func() float64 {
			return m.getMetrics(client, url).status()
		},
	)

	mux.HandleFunc("/", func(w http.ResponseWriter, _ *http.Request) {
		w.Write(rootPageHtml)
	})

	mux.Handle("/metrics", promhttp.Handler())

	return mux

}

func main() {
	picoURL := os.Getenv("PICO_SERVER_URL")
	picoName := os.Getenv("PICO_NAME")
	port := os.Getenv("PORT")

	s := &http.Server{
		Addr:         fmt.Sprintf(":%s", port),
		Handler:      newMux(picoURL, picoName),
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	if err := s.ListenAndServe(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
