package domain

import "time"

type Species struct {
	ID              string    `json:"id" firestore:"-"`
	CommonName      string    `json:"commonName" firestore:"commonName"`
	ScientificName  string    `json:"scientificName" firestore:"scientificName"`
	Description     string    `json:"description" firestore:"description"`
	ImageURL        *string   `json:"imageUrl" firestore:"imageUrl"`
	Habitat         string    `json:"habitat" firestore:"habitat"`
	Diet            string    `json:"diet" firestore:"diet"`
	AverageSizeCm   float64   `json:"averageSizeCm" firestore:"averageSizeCm"`
	AverageWeightKg float64   `json:"averageWeightKg" firestore:"averageWeightKg"`
	MaxSizeCm       float64   `json:"maxSizeCm" firestore:"maxSizeCm"`
	MaxWeightKg     float64   `json:"maxWeightKg" firestore:"maxWeightKg"`
	FishingTips     []string  `json:"fishingTips" firestore:"fishingTips"`
	CreatedAt       time.Time `json:"createdAt" firestore:"createdAt"`
	UpdatedAt       time.Time `json:"updatedAt" firestore:"updatedAt"`
}

type CreateSpeciesRequest struct {
	CommonName      string   `json:"commonName" binding:"required"`
	ScientificName  string   `json:"scientificName" binding:"required"`
	Description     string   `json:"description" binding:"required"`
	ImageURL        *string  `json:"imageUrl"`
	Habitat         string   `json:"habitat"`
	Diet            string   `json:"diet"`
	AverageSizeCm   float64  `json:"averageSizeCm"`
	AverageWeightKg float64  `json:"averageWeightKg"`
	MaxSizeCm       float64  `json:"maxSizeCm"`
	MaxWeightKg     float64  `json:"maxWeightKg"`
	FishingTips     []string `json:"fishingTips"`
}

type UpdateSpeciesRequest struct {
	CommonName      *string  `json:"commonName"`
	ScientificName  *string  `json:"scientificName"`
	Description     *string  `json:"description"`
	ImageURL        *string  `json:"imageUrl"`
	Habitat         *string  `json:"habitat"`
	Diet            *string  `json:"diet"`
	AverageSizeCm   *float64 `json:"averageSizeCm"`
	AverageWeightKg *float64 `json:"averageWeightKg"`
	MaxSizeCm       *float64 `json:"maxSizeCm"`
	MaxWeightKg     *float64 `json:"maxWeightKg"`
	FishingTips     []string `json:"fishingTips"`
}

