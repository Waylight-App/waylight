# Prometheus metrics

Waylight can expose its measurements at `GET /metrics` in the Prometheus
text format, so an existing Prometheus/Grafana setup can scrape them. The
endpoint is **off by default**.

## Enable

```sh
# .env
METRICS_ENABLED=true
METRICS_TOKEN=<at least 32 characters, e.g. the output of: openssl rand -hex 32>
```

The endpoint requires `Authorization: Bearer <METRICS_TOKEN>` and is, like
every other route, limited to `ALLOWED_CIDRS`. It exposes infrastructure
details (service names, container names, disk sizes), so keep the token
secret and do not publish the port.

## Scrape config

```yaml
scrape_configs:
  - job_name: waylight
    scrape_interval: 60s
    metrics_path: /metrics
    authorization:
      type: Bearer
      credentials_file: /etc/prometheus/waylight.token
    static_configs:
      - targets: ["192.168.178.115:8090"]
```

A ready-made Grafana dashboard is in
[`grafana-dashboard.json`](grafana-dashboard.json) (Dashboards → New →
Import, then pick your Prometheus data source).

## Metrics

All names are stable; labels have a small, bounded set of values (names you
configured yourself), never request IDs or paths.

| Metric | Type | Labels | Meaning |
| --- | --- | --- | --- |
| `waylight_build_info` | gauge | `version`, `goversion` | Always 1 |
| `waylight_service_up` | gauge | `service` | 1 if the last health check succeeded |
| `waylight_service_response_seconds` | gauge | `service` | Response time of the last check |
| `waylight_service_cert_expiry_timestamp_seconds` | gauge | `service` | TLS certificate expiry (HTTPS services) |
| `waylight_containers` | gauge | `state` (`running`, `stopped`, `unhealthy`) | Number of containers |
| `waylight_container_running` | gauge | `container` | 1 if the container is running |
| `waylight_backup_ok` | gauge | `name`, `kind` (`backup`, `check`) | 1 if ok or running, 0 if overdue or failed (waiting sources are omitted) |
| `waylight_backup_last_success_timestamp_seconds` | gauge | `name`, `kind` | Last successful run |
| `waylight_host_cpu_ratio` | gauge | – | CPU usage (0–1) |
| `waylight_host_load1` | gauge | – | 1-minute load average |
| `waylight_host_memory_used_bytes`, `waylight_host_memory_total_bytes` | gauge | – | Memory |
| `waylight_host_disk_used_bytes`, `waylight_host_disk_total_bytes` | gauge | `disk` | Per `SYSTEM_DISK_PATHS` entry |
| `waylight_host_temperature_celsius` | gauge | – | Highest sensor reading (omitted without sensors) |
| `waylight_host_uptime_seconds` | gauge | – | Host uptime |
| `waylight_http_requests_total` | counter | `method`, `code` (`2xx` … `5xx`) | Requests served by Waylight |
| `waylight_http_request_duration_seconds_sum` / `_count` | counter | – | Time spent serving requests |
| `go_goroutines`, `go_memstats_heap_alloc_bytes`, `go_memstats_sys_bytes`, `go_gc_cycles_total` | gauge/counter | – | Go runtime |
| `process_start_time_seconds` | gauge | – | Start of the process |

Metrics of disabled modules are simply missing.

## Design

Modules implement the format-neutral `kernel.MetricSource` port and know
nothing about Prometheus. `cmd/server` collects the samples of all enabled
modules and `internal/platform/metrics` renders them by hand (no client
library, see ADR 0010); runtime metrics come from the standard library's
`runtime/metrics`.
