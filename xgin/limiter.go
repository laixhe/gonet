package xgin

import (
	"net/http"
	"sync"

	"github.com/gin-gonic/gin"
	"golang.org/x/time/rate"
)

// LimiterConfig 限流中间件配置。
type LimiterConfig struct {
	// Rate 每秒允许的请求数（令牌桶速率）。
	Rate float64
	// Burst 允许的突发请求数（令牌桶容量）。
	Burst int
}

// UseLimiter 注册全局限流中间件，基于客户端 IP 使用令牌桶算法。
// 超限返回 429 JSON（RateLimitError 格式）。
//
// 用法：
//
//	s.UseLimiter(xgin.LimiterConfig{Rate: 100, Burst: 200})
func (s *Server) UseLimiter(config ...LimiterConfig) *Server {
	cfg := LimiterConfig{Rate: 5, Burst: 10}
	if len(config) > 0 {
		cfg = config[0]
	}

	var mu sync.Mutex
	limiters := make(map[string]*rate.Limiter)

	s.app.Use(func(c *gin.Context) {
		ip := c.ClientIP()
		mu.Lock()
		lim, exists := limiters[ip]
		if !exists {
			lim = rate.NewLimiter(rate.Limit(cfg.Rate), cfg.Burst)
			limiters[ip] = lim
		}
		mu.Unlock()

		if !lim.Allow() {
			c.AbortWithStatusJSON(http.StatusTooManyRequests, RateLimitError())
			return
		}
		c.Next()
	})
	return s
}
