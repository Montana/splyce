# splyce

splyce is a lightweight StatsD-like UDP metrics aggregator with Prometheus output support.

```bash
go run main.go
```

### Send metrics:
```
echo "latency:123|ms" | nc -u -w0 localhost 8125
```

### Prometheus:
Scrape from `http://localhost:9100/metrics`
