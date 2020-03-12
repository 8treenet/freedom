package general

import (
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/jinzhu/gorm"
	"github.com/kataras/iris/context"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
// DefaultBuckets prometheus buckets in seconds.
//DefaultBuckets = []float64{0.3, 1.2, 5.0}
)

const (
	reqsName              = "http_requests_total"
	latencyName           = "http_request_duration_seconds"
	ormName               = "orm_requests_total"
	ormlatencyName        = "orm_duration_seconds"
	httpClientReqsName    = "http_client_requests_total"
	httpClientLatencyName = "http_client_duration_seconds"
	redisReqsName         = "redis_client_requests_total"
	redisLatencyName      = "redis_client_duration_seconds"

	kafkaProducerReqsName    = "kafka_producer_requests_total"
	kafkaProducerLatencyName = "kafka_producer_duration_seconds"
)

// Prometheus is a handler that exposes prometheus metrics for the number of requests,
// the latency and the response size, partitioned by status code, method and HTTP path.
//
// Usage: pass its `ServeHTTP` to a route or globally.
type Prometheus struct {
	reqs    *prometheus.CounterVec
	latency *prometheus.HistogramVec
	listen  string

	ormReqs    *prometheus.CounterVec
	ormLatency *prometheus.HistogramVec

	httpClientReqs    *prometheus.CounterVec
	httpClientLatency *prometheus.HistogramVec

	kafkaProducerReqs    *prometheus.CounterVec
	kafkaProducerLatency *prometheus.HistogramVec

	redisClientReqs    *prometheus.CounterVec
	redisClientLatency *prometheus.HistogramVec
}

type log interface {
	Info(v ...interface{})
}

// newPrometheus returns a new prometheus middleware.
//
// If buckets are empty then `DefaultBuckets` are set.
func newPrometheus(name, listen string) *Prometheus {
	p := Prometheus{}
	p.listen = listen
	p.reqs = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name:        reqsName,
			Help:        "How many HTTP requests processed, partitioned by status code, method and HTTP path.",
			ConstLabels: prometheus.Labels{"service": name},
		},
		[]string{"http_code", "code", "method", "path"},
	)
	prometheus.MustRegister(p.reqs)
	p.latency = prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Name:        latencyName,
		Help:        "How long it took to process the request, partitioned by status code, method and HTTP path.",
		ConstLabels: prometheus.Labels{"service": name},
	},
		[]string{"http_code", "code", "method", "path"},
	)
	prometheus.MustRegister(p.latency)

	p.ormReqs = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name:        ormName,
			Help:        "",
			ConstLabels: prometheus.Labels{"service": name},
		},
		[]string{"result", "model", "method"},
	)
	prometheus.MustRegister(p.ormReqs)

	p.ormLatency = prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Name:        ormlatencyName,
		Help:        "",
		ConstLabels: prometheus.Labels{"service": name},
	},
		[]string{"model", "method", "result"},
	)
	prometheus.MustRegister(p.ormLatency)

	p.httpClientReqs = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name:        httpClientReqsName,
			Help:        "",
			ConstLabels: prometheus.Labels{"service": name},
		},
		[]string{"domain", "http_code", "protocol", "method"},
	)
	prometheus.MustRegister(p.httpClientReqs)

	p.httpClientLatency = prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Name:        httpClientLatencyName,
		Help:        "",
		ConstLabels: prometheus.Labels{"service": name},
	},
		[]string{"domain", "http_code", "protocol", "method"},
	)
	prometheus.MustRegister(p.httpClientLatency)

	p.kafkaProducerReqs = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name:        kafkaProducerReqsName,
			Help:        "",
			ConstLabels: prometheus.Labels{"service": name},
		},
		[]string{"topic", "error"},
	)
	prometheus.MustRegister(p.kafkaProducerReqs)

	p.kafkaProducerLatency = prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Name:        kafkaProducerLatencyName,
		Help:        "",
		ConstLabels: prometheus.Labels{"service": name},
	},
		[]string{"topic", "error"},
	)
	prometheus.MustRegister(p.kafkaProducerLatency)

	p.redisClientReqs = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name:        redisReqsName,
			Help:        "",
			ConstLabels: prometheus.Labels{"service": name},
		},
		[]string{"cmd", "error"},
	)
	prometheus.MustRegister(p.redisClientReqs)

	p.redisClientLatency = prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Name:        redisLatencyName,
		Help:        "",
		ConstLabels: prometheus.Labels{"service": name},
	},
		[]string{"cmd", "error"},
	)
	prometheus.MustRegister(p.redisClientLatency)

	return &p
}

// newPrometheusHandle .
func newPrometheusHandle(p *Prometheus) func(context.Context) {
	http.Handle("/", promhttp.Handler())
	go func() {
		if strings.Index(p.listen, ":") == 0 {

			globalApp.Logger().Infof("Now prometheus listening on: http://0.0.0.0%s\n", p.listen)
		} else {
			globalApp.Logger().Infof("Now prometheus listening on: http://%s\n", p.listen)
		}
		http.ListenAndServe(p.listen, nil)
	}()

	return func(ctx context.Context) {
		start := time.Now()
		ctx.Next()
		r := ctx.Request()
		statusCode := strconv.Itoa(ctx.GetStatusCode())

		path := ctx.GetCurrentRoute().Path()
		p.reqs.WithLabelValues(statusCode, "0", r.Method, path).
			Inc()

		p.latency.WithLabelValues(statusCode, "0", r.Method, path).
			Observe(float64(time.Since(start).Nanoseconds()) / 1000000000)
	}
}

func (p *Prometheus) OrmWithLabelValues(model, method string, e error, starTime time.Time) {
	result := "ok"
	if e != nil {
		result = "error"
	}
	if e == gorm.ErrRecordNotFound {
		result = "ok"
	}

	p.ormReqs.WithLabelValues(model, method, result).Inc()
	p.ormLatency.WithLabelValues(model, method, result).Observe(float64(time.Since(starTime).Nanoseconds()) / 1000000000)
}

func (p *Prometheus) HttpClientWithLabelValues(domain, httpCode, protocol, method string, starTime time.Time) {
	p.httpClientReqs.WithLabelValues(domain, httpCode, protocol, method).Inc()
	p.httpClientLatency.WithLabelValues(domain, httpCode, protocol, method).Observe(float64(time.Since(starTime).Nanoseconds()) / 1000000000)
}

func (p *Prometheus) KafkaProducerWithLabelValues(topic string, e error, starTime time.Time) {
	err := ""
	if e != nil {
		err = e.Error()
	}
	p.kafkaProducerReqs.WithLabelValues(topic, err).Inc()
	p.kafkaProducerLatency.WithLabelValues(topic, err).Observe(float64(time.Since(starTime).Nanoseconds()) / 1000000000)
}

func (p *Prometheus) RedisClientWithLabelValues(cmd string, e error, starTime time.Time) {
	err := ""
	if e != nil {
		err = e.Error()
	}
	p.redisClientReqs.WithLabelValues(cmd, err).Inc()
	p.redisClientLatency.WithLabelValues(cmd, err).Observe(float64(time.Since(starTime).Nanoseconds()) / 1000000000)
}
