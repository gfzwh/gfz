package metrics

import (
	"runtime"
	"sync"
	"time"

	"github.com/gfzwh/gfz/config"
	"github.com/gfzwh/gfz/zzlog"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/push"
	"go.uber.org/zap"
)

type metrics struct {
	registry   *prometheus.Registry
	counterVec *prometheus.CounterVec
	gfzVec     *prometheus.GaugeVec
	summaryVec *prometheus.SummaryVec
}

var instance *metrics
var once sync.Once

func getMetric() *metrics {
	once.Do(func() {
		instance = &metrics{
			registry: prometheus.NewRegistry(),
			counterVec: prometheus.NewCounterVec(
				prometheus.CounterOpts{
					Name: "gfz_call",
					Help: "How many RPC requests processed, partitioned by status code and RPC method.",
				},
				[]string{"code", "method"},
			),
			// 用于统计用（链接、断开连接,panic,错误,...）
			gfzVec: prometheus.NewGaugeVec(
				prometheus.GaugeOpts{
					Namespace: "",
					Subsystem: "",
					Name:      "gfz_sys",
					Help:      "Number of blob storage operations waiting to be processed, partitioned by user and type.",
				},
				[]string{
					// Which user has requested the operation?
					"type",
					// Of what type is the operation?
					"value",
				},
			),
			summaryVec: prometheus.NewSummaryVec(
				prometheus.SummaryOpts{
					Name:       "gfz_call_delay",
					Help:       "The temperature of the frog pond.",
					Objectives: map[float64]float64{0.5: 0.05, 0.9: 0.01, 0.99: 0.001},
				},
				[]string{"method"},
			),
		}

		instance.registry.MustRegister(instance.counterVec)
		instance.registry.MustRegister(instance.gfzVec)
		instance.registry.MustRegister(instance.summaryVec)

		go func() {
			for {
				instance.gfzVec.With(prometheus.Labels{"type": "server", "value": "gorouting"}).Set(float64(runtime.NumGoroutine()))

				if err := push.New(config.Get("metrics", "push").String(""), config.Get("server", "s2sname").String("")).
					Collector(instance.counterVec).
					Collector(instance.gfzVec).
					Collector(instance.summaryVec).
					Push(); err != nil {
					zzlog.Errorw("push metrics err", zap.Error(err))
				}

				instance.gfzVec.Reset()
				instance.counterVec.Reset()
				instance.summaryVec.Reset()

				time.Sleep(1 * time.Second)
			}
		}()
	})

	return instance
}

func Gfz(_type, value string) {
	getMetric().gfzVec.With(prometheus.Labels{"type": _type, "value": value}).Inc()
}

func GfzByAdd(_type, value string, add int64) {
	getMetric().gfzVec.With(prometheus.Labels{"type": _type, "value": value}).Set(float64(add))
}

func MethodCode(method, code string) {
	getMetric().counterVec.WithLabelValues(code, method).Inc()
}

func Summary(method string, startAt time.Time) {
	duration := time.Since(startAt).Milliseconds()
	getMetric().summaryVec.WithLabelValues(method).Observe(float64(duration))
}
