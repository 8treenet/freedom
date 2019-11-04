package middleware

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/kataras/iris/context"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
// DefaultBuckets prometheus buckets in seconds.
//DefaultBuckets = []float64{0.3, 1.2, 5.0}
)

const (
	reqsName    = "http_requests_total"
	latencyName = "http_request_duration_seconds"
)

// Prometheus is a handler that exposes prometheus metrics for the number of requests,
// the latency and the response size, partitioned by status code, method and HTTP path.
//
// Usage: pass its `ServeHTTP` to a route or globally.
type Prometheus struct {
	reqs    *prometheus.CounterVec
	latency *prometheus.HistogramVec
}

type log interface {
	Info(v ...interface{})
}

// NewPrometheus returns a new prometheus middleware.
//
// If buckets are empty then `DefaultBuckets` are set.
func NewPrometheus(name, listen string) *Prometheus {
	p := Prometheus{}
	p.reqs = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name:        reqsName,
			Help:        "How many HTTP requests processed, partitioned by status code, method and HTTP path.",
			ConstLabels: prometheus.Labels{"service": name},
		},
		[]string{"code", "method", "path"},
	)
	prometheus.MustRegister(p.reqs)
	p.latency = prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Name:        latencyName,
		Help:        "How long it took to process the request, partitioned by status code, method and HTTP path.",
		ConstLabels: prometheus.Labels{"service": name},
	},
		[]string{"code", "method", "path"},
	)
	prometheus.MustRegister(p.latency)

	http.Handle("/", promhttp.Handler())
	go func() {
		if strings.Index(listen, ":") == 0 {
			fmt.Printf("Now prometheus listening on: http://0.0.0.0%s\n", listen)
		} else {
			fmt.Printf("Now prometheus listening on: http://%s\n", listen)
		}
		http.ListenAndServe(listen, nil)
	}()
	return &p
}

func (p *Prometheus) ServeHTTP(ctx context.Context) {
	start := time.Now()
	ctx.Next()
	r := ctx.Request()
	statusCode := strconv.Itoa(ctx.GetStatusCode())

	path := ctx.GetCurrentRoute().Path()
	p.reqs.WithLabelValues(statusCode, r.Method, path).
		Inc()

	p.latency.WithLabelValues(statusCode, r.Method, path).
		Observe(float64(time.Since(start).Nanoseconds()) / 1000000000)
}
