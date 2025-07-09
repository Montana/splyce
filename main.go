package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"gopkg.in/yaml.v3"
)

type aggregator struct {
	mu         sync.Mutex
	counters   map[string]float64
	gauges     map[string]float64
	histograms map[string][]float64
}

type config struct {
	ListenUDPPort        int `yaml:"listen_udp_port"`
	HTTPMetricsPort      int `yaml:"http_metrics_port"`
	FlushIntervalSeconds int `yaml:"flush_interval_seconds"`
}

var cfg = config{
	ListenUDPPort:        8125,
	HTTPMetricsPort:      9100,
	FlushIntervalSeconds: 60,
}

var agg = aggregator{
	counters:   make(map[string]float64),
	gauges:     make(map[string]float64),
	histograms: make(map[string][]float64),
}

func loadConfig(path string) {
	data, err := os.ReadFile(path)
	if err != nil {
		log.Printf("splyce: could not read config, using defaults: %v", err)
		return
	}
	err = yaml.Unmarshal(data, &cfg)
	if err != nil {
		log.Fatalf("splyce: failed to parse config: %v", err)
	}
}

func parseLine(line string) {
	parts := strings.SplitN(line, ":", 2)
	if len(parts) != 2 {
		return
	}
	key := parts[0]
	rest := strings.Split(parts[1], "|")
	if len(rest) < 2 {
		return
	}

	valStr := rest[0]
	typ := rest[1]
	value, err := strconv.ParseFloat(valStr, 64)
	if err != nil {
		return
	}

	sampleRate := 1.0
	if len(rest) == 3 && strings.HasPrefix(rest[2], "@") {
		sr, err := strconv.ParseFloat(rest[2][1:], 64)
		if err == nil && sr > 0 {
			sampleRate = sr
		}
	}

	agg.mu.Lock()
	defer agg.mu.Unlock()

	switch typ {
	case "c":
		agg.counters[key] += value / sampleRate
	case "g":
		agg.gauges[key] = value
	case "ms":
		agg.histograms[key] = append(agg.histograms[key], value)
	}
}

func udpListener() {
	addr := net.UDPAddr{Port: cfg.ListenUDPPort, IP: net.ParseIP("0.0.0.0")}
	conn, err := net.ListenUDP("udp", &addr)
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("splyce: listening on UDP :%d", cfg.ListenUDPPort)

	buf := make([]byte, 1024)
	for {
		n, _, err := conn.ReadFromUDP(buf)
		if err != nil {
			continue
		}
		data := string(buf[:n])
		scanner := bufio.NewScanner(strings.NewReader(data))
		for scanner.Scan() {
			parseLine(scanner.Text())
		}
	}
}

func metricsHandler(w http.ResponseWriter, _ *http.Request) {
	agg.mu.Lock()
	defer agg.mu.Unlock()

	for k, v := range agg.counters {
		fmt.Fprintf(w, "splyce_counter_%s %f\n", sanitize(k), v)
	}
	for k, v := range agg.gauges {
		fmt.Fprintf(w, "splyce_gauge_%s %f\n", sanitize(k), v)
	}
	for k, vals := range agg.histograms {
		if len(vals) == 0 {
			continue
		}
		sort.Float64s(vals)
		count := len(vals)
		sum := 0.0
		for _, v := range vals {
			sum += v
		}
		buckets := []float64{50, 100, 250, 500, 1000, 2000, 5000}
		bucketCounts := make(map[float64]int)
		for _, b := range buckets {
			for _, val := range vals {
				if val <= b {
					bucketCounts[b]++
				}
			}
		}
		for _, b := range buckets {
			fmt.Fprintf(w, "splyce_timer_%s_bucket{le=\"%f\"} %d\n", sanitize(k), b, bucketCounts[b])
		}
		fmt.Fprintf(w, "splyce_timer_%s_bucket{le=\"+Inf\"} %d\n", sanitize(k), count)
		fmt.Fprintf(w, "splyce_timer_%s_count %d\n", sanitize(k), count)
		fmt.Fprintf(w, "splyce_timer_%s_sum %f\n", sanitize(k), sum)
	}
}

func sanitize(s string) string {
	return strings.ReplaceAll(s, ".", "_")
}

func startHTTP() {
	http.HandleFunc("/metrics", metricsHandler)
	log.Printf("splyce: serving Prometheus metrics on :%d/metrics", cfg.HTTPMetricsPort)
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", cfg.HTTPMetricsPort), nil))
}

func main() {
	loadConfig("splyce.yaml")
	go udpListener()
	go startHTTP()

	go func() {
		for {
			time.Sleep(time.Duration(cfg.FlushIntervalSeconds) * time.Second)
			agg.mu.Lock()
			agg.histograms = make(map[string][]float64)
			agg.mu.Unlock()
		}
	}()

	select {}
}
