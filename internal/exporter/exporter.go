package exporter

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/showwin/speedtest-go/speedtest"
	log "github.com/sirupsen/logrus"
)

const (
	namespace = "speedtest"
)

var (
	resultLabelNames = []string{"user_lat", "user_lon", "user_ip", "user_isp", "server_lat", "server_lon", "server_id", "server_name", "server_country"}
	up               = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "", "up"),
		"Was the last speedtest successful.",
		nil, nil,
	)
	scrapeDurationSeconds = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "", "scrape_duration_seconds"),
		"Time to perform the last speed test",
		nil, nil,
	)
	info = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "", "info"),
		"Metadata about the speed test client and server",
		resultLabelNames,
		nil,
	)
	serverDistanceKilometers = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "", "server_distance_kilometers"),
		"Distance to the speed test server in kilometers",
		nil, nil,
	)
	latency = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "", "latency_seconds"),
		"Measured latency on last speed test",
		nil, nil,
	)
	upload = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "", "upload_speed_Bps"),
		"Last upload speedtest result",
		nil, nil,
	)
	download = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "", "download_speed_Bps"),
		"Last download speedtest result",
		nil, nil,
	)
)

// Exporter runs speedtest and exports them using
// the prometheus metrics package.
type Exporter struct {
	serverID       int
	serverFallback bool
	timeout        time.Duration
	mu             sync.Mutex
}

// New returns an initialized Exporter.
func New(serverID int, serverFallback bool, timeout time.Duration) (*Exporter, error) {
	return &Exporter{
		serverID:       serverID,
		serverFallback: serverFallback,
		timeout:        timeout,
	}, nil
}

// Describe describes all the metrics. It implements prometheus.Collector.
func (e *Exporter) Describe(ch chan<- *prometheus.Desc) {
	ch <- up
	ch <- scrapeDurationSeconds
	ch <- info
	ch <- serverDistanceKilometers
	ch <- latency
	ch <- upload
	ch <- download
}

// Collect fetches the stats from Starlink dish and delivers them
// as Prometheus metrics. It implements prometheus.Collector.
func (e *Exporter) Collect(ch chan<- prometheus.Metric) {
	if !e.mu.TryLock() {
		log.Warn("speedtest scrape already in progress")
		ch <- prometheus.MustNewConstMetric(up, prometheus.GaugeValue, 0.0)
		return
	}
	defer e.mu.Unlock()

	start := time.Now()
	ok := e.speedtest(ch)

	if ok {
		ch <- prometheus.MustNewConstMetric(up, prometheus.GaugeValue, 1.0)
		ch <- prometheus.MustNewConstMetric(
			scrapeDurationSeconds, prometheus.GaugeValue, time.Since(start).Seconds(),
		)
	} else {
		ch <- prometheus.MustNewConstMetric(up, prometheus.GaugeValue, 0.0)
	}
}

func (e *Exporter) speedtest(ch chan<- prometheus.Metric) bool {
	client := speedtest.New()
	ctx, cancel := context.WithTimeout(context.Background(), e.timeout)
	defer cancel()

	user, err := client.FetchUserInfoContext(ctx)
	if err != nil {
		log.Errorf("could not fetch user information: %s", err.Error())
		return false
	}

	// Returns list of servers in distance order (distance is computed using the
	// user info fetched above).
	serverList, err := client.FetchServerListContext(ctx)
	if err != nil {
		log.Errorf("could not fetch server list: %s", err.Error())
		return false
	}
	if len(serverList) == 0 {
		log.Error("could not fetch server list: no servers returned")
		return false
	}

	var server *speedtest.Server

	if e.serverID == -1 {
		server = serverList[0]
	} else {
		servers, err := serverList.FindServer([]int{e.serverID})
		if err != nil {
			log.Error(err)
			return false
		}

		if servers[0].ID != fmt.Sprintf("%d", e.serverID) && !e.serverFallback {
			log.Errorf("could not find your choosen server ID %d in the list of avaiable servers, server_fallback is not set so failing this test", e.serverID)
			return false
		}

		server = servers[0]
	}

	ch <- prometheus.MustNewConstMetric(info, prometheus.GaugeValue, 1.0, resultLabelValues(user, server)...)
	ch <- prometheus.MustNewConstMetric(serverDistanceKilometers, prometheus.GaugeValue, server.Distance)

	ok := pingTest(ctx, server, ch)
	ok = downloadTest(ctx, server, ch) && ok
	ok = uploadTest(ctx, server, ch) && ok

	return ok
}

func resultLabelValues(user *speedtest.User, server *speedtest.Server) []string {
	return []string{
		user.Lat,
		user.Lon,
		user.IP,
		user.Isp,
		server.Lat,
		server.Lon,
		server.ID,
		server.Name,
		server.Country,
	}
}

func pingTest(ctx context.Context, server *speedtest.Server, ch chan<- prometheus.Metric) bool {
	err := server.PingTestContext(ctx, nil)
	if err != nil {
		log.Errorf("failed to carry out ping test: %s", err.Error())
		return false
	}

	ch <- prometheus.MustNewConstMetric(latency, prometheus.GaugeValue, server.Latency.Seconds())

	return true
}

func downloadTest(ctx context.Context, server *speedtest.Server, ch chan<- prometheus.Metric) bool {
	err := server.DownloadTestContext(ctx)
	if err != nil {
		log.Errorf("failed to carry out download test: %s", err.Error())
		return false
	}

	ch <- prometheus.MustNewConstMetric(download, prometheus.GaugeValue, float64(server.DLSpeed))

	return true
}

func uploadTest(ctx context.Context, server *speedtest.Server, ch chan<- prometheus.Metric) bool {
	err := server.UploadTestContext(ctx)
	if err != nil {
		log.Errorf("failed to carry out upload test: %s", err.Error())
		return false
	}

	ch <- prometheus.MustNewConstMetric(upload, prometheus.GaugeValue, float64(server.ULSpeed))

	return true
}
