# StatsD Aggregator

A lightweight StatsD-like UDP metrics aggregator with Prometheus output support.

## Features
- Counters, gauges, and timers
- Histogram buckets for `ms` timer data
- Prometheus-compatible `/metrics` endpoint
- Dockerfile included

## Usage

### Run directly:
```
go run main.go
```

### Send metrics:
```
echo "latency:123|ms" | nc -u -w0 localhost 8125
```

### Prometheus:
Scrape from `http://localhost:9100/metrics`
