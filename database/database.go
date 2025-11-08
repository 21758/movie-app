package database

import (
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"movie-api/models"
	"time"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

type DB struct {
	*gorm.DB
}

func New(connectionString string) (*DB, error) {
	db, err := gorm.Open(mysql.Open(connectionString), &gorm.Config{})
	if err != nil {
		return nil, err
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, err
	}

	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetMaxOpenConns(100)
	sqlDB.SetConnMaxLifetime(time.Hour)

	if err := runMigrations(db); err != nil {
		return nil, err
	}

	return &DB{db}, nil
}

func generateUUID() string {
	const chars = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	result := make([]byte, 36)
	for i := range result {
		result[i] = chars[rand.Intn(len(chars))]
	}
	return string(result)
}

func runMigrations(db *gorm.DB) error {
	err := db.AutoMigrate(
		&models.Movie{},
		&models.Rating{},
	)
	if err != nil {
		return err
	}

	log.Println("Database migrations completed successfully")
	return nil
}

func (db *DB) CreateMovie(movie *models.Movie) error {
	movie.ID = generateUUID()
	return db.Create(movie).Error
}

func (db *DB) GetMovieByTitle(title string) (*models.Movie, error) {
	var movie models.Movie
	err := db.Where("title = ?", title).First(&movie).Error
	if err != nil {
		return nil, err
	}

	if movie.BoxOfficeData != nil {
		var boxOffice models.BoxOffice
		if err := json.Unmarshal([]byte(*movie.BoxOfficeData), &boxOffice); err == nil {
			movie.BoxOffice = &boxOffice
		}
	}

	return &movie, nil
}

func (db *DB) UpdateMovieBoxOffice(title string, boxOffice *models.BoxOffice) error {
	boxOfficeData, err := json.Marshal(boxOffice)
	if err != nil {
		return err
	}

	boxOfficeStr := string(boxOfficeData)
	return db.Model(&models.Movie{}).
		Where("title = ?", title).
		Update("box_office_data", &boxOfficeStr).Error
}

func (db *DB) UpsertRating(rating *models.Rating) (bool, error) {
	var existingRating models.Rating
	result := db.Where("movie_title = ? AND rater_id = ?", rating.MovieTitle, rating.RaterID).
		First(&existingRating)

	if result.Error == gorm.ErrRecordNotFound {
		rating.ID = generateUUID()
		return true, db.Create(rating).Error
	} else if result.Error != nil {
		return false, result.Error
	}

	rating.ID = existingRating.ID
	rating.CreatedAt = existingRating.CreatedAt
	return false, db.Save(rating).Error
}

func (db *DB) GetRatingAggregate(movieTitle string) (*models.RatingAggregate, error) {
	var aggregate models.RatingAggregate

	var result struct {
		Average float64
		Count   int64
	}

	err := db.Model(&models.Rating{}).
		Select("ROUND(AVG(rating), 1) as average, COUNT(*) as count").
		Where("movie_title = ?", movieTitle).
		Scan(&result).Error

	if err != nil {
		return nil, err
	}

	aggregate.Average = result.Average
	aggregate.Count = int(result.Count)

	return &aggregate, nil
}

func (db *DB) SearchMovies(params map[string]interface{}, limit int, cursor string) (*models.MoviePage, error) {
	query := db.Model(&models.Movie{})

	if q, ok := params["q"].(string); ok && q != "" {
		query = query.Where("title LIKE ?", "%"+q+"%")
	}

	if genre, ok := params["genre"].(string); ok && genre != "" {
		query = query.Where("genre LIKE ?", genre)
	}

	if year, ok := params["year"].(int); ok && year != 0 {
		query = query.Where("YEAR(release_date) = ?", year)
	}

	if distributor, ok := params["distributor"].(string); ok && distributor != "" {
		query = query.Where("distributor LIKE ?", distributor)
	}

	if budget, ok := params["budget"].(int64); ok && budget != 0 {
		query = query.Where("budget <= ?", budget)
	}

	if mpaRating, ok := params["mpaRating"].(string); ok && mpaRating != "" {
		query = query.Where("mpa_rating = ?", mpaRating)
	}

	var movies []models.Movie
	offset := 0

	if cursor != "" {
		fmt.Sscanf(cursor, "offset_%d", &offset)
	}

	err := query.Order("created_at DESC").
		Offset(offset).
		Limit(limit + 1).
		Find(&movies).Error

	if err != nil {
		return nil, err
	}

	for i := range movies {
		if movies[i].BoxOfficeData != nil {
			var boxOffice models.BoxOffice
			if err := json.Unmarshal([]byte(*movies[i].BoxOfficeData), &boxOffice); err == nil {
				movies[i].BoxOffice = &boxOffice
			}
		}
	}

	var nextCursor *string
	if len(movies) > limit {
		movies = movies[:limit]
		nextCursorValue := fmt.Sprintf("offset_%d", offset+limit)
		nextCursor = &nextCursorValue
	}

	return &models.MoviePage{
		Items:      movies,
		NextCursor: nextCursor,
	}, nil
}

func (db *DB) HealthCheck() error {
	sqlDB, err := db.DB.DB()
	if err != nil {
		return err
	}
	return sqlDB.Ping()
}

func (db *DB) Close() error {
	sqlDB, err := db.DB.DB()
	if err != nil {
		return err
	}
	return sqlDB.Close()
}
