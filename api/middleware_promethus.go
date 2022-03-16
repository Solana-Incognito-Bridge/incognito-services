package api

import (
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/incognito-services/constants"
	"github.com/prometheus/client_golang/prometheus"
	ginprometheus "github.com/zsais/go-gin-prometheus"
)

func MeasureResponseDuration(p *ginprometheus.Prometheus) gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.Request.RequestURI == "/healthz" {
			c.Next()
			return
		}

		metric := p.MetricsList[0]
		if metric.ID != constants.MetricResponseDurationSeconds {
			c.Next()
			return
		}

		historyResDuration := metric.MetricCollector.(*prometheus.HistogramVec)

		start := time.Now()
		c.Next()
		duration := time.Since(start)
		statusCode := strconv.Itoa(c.Writer.Status())

		historyResDuration.WithLabelValues(c.Request.Method, statusCode, c.Request.URL.Path).Observe(duration.Seconds())
	}
}
