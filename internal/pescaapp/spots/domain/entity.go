package domain

import "time"

type Spot struct {
	ID              string        `json:"id" firestore:"-"`
	Name            string        `json:"name" firestore:"name"`
	Description     string        `json:"description" firestore:"description"`
	Latitude        float64       `json:"latitude" firestore:"latitude"`
	Longitude       float64       `json:"longitude" firestore:"longitude"`
	Region          string        `json:"region" firestore:"region"`
	WaterType       string        `json:"waterType" firestore:"waterType"`
	BoatAllowed     bool          `json:"boatAllowed" firestore:"boatAllowed"`
	BoatRequired    bool          `json:"boatRequired" firestore:"boatRequired"`
	Access          string        `json:"access" firestore:"access"`
	IsFree          bool          `json:"isFree" firestore:"isFree"`
	EntryFee        *float64      `json:"entryFee,omitempty" firestore:"entryFee,omitempty"`
	CreatedByUserID string        `json:"createdByUserId" firestore:"createdByUserId"`
	Views           int           `json:"views" firestore:"views"`
	AverageRating   float64       `json:"averageRating" firestore:"averageRating"`
	TotalRatings    int           `json:"totalRatings" firestore:"totalRatings"`
	TotalComments   int           `json:"totalComments" firestore:"totalComments"`
	Species         []SpotSpecies `json:"species,omitempty" firestore:"-"`
	CreatedAt       time.Time     `json:"createdAt" firestore:"createdAt"`
	UpdatedAt       time.Time     `json:"updatedAt" firestore:"updatedAt"`
}

type SpotSpecies struct {
	ID               string   `json:"id,omitempty" firestore:"-"`
	SpotID           string   `json:"spotId,omitempty" firestore:"spotId"`
	SpeciesID        string   `json:"speciesId" firestore:"speciesId"`
	RecommendedBaits []string `json:"recommendedBaits" firestore:"recommendedBaits"`
	RecommendedRod   string   `json:"recommendedRod" firestore:"recommendedRod"`
	RecommendedLine  string   `json:"recommendedLine" firestore:"recommendedLine"`
	RecommendedHook  string   `json:"recommendedHook" firestore:"recommendedHook"`
	BestSeason       string   `json:"bestSeason" firestore:"bestSeason"`
	Difficulty       string   `json:"difficulty" firestore:"difficulty"`
	Notes            string   `json:"notes" firestore:"notes"`
}

type CreateSpotRequest struct {
	Name         string        `json:"name" binding:"required"`
	Description  string        `json:"description" binding:"required"`
	Latitude     float64       `json:"latitude" binding:"required"`
	Longitude    float64       `json:"longitude" binding:"required"`
	Region       string        `json:"region" binding:"required"`
	WaterType    string        `json:"waterType"`
	BoatAllowed  bool          `json:"boatAllowed"`
	BoatRequired bool          `json:"boatRequired"`
	Access       string        `json:"access"`
	IsFree       bool          `json:"isFree"`
	EntryFee     *float64      `json:"entryFee,omitempty"`
	Species      []SpotSpecies `json:"species"`
}

type UpdateSpotRequest struct {
	Name         *string       `json:"name"`
	Description  *string       `json:"description"`
	Latitude     *float64      `json:"latitude"`
	Longitude    *float64      `json:"longitude"`
	Region       *string       `json:"region"`
	WaterType    *string       `json:"waterType"`
	BoatAllowed  *bool         `json:"boatAllowed"`
	BoatRequired *bool         `json:"boatRequired"`
	Access       *string       `json:"access"`
	IsFree       *bool         `json:"isFree"`
	EntryFee     *float64      `json:"entryFee"`
	Species      []SpotSpecies `json:"species"`
}

type SpotFilter struct {
	Region       string
	WaterType    string
	BoatRequired *bool
	Latitude     *float64
	Longitude    *float64
	RadiusKm     *float64
	SortBy       string
}

