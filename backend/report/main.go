package main

import (
	"report/internal/config"
	"report/internal/handlers"
	"report/internal/middleware"
	"report/internal/services"
	"report/internal/storage"

	"github.com/gin-gonic/gin"
)

func main() {
	cfg := config.Load()

	clickhouseClient := storage.NewClickHouseClient(cfg.ClickHouse)
	s3Client := storage.NewS3Client(cfg.S3)

	reportService := services.NewReportService(clickhouseClient, s3Client, cfg)
	reportHandler := handlers.NewReportHandler(reportService)
	authMiddleware := middleware.NewAuthMiddleware()

	r := gin.Default()

	r.Use(func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Origin, Content-Type, Accept, Authorization")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	})

	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	api := r.Group("/api/v1")
	{
		reports := api.Group("/reports")
		{
			reports.GET("", authMiddleware.RequireAuth(), reportHandler.GetReports)
			reports.POST("/generate", authMiddleware.RequireAuth(), reportHandler.GenerateReports)
		}
	}

	if err := r.Run(":5003"); err != nil {
		panic(err)
	}
}
