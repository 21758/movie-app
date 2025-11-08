package main

import (
	"log"
	"movie-api/config"
	"movie-api/database"
	"movie-api/handlers"
	"movie-api/middleware"
	"movie-api/services"

	"github.com/gin-gonic/gin"
)

func main() {
	cfg := config.Load()

	db, err := database.New(cfg.DBURL)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	boxOfficeService := services.NewBoxOfficeService(cfg.BoxOfficeURL, cfg.BoxOfficeAPIKey)
	movieHandler := handlers.NewMovieHandler(db, boxOfficeService)

	router := gin.Default()

	router.Use(middleware.AuthMiddleware(cfg.AuthToken))

	router.GET("/healthz", movieHandler.HealthCheck)

	movies := router.Group("/movies")
	{
		movies.GET("", movieHandler.GetMovies)
		movies.POST("", movieHandler.CreateMovie)

		ratingRoutes := movies.Group("/:title")
		ratingRoutes.Use(middleware.RaterIDMiddleware())
		{
			ratingRoutes.POST("/ratings", movieHandler.SubmitRating)
			ratingRoutes.GET("/rating", movieHandler.GetRatingAggregate)
		}
	}

	log.Printf("Server starting on port %s", cfg.Port)
	if err := router.Run("0.0.0.0:" + cfg.Port); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
