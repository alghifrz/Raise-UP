package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/diuk/raiseup/config"
	db "github.com/diuk/raiseup/db/generated"
	"github.com/diuk/raiseup/internal/activity"
	"github.com/diuk/raiseup/internal/announcement"
	"github.com/diuk/raiseup/internal/auth"
	"github.com/diuk/raiseup/internal/bot"
	"github.com/diuk/raiseup/internal/chat"
	"github.com/diuk/raiseup/internal/complaint"
	"github.com/diuk/raiseup/internal/dashboard"
	"github.com/diuk/raiseup/internal/dues"
	"github.com/diuk/raiseup/internal/finance"
	"github.com/diuk/raiseup/internal/gallery"
	"github.com/diuk/raiseup/internal/resident"
	sitesettings "github.com/diuk/raiseup/internal/site_settings"
	supabasestorage "github.com/diuk/raiseup/internal/storage"
	"github.com/diuk/raiseup/internal/village"
	"github.com/diuk/raiseup/internal/whatsapp"
	"github.com/diuk/raiseup/middleware"
	"github.com/diuk/raiseup/pkg/database"
	jwtutil "github.com/diuk/raiseup/pkg/jwt"
	"github.com/diuk/raiseup/pkg/logger"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
)

func main() {
	log := logger.New()

	cfg, err := config.Load()
	if err != nil {
		log.Error("failed to load config", "error", err)
		os.Exit(1)
	}

	ctx, cancelApp := context.WithCancel(context.Background())
	defer cancelApp()

	pool, err := database.NewPool(ctx, cfg.DatabaseURL)
	if err != nil {
		log.Error("failed to connect to database", "error", err)
		os.Exit(1)
	}
	defer pool.Close()

	log.Info("database connection established")
	if cfg.WhatsApp.Enabled {
		log.Info("whatsapp cloud api enabled")
	} else {
		log.Info("whatsapp cloud api disabled")
	}
	if cfg.WhatsApp.BotEnabled {
		log.Info("whatsapp public bot enabled")
	}

	router := setupRouter(ctx, cfg, log, pool)

	srv := &http.Server{
		Addr:              fmt.Sprintf(":%s", cfg.Port),
		Handler:           router,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       15 * time.Second,
		WriteTimeout:      15 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	go func() {
		log.Info("starting HTTP server", "port", cfg.Port)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Error("HTTP server error", "error", err)
			os.Exit(1)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	sig := <-quit

	log.Info("shutdown signal received", "signal", sig.String())
	cancelApp()

	shutdownCtx, cancel := context.WithTimeout(context.Background(), cfg.ShutdownTimeout)
	defer cancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Error("graceful shutdown failed", "error", err)
		os.Exit(1)
	}

	log.Info("server stopped")
}

func setupRouter(appCtx context.Context, cfg *config.Config, log *slog.Logger, pool *pgxpool.Pool) *gin.Engine {
	gin.SetMode(gin.ReleaseMode)

	tokens := jwtutil.NewManager(cfg.JWTSecret, cfg.JWTExpiresIn)
	authRepo := auth.NewRepository(pool)
	authService := auth.NewService(authRepo, tokens)
	authHandler := auth.NewHandler(authService, log)

	residentRepo := resident.NewRepository(pool)
	residentService := resident.NewService(residentRepo)
	residentHandler := resident.NewHandler(residentService, log)

	complaintRepo := complaint.NewRepository(pool)
	complaintService := complaint.NewService(
		complaintRepo,
		complaint.AdaptResidentReader(residentRepo.GetByID, func(err error) bool {
			return errors.Is(err, resident.ErrNotFound)
		}),
	)
	complaintHandler := complaint.NewHandler(complaintService, log)

	announcementRepo := announcement.NewRepository(pool)
	announcementService := announcement.NewService(
		announcementRepo,
		announcement.AdaptResidentReader(residentRepo.GetByID, func(err error) bool {
			return errors.Is(err, resident.ErrNotFound)
		}),
	)
	announcementHandler := announcement.NewHandler(announcementService, log)

	financeRepo := finance.NewRepository(pool)
	financeService := finance.NewService(financeRepo)
	financeHandler := finance.NewHandler(financeService, log)

	duesRepo := dues.NewRepository(pool)
	duesService := dues.NewService(
		duesRepo,
		dues.AdaptResidentReader(residentRepo.GetByID, func(err error) bool {
			return errors.Is(err, resident.ErrNotFound)
		}),
	)
	duesHandler := dues.NewHandler(duesService, log)

	activityRepo := activity.NewRepository(pool)
	activityService := activity.NewService(activityRepo)
	activityHandler := activity.NewHandler(activityService, log)

	galleryRepo := gallery.NewRepository(pool)
	galleryService := gallery.NewService(galleryRepo)
	if cfg.Supabase.IsConfigured() {
		galleryService.SetObjectStorage(
			supabasestorage.NewClient(cfg.Supabase.URL, cfg.Supabase.ServiceRoleKey),
			cfg.Supabase.GalleryBucket,
		)
		log.Info("supabase gallery storage enabled", "bucket", cfg.Supabase.GalleryBucket)
	} else {
		log.Info("supabase gallery storage disabled")
	}
	galleryHandler := gallery.NewHandler(
		galleryService,
		log,
		cfg.Supabase.GalleryMaxUploadBytes,
	)

	siteSettingsRepo := sitesettings.NewRepository(pool)
	siteSettingsService := sitesettings.NewService(siteSettingsRepo)
	siteSettingsHandler := sitesettings.NewHandler(siteSettingsService, log)

	villageRepo := village.NewRepository(pool)
	villageService := village.NewService(villageRepo)
	villageHandler := village.NewHandler(villageService, log)

	dashboardRepo := dashboard.NewRepository(pool)
	dashboardService := dashboard.NewService(dashboardRepo)
	dashboardHandler := dashboard.NewHandler(dashboardService, log)

	waMessenger := whatsapp.NewMessengerAdapter(cfg.WhatsApp)
	chatRepo := chat.NewRepository(pool)
	chatService := chat.NewService(chatRepo, waMessenger, whatsapp.PhoneHelper{})
	announcementService.SetMessenger(chatService)
	duesService.SetMessenger(chatService)
	chatHandler := chat.NewHandler(chatService, log)

	reminderScheduler := activity.NewReminderScheduler(activityRepo, residentRepo, chatService, log)
	go reminderScheduler.Run(appCtx)

	botSessionStore := bot.NewPostgresSessionStore(pool)
	botService := bot.NewService(
		chatService,
		waMessenger,
		announcementService,
		activityService,
		financeService,
		complaintService,
		botSessionStore,
		cfg.WhatsApp.BotEnabled,
		log,
	)
	waService := whatsapp.NewService(cfg.WhatsApp, waMessenger, chatService, botService, log)
	waHandler := whatsapp.NewHandler(waService, log)

	r := gin.New()
	r.Use(gin.Recovery())
	r.Use(middleware.RequestLogger(log))
	r.Use(middleware.CORS(cfg.CORSAllowedOrigins))

	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	protected := []gin.HandlerFunc{
		middleware.RequireAuth(tokens),
		middleware.RequireRole(db.UserRoleSUPERADMIN, db.UserRoleADMINRW),
	}

	api := r.Group("/api/v1")
	authHandler.RegisterRoutes(api, middleware.RequireAuth(tokens))

	// Public portal (read-only)
	siteSettingsHandler.RegisterPublicRoutes(api)
	villageHandler.RegisterPublicRoutes(api)
	announcementHandler.RegisterPublicRoutes(api)
	galleryHandler.RegisterPublicRoutes(api)
	activityHandler.RegisterPublicRoutes(api)

	// Public WhatsApp webhook (Meta)
	waHandler.RegisterWebhookRoutes(api)

	residentHandler.RegisterRoutes(api, protected...)
	complaintHandler.RegisterRoutes(api, protected...)
	announcementHandler.RegisterRoutes(api, protected...)
	financeHandler.RegisterRoutes(api, protected...)
	duesHandler.RegisterRoutes(api, protected...)
	activityHandler.RegisterRoutes(api, protected...)
	galleryHandler.RegisterRoutes(api, protected...)
	siteSettingsHandler.RegisterRoutes(api, protected...)
	villageHandler.RegisterRoutes(api, protected...)
	dashboardHandler.RegisterRoutes(api, protected...)
	chatHandler.RegisterRoutes(api, protected...)
	waHandler.RegisterRoutes(api, protected...)

	return r
}
