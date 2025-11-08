package models

import (
	"time"
)

type Movie struct {
	ID            string    `json:"id" gorm:"type:char(36);primaryKey"`
	Title         string    `json:"title" gorm:"uniqueIndex;not null;size:255"`
	ReleaseDate   string    `json:"releaseDate" gorm:"type:date;not null"`
	Genre         string    `json:"genre" gorm:"not null;size:100"`
	Distributor   *string   `json:"distributor" gorm:"size:255"`
	Budget        *int64    `json:"budget"`
	MpaRating     *string   `json:"mpaRating" gorm:"size:10"`
	BoxOfficeData *string   `json:"-" gorm:"type:json"`
	CreatedAt     time.Time `json:"-"`
	UpdatedAt     time.Time `json:"-"`

	BoxOffice *BoxOffice `json:"boxOffice" gorm:"-"`
}

type BoxOffice struct {
	Revenue struct {
		Worldwide         int64 `json:"worldwide"`
		OpeningWeekendUSA int64 `json:"openingWeekendUsa"`
	} `json:"revenue"`
	Currency    string    `json:"currency"`
	Source      string    `json:"source"`
	LastUpdated time.Time `json:"lastUpdated"`
}

type Rating struct {
	ID         string    `json:"id" gorm:"type:char(36);primaryKey"`
	MovieTitle string    `json:"movieTitle" gorm:"not null;size:255;index:idx_movie_rater"`
	RaterID    string    `json:"raterId" gorm:"not null;size:255;index:idx_movie_rater"`
	Rating     float64   `json:"rating" gorm:"type:decimal(3,1);not null"`
	CreatedAt  time.Time `json:"-"`
	UpdatedAt  time.Time `json:"-"`
}

type RatingAggregate struct {
	Average float64 `json:"average"`
	Count   int     `json:"count"`
}

type MoviePage struct {
	Items      []Movie `json:"items"`
	NextCursor *string `json:"nextCursor,omitempty"`
}

type MovieCreate struct {
	Title       string  `json:"title" binding:"required,min=1"`
	Genre       string  `json:"genre" binding:"required"`
	ReleaseDate string  `json:"releaseDate" binding:"required"`
	Distributor *string `json:"distributor,omitempty"`
	Budget      *int64  `json:"budget,omitempty"`
	MpaRating   *string `json:"mpaRating,omitempty"`
}

type RatingSubmit struct {
	Rating float64 `json:"rating" binding:"required"`
}
