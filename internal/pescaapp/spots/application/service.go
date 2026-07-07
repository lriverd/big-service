package application

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/lriverd/big-service/internal/pescaapp/spots/domain"
	userDomain "github.com/lriverd/big-service/internal/pescaapp/users/domain"
	"github.com/lriverd/big-service/internal/platform/cache"
	apperrors "github.com/lriverd/big-service/internal/shared/errors"
	log "github.com/sirupsen/logrus"
)

type SpotService struct {
	spotRepo                    domain.SpotRepository
	speciesRepo                 domain.SpotSpeciesRepository
	cache                       *cache.Cache
	dailySpotLimitDefault       int
	limitProvider               domain.DailyLimitProvider
	duplicateSearchRadiusMeters float64
	duplicateSearchMaxResults   int
	reputation                  domain.ReputationConfig
}

func NewSpotService(
	spotRepo domain.SpotRepository,
	speciesRepo domain.SpotSpeciesRepository,
	c *cache.Cache,
	dailySpotLimitDefault int,
	limitProvider domain.DailyLimitProvider,
	duplicateSearchRadiusMeters float64,
	duplicateSearchMaxResults int,
	reputation domain.ReputationConfig,
) *SpotService {
	return &SpotService{
		spotRepo:                    spotRepo,
		speciesRepo:                 speciesRepo,
		cache:                       c,
		dailySpotLimitDefault:       dailySpotLimitDefault,
		limitProvider:               limitProvider,
		duplicateSearchRadiusMeters: duplicateSearchRadiusMeters,
		duplicateSearchMaxResults:   duplicateSearchMaxResults,
		reputation:                  reputation,
	}
}

// InvalidateSpotCache implements the moderation module's SpotCacheInvalidator
// port, letting it drop the cached copy of a spot right after auto-hiding it
// instead of waiting for the cache entry to expire.
func (s *SpotService) InvalidateSpotCache(id string) {
	s.cache.Delete(fmt.Sprintf("spot:%s", id))
	s.cache.DeleteByPrefix("spot:list")
}

// FindDuplicateCandidates looks for existing visible spots near the given
// coordinates. If radiusMetersOverride is nil, the configured default radius
// is used; otherwise the override is clamped to a sane multiple of the
// default so callers can't force an unbounded full-table scan.
func (s *SpotService) FindDuplicateCandidates(ctx context.Context, lat, lng float64, radiusMetersOverride *float64) ([]domain.DuplicateCandidate, error) {
	radius := s.duplicateSearchRadiusMeters
	if radiusMetersOverride != nil {
		radius = *radiusMetersOverride
		const maxOverrideMultiplier = 5
		if maxAllowed := s.duplicateSearchRadiusMeters * maxOverrideMultiplier; radius > maxAllowed {
			radius = maxAllowed
		}
		if radius < 1 {
			radius = 1
		}
	}
	return s.spotRepo.FindNearbyForDuplicateCheck(ctx, lat, lng, radius, s.duplicateSearchMaxResults)
}

// dailySpotLimitFor resolves the effective daily spot-creation limit for a
// user: their active override if one exists (e.g. from a penalty), otherwise
// the configured default.
func (s *SpotService) dailySpotLimitFor(ctx context.Context, userID string) (int, error) {
	override, err := s.limitProvider.GetDailySpotLimitOverride(ctx, userID)
	if err != nil {
		return 0, fmt.Errorf("resolving daily spot limit override: %w", err)
	}
	if override != nil {
		return *override, nil
	}
	return s.dailySpotLimitDefault, nil
}

func (s *SpotService) List(ctx context.Context, limit, offset int, filter domain.SpotFilter) ([]*domain.Spot, int, error) {
	spots, total, err := s.spotRepo.List(ctx, limit, offset, filter)
	if err != nil {
		return nil, 0, err
	}
	for _, spot := range spots {
		species, _ := s.speciesRepo.ListBySpot(ctx, spot.ID)
		spot.Species = species
	}
	return spots, total, nil
}

func (s *SpotService) GetByID(ctx context.Context, id string) (*domain.Spot, error) {
	cacheKey := fmt.Sprintf("spot:%s", id)
	if cached, found := s.cache.Get(cacheKey); found {
		spot := cached.(*domain.Spot)
		if !spot.IsVisible() {
			return nil, apperrors.NotFound("Spot")
		}
		return spot, nil
	}

	spot, err := s.spotRepo.FindByID(ctx, id)
	if err != nil {
		return nil, apperrors.NotFound("Spot")
	}
	if !spot.IsVisible() {
		return nil, apperrors.NotFound("Spot")
	}

	species, _ := s.speciesRepo.ListBySpot(ctx, id)
	spot.Species = species

	// Increment views asynchronously
	go func() {
		_ = s.spotRepo.IncrementViews(context.Background(), id)
	}()

	s.cache.Set(cacheKey, spot, 5*time.Minute)
	return spot, nil
}

