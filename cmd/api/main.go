package main

import (
	"context"
	"fmt"
	"time"

	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"

	// Platform
	"github.com/lriverd/big-service/internal/platform/cache"
	"github.com/lriverd/big-service/internal/platform/config"
	fb "github.com/lriverd/big-service/internal/platform/firebase"
	"github.com/lriverd/big-service/internal/platform/logger"
	"github.com/lriverd/big-service/internal/platform/middleware"

	// Auth
	authApp "github.com/lriverd/big-service/internal/pescaapp/auth/application"
	authInfra "github.com/lriverd/big-service/internal/pescaapp/auth/infrastructure"
	authIface "github.com/lriverd/big-service/internal/pescaapp/auth/interfaces"

	// Users
	usersApp "github.com/lriverd/big-service/internal/pescaapp/users/application"
	usersInfra "github.com/lriverd/big-service/internal/pescaapp/users/infrastructure"
	usersIface "github.com/lriverd/big-service/internal/pescaapp/users/interfaces"

	// Species
	speciesApp "github.com/lriverd/big-service/internal/pescaapp/species/application"
	speciesInfra "github.com/lriverd/big-service/internal/pescaapp/species/infrastructure"
	speciesIface "github.com/lriverd/big-service/internal/pescaapp/species/interfaces"

	// Spots
	spotsApp "github.com/lriverd/big-service/internal/pescaapp/spots/application"
	spotsInfra "github.com/lriverd/big-service/internal/pescaapp/spots/infrastructure"
	spotsIface "github.com/lriverd/big-service/internal/pescaapp/spots/interfaces"

	// Comments
	commentsApp "github.com/lriverd/big-service/internal/pescaapp/comments/application"
	commentsInfra "github.com/lriverd/big-service/internal/pescaapp/comments/infrastructure"
	commentsIface "github.com/lriverd/big-service/internal/pescaapp/comments/interfaces"

	// Ratings
	ratingsApp "github.com/lriverd/big-service/internal/pescaapp/ratings/application"
	ratingsInfra "github.com/lriverd/big-service/internal/pescaapp/ratings/infrastructure"
	ratingsIface "github.com/lriverd/big-service/internal/pescaapp/ratings/interfaces"

	// Search
	searchApp "github.com/lriverd/big-service/internal/pescaapp/search/application"
	searchIface "github.com/lriverd/big-service/internal/pescaapp/search/interfaces"

	// Statistics
	statsApp "github.com/lriverd/big-service/internal/pescaapp/statistics/application"
	statsIface "github.com/lriverd/big-service/internal/pescaapp/statistics/interfaces"

	// Moderation
	moderationApp "github.com/lriverd/big-service/internal/pescaapp/moderation/application"
	moderationInfra "github.com/lriverd/big-service/internal/pescaapp/moderation/infrastructure"
	moderationIface "github.com/lriverd/big-service/internal/pescaapp/moderation/interfaces"

	// Reputation
	reputationApp "github.com/lriverd/big-service/internal/pescaapp/reputation/application"
	reputationInfra "github.com/lriverd/big-service/internal/pescaapp/reputation/infrastructure"
	reputationIface "github.com/lriverd/big-service/internal/pescaapp/reputation/interfaces"
	spotsDomain "github.com/lriverd/big-service/internal/pescaapp/spots/domain"
)

