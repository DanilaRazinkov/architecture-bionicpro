package main

import (
	"auth/internal/config"
	"auth/internal/handlers"
	"auth/internal/middleware"
	"auth/internal/services"
	"auth/internal/storage"
	"log"

	"github.com/gin-gonic/gin"
)

func main() {
	cfg := config.Load()

	redisClient := storage.NewRedisClient(cfg.Redis)
	keycloakService := services.NewKeycloakService(&cfg.Keycloak)
	sessionService := services.NewSessionService(redisClient, keycloakService)

	authHandler := handlers.NewAuthHandler(sessionService, keycloakService)
	authMiddleware := middleware.NewAuthMiddleware(sessionService)

	r := gin.Default()

	r.Use(func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "http://localhost:3000")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Origin, Content-Type, Accept, Authorization")
		c.Header("Access-Control-Allow-Credentials", "true")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	})

	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	r.GET("/login", authHandler.Login)

	api := r.Group("/api")
	{
		auth := api.Group("/auth")
		{
			auth.GET("/status", authHandler.GetAuthStatus)
			auth.GET("/login", authHandler.Login)
			auth.GET("/callback", authHandler.HandleCallback)
			auth.GET("/me", authMiddleware.RequireAuth(), authHandler.GetUserInfo)
			auth.POST("/refresh", authMiddleware.RequireAuth(), authHandler.RefreshToken)
		}

		reports := api.Group("/reports")
		{
			reports.GET("", authMiddleware.RequireAuth(), authHandler.GetReports)
			reports.POST("/generate", authMiddleware.RequireAuth(), authHandler.GenerateReports)
		}
	}

	log.Println("Starting bionicpro-auth server on :5001")
	if err := r.Run(":5001"); err != nil {
		log.Fatal("Failed to start server:", err)
	}
}
