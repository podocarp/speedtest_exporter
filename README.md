<p align=center>
  <img alt=logo src="https://github.com/danopstech/speedtest_exporter/raw/main/.docs/assets/logo.jpg" height=150 />
  <h3 align=center>Speedtest Prometheus Exporter</h3>
</p>

---
A [Speedtest](https://www.speedtest.net) exporter for Prometheus.

Forked from https://github.com/danopstech/speedtest_exporter since there has not
been any updates there for a few years.
I will try to keep the binary up to date and fix any issues I find since I find
this tool useful, but I will not maintain the docker images since I don't use
them.

## Usage:

### Flags

`speedtest_exporter` is configured by optional command line flags

```bash
$ ./speedtest_exporter --help
Usage of speedtest_exporter:
  -port string
        listening port to expose metrics on (default "9090")
  -server_fallback
        If the server_id given is not available, should we fallback to closest available server
  -server_id int
        Speedtest.net server ID to run test against, -1 will pick the closest server to your location (default -1)
  -timeout int
        request timeout for the execution of the speedtest (default 60)

```

### Binaries

For pre-built binaries please take a look at the [releases](https://github.com/danopstech/speedtest_exporter/releases).

### Setup Prometheus to scrape `speedtest_exporter`

Configure [Prometheus](https://prometheus.io/) to scrape metrics from localhost:9090/metrics

This exporter locks (one concurrent scrape at a time) as it conducts the speedtest when scraped, **remember set scrape interval, and scrap timeout** accordingly as per example.

```yaml
...
scrape_configs
    - job_name: speedtest
      scrape_interval: 60m
      scrape_timeout:  60s
      static_configs:
        - targets: ['localhost:9090']
...
```

## Exported Metrics:

```
# HELP speedtest_download_speed_Bps Last download speedtest result
# TYPE speedtest_download_speed_Bps gauge
# HELP speedtest_latency_seconds Measured latency on last speed test
# TYPE speedtest_latency_seconds gauge
# HELP speedtest_scrape_duration_seconds Time to preform last speed test
# TYPE speedtest_scrape_duration_seconds gauge
# HELP speedtest_up Was the last speedtest successful.
# TYPE speedtest_up gauge
# HELP speedtest_upload_speed_Bps Last upload speedtest result
# TYPE speedtest_upload_speed_Bps gauge
```
## Example Grafana Dashboard:

https://grafana.com/grafana/dashboards/14336

<p align="center">
	<img src="https://github.com/danopstech/speedtest_exporter/raw/main/.docs/assets/screenshot.jpg" width="95%">
</p>