func main() {
	// Load configuration
	cfg := config.Load()

	// Setup logger
	logger.Setup(cfg.LogLevel, cfg.Environment)

	// Initialize Firebase
	ctx := context.Background()
	fbClient, err := fb.NewClient(ctx, cfg.FirebaseProjectID, cfg.CredentialsJSON)
	if err != nil {
		log.WithError(err).Fatal("Failed to initialize Firebase")
	}
	defer fbClient.Close()

	// Initialize cache
	appCache := cache.New(5*time.Minute, 10*time.Minute)

	// ========== REPOSITORIES ==========
	userRepo := usersInfra.NewUserRepository(fbClient.Firestore)
	favRepo := usersInfra.NewFavoriteRepository(fbClient.Firestore)
	speciesRepo := speciesInfra.NewSpeciesRepository(fbClient.Firestore)
	spotRepo := spotsInfra.NewSpotRepository(fbClient.Firestore)
	spotSpeciesRepo := spotsInfra.NewSpotSpeciesRepository(fbClient.Firestore)
	commentRepo := commentsInfra.NewCommentRepository(fbClient.Firestore)
	likeRepo := commentsInfra.NewCommentLikeRepository(fbClient.Firestore)
	ratingRepo := ratingsInfra.NewRatingRepository(fbClient.Firestore)
	reportRepo := moderationInfra.NewReportRepository(fbClient.Firestore)
	reputationRepo := reputationInfra.NewReputationRepository(fbClient.Firestore)
	penaltyRepo := reputationInfra.NewPenaltyRepository(fbClient.Firestore)

	// ========== ADAPTERS ==========
	googleValidator := authInfra.NewGoogleTokenValidator(fbClient.Auth, cfg.GoogleClientID)
	userInfoAdapter := commentsInfra.NewUserInfoAdapter(userRepo)

	// ========== SERVICES ==========
	userService := usersApp.NewUserService(userRepo, appCache)
	penaltyEvaluator := reputationApp.NewPenaltyEvaluator(
		reputationRepo, userService, penaltyRepo, userService,
		cfg.PenaltyRejectedContentRateThreshold,
		cfg.PenaltyRejectedContentMinSampleSize,
		cfg.ReputationDeltaRejectedContentPenalty,
		cfg.PenaltyDailyLimitReducedValue,
		time.Duration(cfg.PenaltyDailyLimitDurationHours)*time.Hour,
	)
	reputationService := reputationApp.NewReputationService(reputationRepo, userService, userService, penaltyEvaluator)
	spotService := spotsApp.NewSpotService(spotRepo, spotSpeciesRepo, appCache, cfg.DailySpotLimitDefault, userService, cfg.DuplicateSearchRadiusMeters, cfg.DuplicateSearchMaxResults, spotsDomain.ReputationConfig{
		Recorder:      reputationService,
		DeltaVerified: cfg.ReputationDeltaSpotVerified,
		DeltaHidden:   cfg.ReputationDeltaSpotHidden,
		DeltaDeleted:  cfg.ReputationDeltaSpotDeleted,
	})
	favService := usersApp.NewFavoriteService(favRepo, spotService)
	authService := authApp.NewAuthService(googleValidator, userRepo, fbClient.Auth, cfg.JWTSecret, cfg.JWTExpiryMinutes, cfg.RegistrationEnabled)
	speciesService := speciesApp.NewSpeciesService(speciesRepo, appCache)
	commentService := commentsApp.NewCommentService(commentRepo, likeRepo, userInfoAdapter, spotRepo)
	ratingService := ratingsApp.NewRatingService(ratingRepo, spotRepo, reputationService, cfg.ReputationDeltaGoodRating, cfg.ReputationGoodRatingStarsThreshold)
	searchService := searchApp.NewSearchService(spotRepo, speciesRepo, userRepo)
	statsService := statsApp.NewStatsService(fbClient.Firestore, spotRepo, userRepo, appCache)
	reportService := moderationApp.NewReportService(reportRepo, spotService, spotService, reputationService, cfg.ReportHideThreshold, cfg.ReputationDeltaSpotHidden)

	// ========== HANDLERS ==========
	authHandler := authIface.NewAuthHandler(authService)
	userHandler := usersIface.NewUserHandler(userService)
	favHandler := usersIface.NewFavoriteHandler(favService)
	speciesHandler := speciesIface.NewSpeciesHandler(speciesService)
	spotHandler := spotsIface.NewSpotHandler(spotService)
	commentHandler := commentsIface.NewCommentHandler(commentService)
	ratingHandler := ratingsIface.NewRatingHandler(ratingService)
	searchHandler := searchIface.NewSearchHandler(searchService)
	statsHandler := statsIface.NewStatsHandler(statsService)
	reportHandler := moderationIface.NewReportHandler(reportService)
	reputationHandler := reputationIface.NewReputationHandler(reputationService, penaltyEvaluator)

	// ========== GIN ENGINE ==========
	if cfg.Environment == "production" {
		gin.SetMode(gin.ReleaseMode)
	}
	r := gin.New()

	// Global middleware
	authMw := middleware.NewAuthMiddleware(cfg)
	rateLimiter := middleware.NewRateLimiter(appCache, cfg.RateLimitPerMin)

	r.Use(middleware.RequestID())
	r.Use(middleware.AccessLogger())
	r.Use(middleware.Recovery())
	r.Use(middleware.ErrorHandler())
	r.Use(middleware.CORSMiddleware(cfg.AllowedOrigins))
	r.Use(rateLimiter.Limit())

	// Health check
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok", "service": "pescaapp", "timestamp": time.Now().UTC()})
	})

	// API v1
	v1 := r.Group("/v1")

	// ========== REGISTER ROUTES ==========
	authIface.RegisterRoutes(v1, authHandler, authMw)
	usersIface.RegisterRoutes(v1, userHandler, favHandler, authMw)
	speciesIface.RegisterRoutes(v1, speciesHandler, authMw)
	spotsIface.RegisterRoutes(v1, spotHandler, authMw)
	commentsIface.RegisterRoutes(v1, commentHandler, authMw)
	ratingsIface.RegisterRoutes(v1, ratingHandler, authMw)
	searchIface.RegisterRoutes(v1, searchHandler)
	statsIface.RegisterRoutes(v1, statsHandler)
	moderationIface.RegisterRoutes(v1, reportHandler, authMw)
	reputationIface.RegisterRoutes(v1, reputationHandler, authMw)

	// Start server
	addr := fmt.Sprintf(":%s", cfg.Port)
	log.WithField("port", cfg.Port).Info("Starting PescaApp server")
	if err := r.Run(addr); err != nil {
		log.WithError(err).Fatal("Server failed to start")
	}
}
