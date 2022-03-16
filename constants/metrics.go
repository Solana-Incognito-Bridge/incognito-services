package constants

import (
	ginprometheus "github.com/zsais/go-gin-prometheus"
)

const MetricResponseDurationSeconds = "ResponseDurationSeconds"

var metrics = &ginprometheus.Metric{
	ID:          MetricResponseDurationSeconds,
	Name:        "http_server_request_duration_seconds",
	Description: "Histogram of response time for handler in seconds",
	Type:        "histogram_vec",
	Args:        []string{"method", "status_code", "url"},
}

var ResponseTimeHistogram = []*ginprometheus.Metric{
	metrics,
}
