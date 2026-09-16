package exporter

import (
	"reflect"
	"strings"
	"testing"

	"github.com/showwin/speedtest-go/speedtest"
)

func TestMetricDescriptorsHaveStableLabels(t *testing.T) {
	descriptors := map[string]string{
		"up":              up.String(),
		"scrape duration": scrapeDurationSeconds.String(),
		"server distance": serverDistanceKilometers.String(),
		"latency":         latency.String(),
		"upload":          upload.String(),
		"download":        download.String(),
	}

	for name, descriptor := range descriptors {
		if strings.Contains(descriptor, "test_uuid") {
			t.Errorf("%s descriptor contains test_uuid: %s", name, descriptor)
		}
		if !strings.Contains(descriptor, "variableLabels: {}") {
			t.Errorf("%s descriptor has variable labels: %s", name, descriptor)
		}
	}
}

func TestInfoDescriptorLabels(t *testing.T) {
	descriptor := info.String()
	if strings.Contains(descriptor, "test_uuid") {
		t.Fatalf("info descriptor contains test_uuid: %s", descriptor)
	}

	for _, label := range resultLabelNames {
		if !strings.Contains(descriptor, label) {
			t.Errorf("info descriptor is missing %q: %s", label, descriptor)
		}
	}
}

func TestResultLabelValues(t *testing.T) {
	user := &speedtest.User{
		Lat: "1.23",
		Lon: "4.56",
		IP:  "203.0.113.1",
		Isp: "Example ISP",
	}
	server := &speedtest.Server{
		Lat:     "7.89",
		Lon:     "0.12",
		ID:      "42",
		Name:    "Example Server",
		Country: "Example Country",
	}

	want := []string{
		"1.23",
		"4.56",
		"203.0.113.1",
		"Example ISP",
		"7.89",
		"0.12",
		"42",
		"Example Server",
		"Example Country",
	}
	if got := resultLabelValues(user, server); !reflect.DeepEqual(got, want) {
		t.Fatalf("resultLabelValues() = %q, want %q", got, want)
	}
}
