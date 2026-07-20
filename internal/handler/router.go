package handler

import (
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

func NewRouter(authHandler *AuthHandler, gameRecordHandler *GameRecordHandler, authMiddleware gin.HandlerFunc) *gin.Engine {
	r := gin.Default()

	r.Use(cors.New(cors.Config{
		AllowOrigins:     allowedOrigins(),
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Authorization", "Content-Type"},
		AllowCredentials: false,
		MaxAge:           12 * time.Hour,
	}))

	// ヘルスチェック(認証不要・DBアクセスなし)。Renderのスピンダウン回避のキープアライブpingにも使う。
	r.GET("/health", func(ctx *gin.Context) {
		ctx.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	// Swagger
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	v1 := r.Group("api/v1")
	{
		// 認証不要(Supabase Authへのプロキシ)
		auth := v1.Group("/auth")
		{
			auth.POST("/signup", authHandler.SignUp)
			auth.POST("/login", authHandler.Login)
			auth.POST("/refresh", authHandler.Refresh)
			auth.POST("/recover", authHandler.RequestPasswordReset)
			auth.POST("/reset-password", authHandler.ResetPassword)
		}

		// 認証必要
		protected := v1.Group("")
		protected.Use(authMiddleware)
		{
			records := protected.Group("/records")
			{
				records.POST("", gameRecordHandler.CreateGameRecord)
				records.GET("", gameRecordHandler.GetGameRecords)
				records.PUT("/:id", gameRecordHandler.UpdateGameRecord)
				records.DELETE("/:id", gameRecordHandler.DeleteGameRecord)
			}

			stats := protected.Group("/stats")
			{
				stats.GET("/ratings", gameRecordHandler.GetDailyRatings)
				stats.GET("/summary", gameRecordHandler.GetSummaryStats)
			}
		}
	}
	return r
}

// allowedOrigins はALLOWED_ORIGINS環境変数(カンマ区切り)からCORS許可オリジンを組み立てる。
func allowedOrigins() []string {
	raw := os.Getenv("ALLOWED_ORIGINS")
	if raw == "" {
		return []string{"http://localhost:3000"}
	}
	origins := strings.Split(raw, ",")
	for i, o := range origins {
		origins[i] = strings.TrimSpace(o)
	}
	return origins
}
