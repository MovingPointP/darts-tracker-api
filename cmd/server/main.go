package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/MovingPointP/darts-tracker-api/docs"
	"github.com/MovingPointP/darts-tracker-api/internal/handler"
	"github.com/MovingPointP/darts-tracker-api/internal/handler/middleware"
	"github.com/MovingPointP/darts-tracker-api/internal/infrastructure/db"
	"github.com/MovingPointP/darts-tracker-api/internal/infrastructure/persistence"
	"github.com/MovingPointP/darts-tracker-api/internal/infrastructure/supabaseauth"
	"github.com/MovingPointP/darts-tracker-api/internal/usecase"
	"github.com/joho/godotenv"
)

// @title           darts-tracker-api
// @version         1.0
// @description     ダーツ得点記録アプリ REST API
// @BasePath        /api/v1
// @securityDefinitions.apikey BearerAuth
// @in              header
// @name            Authorization
func main() {
	// .envの読み込み
	_ = godotenv.Load()

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	host := os.Getenv("HOST")
	if host == "" {
		host = fmt.Sprintf("localhost:%s", port)
	}
	docs.SwaggerInfo.Host = host

	conn, err := db.NewDB()
	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}

	supabaseClient, err := supabaseauth.NewClient()
	if err != nil {
		log.Fatalf("failed to create supabase auth client: %v", err)
	}

	authMiddleware, err := middleware.NewAuthMiddleware(context.Background())
	if err != nil {
		log.Fatalf("failed to initialize auth middleware: %v", err)
	}

	// DI
	gameRecordRepo := persistence.NewGameRecordRepository(conn)
	gameRecordUsecase := usecase.NewGameRecordUsecase(gameRecordRepo)

	authHandler := handler.NewAuthHandler(supabaseClient)
	gameRecordHandler := handler.NewGameRecordHandler(gameRecordUsecase)

	// ルーター
	r := handler.NewRouter(authHandler, gameRecordHandler, authMiddleware)

	log.Printf("Server starting on port %s", port)
	if err := r.Run(":" + port); err != nil {
		log.Fatalf("failed to start server: %v", err)
	}
}
