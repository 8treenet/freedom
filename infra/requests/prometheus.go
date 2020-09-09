package requests

import (
	"fmt"
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

const (
	httpClientReqsName    = "http_client_requests_total"
	httpClientLatencyName = "http_client_duration_seconds"
)

type prom interface {
	RegisterHistogram(histogram *prometheus.HistogramVec)
	RegisterCounter(conter *prometheus.CounterVec)
}

// InstallPrometheus .
func InstallPrometheus(serviceName string, p prom) {
	httpClientReqs := prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name:        httpClientReqsName,
			Help:        "",
			ConstLabels: prometheus.Labels{"service": serviceName},
		},
		[]string{"domain", "http_code", "protocol", "method"},
	)
	p.RegisterCounter(httpClientReqs)
	httpClientLatency := prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Name:        httpClientLatencyName,
		Help:        "",
		ConstLabels: prometheus.Labels{"service": serviceName},
	},
		[]string{"domain", "http_code", "protocol", "method"},
	)
	p.RegisterHistogram(httpClientLatency)

	UseMiddleware(func(middle Middleware) {
		now := time.Now()
		middle.Next()
		domain := middle.GetRequest().URL.Host
		method := middle.GetRequest().Method
		rep := middle.GetRespone()
		code := ""
		protocol := ""
		if rep.Error != nil {
			protocol = rep.Proto
			code = fmt.Sprint(rep.StatusCode)
		} else {
			code = fmt.Sprintf("dial tcp %s: i/o timeout", domain)
		}
		httpClientReqs.WithLabelValues(domain, code, protocol, method).Inc()
		httpClientLatency.WithLabelValues(domain, code, protocol, method).Observe(float64(time.Since(now).Nanoseconds()) / 1000000000)
	})
}
