package metrics

import (
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
	socketVec  *prometheus.GaugeVec
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
			// 用于统计用（链接、断开连接,...）
			socketVec: prometheus.NewGaugeVec(
				prometheus.GaugeOpts{
					Namespace: "",
					Subsystem: "",
					Name:      "conn_queued",
					Help:      "Number of blob storage operations waiting to be processed, partitioned by user and type.",
				},
				[]string{
					// Which user has requested the operation?
					"host",
					// Of what type is the operation?
					"type",
				},
			),
		}

		instance.registry.MustRegister(instance.counterVec)
		instance.registry.MustRegister(instance.socketVec)

		go func() {
			for i := 0; i < 1000; i++ {
				if err := push.New(config.Get("metrics", "push").String(""), config.Get("server", "s2sname").String("")).
					Collector(instance.counterVec).
					Collector(instance.socketVec).
					Push(); err != nil {
					zzlog.Errorw("push metrics err", zap.Error(err))
				}
				time.Sleep(1 * time.Second)
			}
		}()
	})

	return instance
}

func SocketVec(host, _type string) {
	getMetric().socketVec.With(prometheus.Labels{"type": _type, "host": host}).Inc()
}

func MethodCode(method, code string) {
	getMetric().counterVec.WithLabelValues(code, method).Inc()
}
