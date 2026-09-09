package middleware

import (
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

// CORS returns the Gin CORS middleware. When allowedOrigins is non-empty it is
// used verbatim (production domains such as https://imageforge.example.com);
// otherwise the default localhost dev origins are permitted.
func CORS(allowedOrigins []string) gin.HandlerFunc {
	origins := allowedOrigins
	if len(origins) == 0 {
		origins = []string{
			"http://localhost:5173",
			"http://localhost:3000",
			"http://127.0.0.1:5173",
		}
	}
	return cors.New(cors.Config{
		AllowOrigins:     origins,
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization", "X-Request-ID"},
		ExposeHeaders:    []string{"X-Request-ID", "Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	})
}
