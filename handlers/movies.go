package handlers

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	"movie-api/database"
	"movie-api/models"
	"movie-api/services"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type MovieHandler struct {
	db               *database.DB
	boxOfficeService *services.BoxOfficeService
}

func NewMovieHandler(db *database.DB, boxOfficeService *services.BoxOfficeService) *MovieHandler {
	return &MovieHandler{
		db:               db,
		boxOfficeService: boxOfficeService,
	}
}

func (h *MovieHandler) CreateMovie(c *gin.Context) {
	var createReq models.MovieCreate

	if err := c.ShouldBindJSON(&createReq); err != nil {
		c.JSON(http.StatusUnprocessableEntity, gin.H{
			"code":    "BAD_REQUEST",
			"message": "Invalid request body: " + err.Error(),
		})
		return
	}

	if _, err := time.Parse("2006-01-02", createReq.ReleaseDate); err != nil {
		c.JSON(http.StatusUnprocessableEntity, gin.H{
			"code":    "BAD_REQUEST",
			"message": "Invalid release date format, expected YYYY-MM-DD",
		})
		return
	}

	existingMovie, err := h.db.GetMovieByTitle(createReq.Title)
	if err == nil && existingMovie != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    "BAD_REQUEST",
			"message": "Movie with this title already exists",
		})
		return
	}

	movie := &models.Movie{
		Title:       createReq.Title,
		Genre:       createReq.Genre,
		ReleaseDate: createReq.ReleaseDate,
		Distributor: createReq.Distributor,
		Budget:      createReq.Budget,
		MpaRating:   createReq.MpaRating,
	}

	if err := h.db.CreateMovie(movie); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    "INTERNAL_ERROR",
			"message": "Failed to create movie",
		})
		return
	}

	go h.enrichMovieWithBoxOfficeData(movie)

	c.Header("Location", "/movies/"+movie.Title)
	c.JSON(http.StatusCreated, movie)
}

func (h *MovieHandler) enrichMovieWithBoxOfficeData(movie *models.Movie) {
	boxOfficeData, err := h.boxOfficeService.GetMovieData(movie.Title)
	if err != nil {
		fmt.Printf("Failed to fetch box office data for %s: %v\n", movie.Title, err)
		return
	}

	if err := h.db.UpdateMovieBoxOffice(movie.Title, boxOfficeData); err != nil {
		fmt.Printf("Failed to update box office data for %s: %v\n", movie.Title, err)
	}
}

func (h *MovieHandler) GetMovies(c *gin.Context) {
	params := make(map[string]interface{})

	if q := c.Query("q"); q != "" {
		params["q"] = q
	}
	if genre := c.Query("genre"); genre != "" {
		params["genre"] = genre
	}
	if year := c.Query("year"); year != "" {
		if yearInt, err := strconv.Atoi(year); err == nil {
			params["year"] = yearInt
		}
	}
	if distributor := c.Query("distributor"); distributor != "" {
		params["distributor"] = distributor
	}
	if budget := c.Query("budget"); budget != "" {
		if budgetInt, err := strconv.ParseInt(budget, 10, 64); err == nil {
			params["budget"] = budgetInt
		}
	}
	if mpaRating := c.Query("mpaRating"); mpaRating != "" {
		params["mpaRating"] = mpaRating
	}

	limit := 10
	if limitStr := c.Query("limit"); limitStr != "" {
		if limitInt, err := strconv.Atoi(limitStr); err == nil && limitInt > 0 {
			limit = limitInt
		}
	}

	cursor := c.Query("cursor")

	movies, err := h.db.SearchMovies(params, limit, cursor)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    "INTERNAL_ERROR",
			"message": "Failed to search movies",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"items":      movies.Items,
		"nextCursor": movies.NextCursor,
	})
}

func (h *MovieHandler) SubmitRating(c *gin.Context) {
	title := c.Param("title")
	raterID := c.GetString("rater_id")

	var ratingReq models.RatingSubmit

	if err := c.ShouldBindJSON(&ratingReq); err != nil {
		c.JSON(http.StatusUnprocessableEntity, gin.H{
			"code":    "BAD_REQUEST",
			"message": "Invalid rating value",
		})
		return
	}

	validRatings := []float64{0.5, 1.0, 1.5, 2.0, 2.5, 3.0, 3.5, 4.0, 4.5, 5.0}
	valid := false
	for _, v := range validRatings {
		if ratingReq.Rating == v {
			valid = true
			break
		}
	}
	if !valid {
		c.JSON(http.StatusUnprocessableEntity, gin.H{
			"code":    "BAD_REQUEST",
			"message": "Rating must be one of: 0.5, 1.0, 1.5, 2.0, 2.5, 3.0, 3.5, 4.0, 4.5, 5.0",
		})
		return
	}

	_, err := h.db.GetMovieByTitle(title)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{
				"code":    "NOT_FOUND",
				"message": "Movie not found",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    "INTERNAL_ERROR",
			"message": "Failed to check movie existence",
		})
		return
	}

	rating := &models.Rating{
		MovieTitle: title,
		RaterID:    raterID,
		Rating:     ratingReq.Rating,
	}

	isNew, err := h.db.UpsertRating(rating)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    "INTERNAL_ERROR",
			"message": "Failed to submit rating",
		})
		return
	}

	response := gin.H{
		"movieTitle": rating.MovieTitle,
		"raterId":    rating.RaterID,
		"rating":     rating.Rating,
	}

	if isNew {
		c.Header("Location", "/movies/"+title+"/ratings")
		c.JSON(http.StatusCreated, response)
	} else {
		c.JSON(http.StatusOK, response)
	}
}

func (h *MovieHandler) GetRatingAggregate(c *gin.Context) {
	title := c.Param("title")

	_, err := h.db.GetMovieByTitle(title)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{
				"code":    "NOT_FOUND",
				"message": "Movie not found",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    "INTERNAL_ERROR",
			"message": "Failed to check movie existence",
		})
		return
	}

	aggregate, err := h.db.GetRatingAggregate(title)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    "INTERNAL_ERROR",
			"message": "Failed to get rating aggregate",
		})
		return
	}

	c.JSON(http.StatusOK, aggregate)
}

func (h *MovieHandler) HealthCheck(c *gin.Context) {
	if err := h.db.HealthCheck(); err != nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"status": "unhealthy",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status": "healthy",
	})
}
