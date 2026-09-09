// Copyright 2024 ImageForge
// 安全响应头中间件 — 添加浏览器安全头防护。

package middleware

import (
	"github.com/gin-gonic/gin"
)

// SecurityHeaders 添加安全响应头，防护常见 Web 攻击。
// 注意：Content-Security-Policy 使用较宽松的配置，允许内联样式（Vue 需要）。
// 生产环境可根据需要收紧。
func SecurityHeaders() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 防止 MIME 类型嗅探
		c.Header("X-Content-Type-Options", "nosniff")
		// 点击劫持防护（Caddy 也设置，双重保障）
		c.Header("X-Frame-Options", "DENY")
		// XSS 过滤器（旧版浏览器）
		c.Header("X-XSS-Protection", "1; mode=block")
		// Referrer 策略
		c.Header("Referrer-Policy", "strict-origin-when-cross-origin")
		// 权限策略（限制浏览器 API）
		c.Header("Permissions-Policy", "camera=(), microphone=(), geolocation=()")
		// CSP — 允许 self 和 data: URI（图片），内联样式（Vue 需要）
		c.Header("Content-Security-Policy",
			"default-src 'self'; "+
				"img-src 'self' data: blob:; "+
				"style-src 'self' 'unsafe-inline'; "+
				"script-src 'self'; "+
				"connect-src 'self'; "+
				"font-src 'self' data:; "+
				"frame-ancestors 'none'; "+
				"base-uri 'self'; "+
				"form-action 'self'")
		// HSTS — 仅在 HTTPS 时由 Caddy 设置，后端也做兜底
		c.Header("Strict-Transport-Security", "max-age=63072000; includeSubDomains; preload")
		c.Next()
	}
}
