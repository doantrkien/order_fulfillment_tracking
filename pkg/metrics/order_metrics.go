package metrics

import (
	"runtime"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	OrderCreatedTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "order_created_total",
		Help: "Số đơn hàng tạo mới",
	}, []string{"result"})

	OrderStatusUpdatedTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "order_status_updated_total",
		Help: "Số lần cập nhật trạng thái",
	}, []string{"new_status", "result"})

	OrderHTTPRequestTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "order_http_requests_total",
		Help: "Tổng HTTP request vào order API",
	}, []string{"method", "path", "status_code"})

	OrderHTTPDuration = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "order_http_duration_seconds",
		Help:    "Latency các endpoint order",
		Buckets: []float64{0.01, 0.05, 0.1, 0.3, 0.5, 1},
	}, []string{"method", "path"})

	OrderDBQueryDuration = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "order_db_query_duration_seconds",
		Help:    "Thời gian query DB của order",
		Buckets: []float64{0.001, 0.005, 0.01, 0.05, 0.1, 0.5},
	}, []string{"operation"})

	AppHeapAllocBytes = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "app_heap_alloc_bytes",
		Help: "Bộ nhớ heap hiện tại được cấp phát bởi môi trường chạy Go",
	})

	AppHeapSysBytes = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "app_heap_sys_bytes",
		Help: "Tổng bộ nhớ được cấp phát bởi môi trường chạy Go",
	})
)

func StartMemoryCollector(interval time.Duration) {
	go func() {
		var mem runtime.MemStats
		for {
			runtime.ReadMemStats(&mem)
			AppHeapAllocBytes.Set(float64(mem.HeapAlloc))
			AppHeapSysBytes.Set(float64(mem.HeapSys))
			time.Sleep(interval)
		}
	}()
}
