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

## Build and Run

Clone the repository and build the binary:

```bash
git clone https://github.com/your-username/splyce.git
cd splyce
go build -o splyce main.go
```
Run the aggregator:

```bash
./splyce
```
## Example Metrics Input

```bash
echo "api_hits:1|c|@0.5" | nc -u -w0 127.0.0.1 8125
echo "latency:250|ms"   | nc -u -w0 127.0.0.1 8125
```

## Author
Michael Mendy (c) 2025.