func (s *SpotService) Create(ctx context.Context, req domain.CreateSpotRequest, userID, userEmail, userName string) (*domain.Spot, error) {
	limit, err := s.dailySpotLimitFor(ctx, userID)
	if err != nil {
		return nil, err
	}
	count, err := s.spotRepo.CountCreatedSince(ctx, userID, time.Now().UTC().Add(-24*time.Hour))
	if err != nil {
		return nil, fmt.Errorf("counting spots created in the last 24h: %w", err)
	}
	if count >= limit {
		return nil, apperrors.New(http.StatusTooManyRequests, "DAILY_LIMIT_EXCEEDED",
			fmt.Sprintf("Has alcanzado el límite de %d spots nuevos por día. Intenta nuevamente más tarde.", limit))
	}

	// Hard rule, no override: two spots can never exist within the configured
	// minimum distance of each other. This always uses the trusted
	// server-side default radius (nil override), never a client-supplied
	// one, so it can't be bypassed by asking for a tiny search radius.
	// If the check itself fails, creation fails too (fail closed) — this is
	// a data-integrity rule, not a best-effort UX nicety, so we'd rather
	// reject the request than risk letting a duplicate through.
	candidates, dupErr := s.FindDuplicateCandidates(ctx, req.Latitude, req.Longitude, nil)
	if dupErr != nil {
		return nil, fmt.Errorf("checking minimum distance to existing spots: %w", dupErr)
	}
	if len(candidates) > 0 {
		return nil, &domain.DuplicateCandidatesError{Candidates: candidates}
	}

	shortName := strings.Split(userName, " ")[0]

	spot := &domain.Spot{
		Name:            req.Name,
		Description:     req.Description,
		Latitude:        req.Latitude,
		Longitude:       req.Longitude,
		Region:          req.Region,
		WaterType:       req.WaterType,
		BoatAllowed:     req.BoatAllowed,
		BoatRequired:    req.BoatRequired,
		Access:          req.Access,
		IsFree:          req.IsFree,
		EntryFee:        req.EntryFee,
		CreatedByUserID: userID,
		CreatedByEmail:  userEmail,
		CreatedByName:   shortName,
	}

	created, err := s.spotRepo.Create(ctx, spot)
	if err != nil {
		return nil, err
	}

	if len(req.Species) > 0 {
		if err := s.speciesRepo.SetForSpot(ctx, created.ID, req.Species); err != nil {
			log.WithError(err).Warn("Failed to set species for spot")
		}
		created.Species = req.Species
	}

	s.cache.DeleteByPrefix("spot:list")
	log.WithField("spotId", created.ID).Info("Spot created")
	return created, nil
}

func (s *SpotService) Update(ctx context.Context, id string, req domain.UpdateSpotRequest, userID, role string) (*domain.Spot, error) {
	spot, err := s.spotRepo.FindByID(ctx, id)
	if err != nil {
		return nil, apperrors.NotFound("Spot")
	}

	if spot.CreatedByUserID != userID && role != "admin" {
		return nil, apperrors.Forbidden("Only the creator or admin can update this spot")
	}

	updated, err := s.spotRepo.Update(ctx, id, req)
	if err != nil {
		return nil, err
	}

	if req.Species != nil {
		if err := s.speciesRepo.SetForSpot(ctx, id, req.Species); err != nil {
			log.WithError(err).Warn("Failed to update species for spot")
		}
	}

	s.cache.Delete(fmt.Sprintf("spot:%s", id))
	s.cache.DeleteByPrefix("spot:list")
	return updated, nil
}

func (s *SpotService) Delete(ctx context.Context, id, userID, role string) error {
	spot, err := s.spotRepo.FindByID(ctx, id)
	if err != nil {
		return apperrors.NotFound("Spot")
	}
	if spot.CreatedByUserID != userID && role != "admin" {
		return apperrors.Forbidden("Only the creator or admin can delete this spot")
	}

	if err := s.speciesRepo.DeleteBySpot(ctx, id); err != nil {
		log.WithError(err).Warn("Failed to delete spot species")
	}
	if err := s.spotRepo.Delete(ctx, id); err != nil {
		return err
	}

	s.cache.Delete(fmt.Sprintf("spot:%s", id))
	s.cache.DeleteByPrefix("spot:list")
	log.WithField("spotId", id).Info("Spot deleted")
	return nil
}

