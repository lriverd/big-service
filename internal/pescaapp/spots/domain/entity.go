package domain

import "time"

// SpotStatus models the moderation lifecycle of a Spot.
type SpotStatus string

const (
	SpotStatusPending  SpotStatus = "PENDING"
	SpotStatusVerified SpotStatus = "VERIFIED"
	SpotStatusHidden   SpotStatus = "HIDDEN"
	SpotStatusDeleted  SpotStatus = "DELETED"
)

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
	CreatedByEmail  string        `json:"-" firestore:"createdByEmail"`
	CreatedByName   string        `json:"createdByName" firestore:"createdByName"`
	Views           int           `json:"views" firestore:"views"`
	AverageRating   float64       `json:"averageRating" firestore:"averageRating"`
	TotalRatings    int           `json:"totalRatings" firestore:"totalRatings"`
	TotalComments   int           `json:"totalComments" firestore:"totalComments"`
	Status          string        `json:"status" firestore:"status"`
	ReportCount     int           `json:"reportCount" firestore:"reportCount"`
	Species         []SpotSpecies `json:"species,omitempty" firestore:"-"`
	CreatedAt       time.Time     `json:"createdAt" firestore:"createdAt"`
	UpdatedAt       time.Time     `json:"updatedAt" firestore:"updatedAt"`
}

// EffectiveStatus returns the spot's moderation status. Spots persisted
// before this field existed have an empty Status; those are treated as
// PENDING rather than VERIFIED, since an empty/legacy status has never
// actually gone through any verification step.
func (s *Spot) EffectiveStatus() SpotStatus {
	if s.Status == "" {
		return SpotStatusPending
	}
	return SpotStatus(s.Status)
}

// IsVisible reports whether the spot should appear in public listings/detail.
// HIDDEN and DELETED spots are excluded; PENDING and VERIFIED are shown.
func (s *Spot) IsVisible() bool {
	switch s.EffectiveStatus() {
	case SpotStatusHidden, SpotStatusDeleted:
		return false
	default:
		return true
	}
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

// DuplicateCandidate is an existing spot found near a proposed new spot's
// coordinates, together with the distance between them.
type DuplicateCandidate struct {
	Spot           *Spot   `json:"spot"`
	DistanceMeters float64 `json:"distanceMeters"`
}

// DuplicateCandidatesError signals that spot creation was rejected because
// an existing spot was found within the minimum allowed distance of the
// proposed location. This is a hard rule with no override — two spots can
// never exist closer than the configured minimum distance apart.
type DuplicateCandidatesError struct {
	Candidates []DuplicateCandidate
}

func (e *DuplicateCandidatesError) Error() string {
	return "an existing spot is too close to the proposed location"
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

// UpdateSpotStatusRequest is the admin-only request to transition a spot's
// moderation status (verify, hide/delete for cause, or restore).
type UpdateSpotStatusRequest struct {
	Status SpotStatus `json:"status" binding:"required"`
	Reason string     `json:"reason"`
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
