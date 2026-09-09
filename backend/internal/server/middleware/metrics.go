// Copyright 2024 ImageForge
// Prometheus 指标收集中间件。

package middleware

import (
	"strconv"
	"sync/atomic"
	"time"

	"github.com/gin-gonic/gin"
)

// metrics 保存全局指标。
var metrics struct {
	requestsTotal   int64
	requestDuration int64 // 纳秒
	errorsTotal     int64
}

// Metrics 收集中间件 — 记录请求数、延迟、错误率。
func Metrics() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		atomic.AddInt64(&metrics.requestsTotal, 1)

		c.Next()

		duration := time.Since(start).Nanoseconds()
		atomic.AddInt64(&metrics.requestDuration, duration)

		if c.Writer.Status() >= 400 {
			atomic.AddInt64(&metrics.errorsTotal, 1)
		}
	}
}

// MetricsHandler 返回 /metrics 端点的 Prometheus 格式指标。
func MetricsHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		reqTotal := atomic.LoadInt64(&metrics.requestsTotal)
		reqDuration := atomic.LoadInt64(&metrics.requestDuration)
		errTotal := atomic.LoadInt64(&metrics.errorsTotal)

		// 计算平均延迟（毫秒）
		var avgLatency float64
		if reqTotal > 0 {
			avgLatency = float64(reqDuration) / float64(reqTotal) / 1e6
		}

		c.Header("Content-Type", "text/plain; version=0.0.4; charset=utf-8")
		c.String(200, `# HELP imageforge_requests_total Total number of requests.
# TYPE imageforge_requests_total counter
imageforge_requests_total %d
# HELP imageforge_errors_total Total number of error responses (4xx/5xx).
# TYPE imageforge_errors_total counter
imageforge_errors_total %d
# HELP imageforge_request_duration_ms Average request latency in milliseconds.
# TYPE imageforge_request_duration_ms gauge
imageforge_request_duration_ms %.3f
`, reqTotal, errTotal, avgLatency)
	}
}

// StatusCodeString 辅助函数（避免未使用导入）。
func StatusCodeString(code int) string {
	return strconv.Itoa(code)
}
