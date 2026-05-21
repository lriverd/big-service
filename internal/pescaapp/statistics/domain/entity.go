package domain

import "time"

type SpotStats struct {
	SpotID          string       `json:"spotId"`
	Name            string       `json:"name"`
	Visits          int          `json:"visits"`
	UniqueUsers     int          `json:"uniqueUsers"`
	AverageRating   float64      `json:"averageRating"`
	TotalRatings    int          `json:"totalRatings"`
	TotalComments   int          `json:"totalComments"`
	TopSpecies      []TopSpecies `json:"topSpecies"`
	LastCommentDate *time.Time   `json:"lastCommentDate"`
	CreatedAt       time.Time    `json:"createdAt"`
}

type TopSpecies struct {
	SpeciesID string `json:"speciesId"`
	Name      string `json:"name"`
	Mentions  int    `json:"mentions"`
}

type UserDetailedStats struct {
	UserID          string   `json:"userId"`
	Name            string   `json:"name"`
	SpotsCreated    int      `json:"spotsCreated"`
	CommentsCount   int      `json:"commentsCount"`
	RatingsCount    int      `json:"ratingsCount"`
	AverageRating   float64  `json:"averageRating"`
	FavoriteRegions []string `json:"favoriteRegions"`
	FavoriteSpecies []string `json:"favoriteSpecies"`
	JoiningDate     string   `json:"joiningDate"`
}

type PopularSpot struct {
	ID            string  `json:"id"`
	Name          string  `json:"name"`
	Region        string  `json:"region"`
	AverageRating float64 `json:"averageRating"`
	TotalRatings  int     `json:"totalRatings"`
	Views         int     `json:"views"`
	TotalComments int     `json:"totalComments"`
}