func (s *SpotService) FindNearby(ctx context.Context, spotID string, radiusKm float64, limit int) ([]*domain.Spot, error) {
	spot, err := s.spotRepo.FindByID(ctx, spotID)
	if err != nil {
		return nil, apperrors.NotFound("Spot")
	}
	return s.spotRepo.FindNearby(ctx, spot.Latitude, spot.Longitude, radiusKm, limit)
}

// GetSpotBasicInfo implements the SpotInfoProvider interface for favorites
func (s *SpotService) GetSpotBasicInfo(ctx context.Context, spotID string) (*userDomain.FavoriteSpot, error) {
	spot, err := s.spotRepo.FindByID(ctx, spotID)
	if err != nil {
		return nil, err
	}
	return &userDomain.FavoriteSpot{
		ID:     spot.ID,
		Name:   spot.Name,
		Region: spot.Region,
		Rating: spot.AverageRating,
	}, nil
}

func (s *SpotService) Search(ctx context.Context, query string, limit int) ([]*domain.Spot, error) {
	return s.spotRepo.Search(ctx, query, limit)
}

// UpdateStatusByAdmin transitions a spot's moderation status (verify,
// hide/delete for cause, or restore a hidden spot back to VERIFIED) and
// records the corresponding reputation event for the spot's owner, when the
// transition is one that should affect reputation (see
// reputationEventForTransition).
func (s *SpotService) UpdateStatusByAdmin(ctx context.Context, id string, newStatus domain.SpotStatus, reason string) (*domain.Spot, error) {
	spot, err := s.spotRepo.FindByID(ctx, id)
	if err != nil {
		return nil, apperrors.NotFound("Spot")
	}
	oldStatus := spot.EffectiveStatus()

	updated, err := s.spotRepo.UpdateStatus(ctx, id, newStatus)
	if err != nil {
		return nil, fmt.Errorf("updating spot status: %w", err)
	}

	s.cache.Delete(fmt.Sprintf("spot:%s", id))
	s.cache.DeleteByPrefix("spot:list")

	if eventType, delta, ok := s.reputationEventForTransition(oldStatus, newStatus); ok {
		if err := s.reputation.Recorder.RecordReputationEvent(ctx, spot.CreatedByUserID, eventType, delta, id, reason); err != nil {
			log.WithError(err).WithField("spotId", id).Warn("Failed to record reputation event for spot status change")
		}
	}

	log.WithFields(log.Fields{"spotId": id, "from": oldStatus, "to": newStatus}).Info("Spot status updated by admin")
	return updated, nil
}

// reputationEventForTransition decides which reputation event (if any) an
// admin-driven status transition should produce. Restoring a HIDDEN/DELETED
// spot back to VERIFIED is deliberately not reputation-worthy — only an
// initial PENDING → VERIFIED verification rewards the owner.
func (s *SpotService) reputationEventForTransition(oldStatus, newStatus domain.SpotStatus) (eventType string, delta int, ok bool) {
	switch newStatus {
	case domain.SpotStatusVerified:
		if oldStatus == domain.SpotStatusPending {
			return "SPOT_VERIFIED", s.reputation.DeltaVerified, true
		}
		return "", 0, false
	case domain.SpotStatusHidden:
		return "SPOT_HIDDEN", s.reputation.DeltaHidden, true
	case domain.SpotStatusDeleted:
		return "SPOT_DELETED", s.reputation.DeltaDeleted, true
	default:
		return "", 0, false
	}
}

// GetSpotOwnerID implements the moderation module's SpotOwnerLookup port.
func (s *SpotService) GetSpotOwnerID(ctx context.Context, id string) (string, error) {
	spot, err := s.spotRepo.FindByID(ctx, id)
	if err != nil {
		return "", apperrors.NotFound("Spot")
	}
	return spot.CreatedByUserID, nil
}

// Mine returns the caller's own spots regardless of status, so an owner can
// see their PENDING/HIDDEN spots that are excluded from public listings.
func (s *SpotService) Mine(ctx context.Context, userID string, limit, offset int) ([]*domain.Spot, int, error) {
	spots, total, err := s.spotRepo.FindByCreatedByUserID(ctx, userID, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	for _, spot := range spots {
		species, _ := s.speciesRepo.ListBySpot(ctx, spot.ID)
		spot.Species = species
	}
	return spots, total, nil
}
