package main

import (
    "bufio"
    "fmt"
    "log"
    "math"
    "net"
    "net/http"
    "sort"
    "strconv"
    "strings"
    "sync"
    "time"
)

type aggregator struct {
    mu         sync.Mutex
    counters   map[string]float64
    gauges     map[string]float64
    histograms map[string][]float64
}

var agg = aggregator{
    counters:   make(map[string]float64),
    gauges:     make(map[string]float64),
    histograms: make(map[string][]float64),
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
    addr := net.UDPAddr{Port: 8125, IP: net.ParseIP("0.0.0.0")}
    conn, err := net.ListenUDP("udp", &addr)
    if err != nil {
        log.Fatal(err)
    }
    log.Println("Listening on UDP :8125")

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
        fmt.Fprintf(w, "splyce_counter_%s %f
", sanitize(k), v)
    }
    for k, v := range agg.gauges {
        fmt.Fprintf(w, "splyce_gauge_%s %f
", sanitize(k), v)
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
            fmt.Fprintf(w, "splyce_timer_%s_bucket{le=\"%f\"} %d
", sanitize(k), b, bucketCounts[b])
        }
        fmt.Fprintf(w, "splyce_timer_%s_bucket{le=\"+Inf\"} %d
", sanitize(k), count)
        fmt.Fprintf(w, "splyce_timer_%s_count %d
", sanitize(k), count)
        fmt.Fprintf(w, "splyce_timer_%s_sum %f
", sanitize(k), sum)
    }
}

func sanitize(s string) string {
    return strings.ReplaceAll(s, ".", "_")
}

func startHTTP() {
    http.HandleFunc("/metrics", metricsHandler)
    log.Println("Serving Prometheus metrics on :9100/metrics")
    log.Fatal(http.ListenAndServe(":9100", nil))
}

func main() {
    go udpListener()
    go startHTTP()

    go func() {
        for {
            time.Sleep(60 * time.Second)
            agg.mu.Lock()
            agg.histograms = make(map[string][]float64)
            agg.mu.Unlock()
        }
    }()

    select {}
}
