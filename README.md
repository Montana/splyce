# splyce

splyce is a lightweight StatsD-compatible metrics aggregator written in Go. It collects metrics over UDP, performs local aggregation, and exposes results in Prometheus-compatible format over HTTP.

## Features

- Supports StatsD metric types:
  - Counters (`|c`)
  - Gauges (`|g`)
  - Timers (`|ms`)
  - Sampling (`|@0.1`)
- Prometheus `/metrics` endpoint
- YAML configuration file (`splyce.yaml`)
- Periodic flushing of histograms
- Minimal resource usage and zero external dependencies

<br>


<img width="983" alt="Screenshot 2025-07-08 at 7 19 00 PM" src="https://github.com/user-attachments/assets/031ba313-ef10-4be1-80b5-4ebc4719e242" />


## Build and Run

Clone the repository and build the binary:

```bash
git clone https://github.com/montana/splyce.git
cd splyce
go build -o splyce main.go
```
Run the aggregator:

```bash
./splyce
```

To use the `splyce.yaml` instead of Prometheus:

```bash
./splyce --config=/path/to/splyce.yaml
```

## Example Metrics Input

```bash
echo "api_hits:1|c|@0.5" | nc -u -w0 127.0.0.1 8125
echo "latency:250|ms"   | nc -u -w0 127.0.0.1 8125
```
## Prometheus Integration

If you'd like to use Prometheus: 

```bash
scrape_configs:
  - job_name: 'splyce'
    static_configs:
      - targets: ['localhost:9100']
```
splyce in particular exposes these metrics:

```bash
splyce_counter_api_requests 5
splyce_gauge_loadavg 0.42
splyce_timer_latency_bucket{le="100"} 1
splyce_timer_latency_count 1
splyce_timer_latency_sum 123
```

## Author
Michael Mendy (c) 2025.
