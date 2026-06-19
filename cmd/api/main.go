package main

import (
	"context"
	"errors"
	"log"
	"log/slog"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/jackc/pgx/v5"

	"be-fittracker/internal/config"
	"be-fittracker/internal/database"
	appmiddleware "be-fittracker/internal/middleware"
	authmodule "be-fittracker/internal/modules/auth"
	mealmodule "be-fittracker/internal/modules/meal"
	progressmodule "be-fittracker/internal/modules/progress"
	reportmodule "be-fittracker/internal/modules/report"
	workoutmodule "be-fittracker/internal/modules/workout"
	"be-fittracker/internal/utils"
)

func main() {
	cfg := config.Load()
	logger := utils.NewLogger(cfg.Env)
	slog.SetDefault(logger)

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	logger.Info("connecting to postgres", databaseLogFields(cfg.DatabaseURL)...)
	conn, err := pgx.Connect(context.Background(), cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("Failed to connect to the database: %v", err)
	}
	// running test query to verify connection
	if err := conn.Ping(context.Background()); err != nil {
		log.Fatalf("Failed to ping the database: %v", err)
	}

	logger.Info("postgres connected")
	defer conn.Close(context.Background())
	db := database.NewLockedConn(conn)

	logger.Info("connecting to redis")
	redisClient, err := database.OpenRedis(ctx, cfg.RedisAddr, cfg.RedisPassword, cfg.RedisDB)
	if err != nil {
		if cfg.RedisRequired {
			logger.Error("redis connection failed", "error", err)
			os.Exit(1)
		}
		logger.Warn("redis unavailable; continuing without redis", "error", err)
	} else {
		defer redisClient.Close()
		logger.Info("redis connected")
	}

	router := chi.NewRouter()
	router.Use(middleware.RequestID)
	router.Use(middleware.RealIP)
	router.Use(appmiddleware.RequestLogger(logger))
	router.Use(appmiddleware.Recoverer(logger))
	router.Use(appmiddleware.CORS(cfg.CORSAllowedOrigins))

	router.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		utils.WriteJSON(w, http.StatusOK, map[string]string{"status": "ok"})
	})

	router.Route("/api", func(r chi.Router) {
		r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
			utils.WriteJSON(w, http.StatusOK, map[string]string{"status": "ok"})
		})

		authRepo := authmodule.NewRepository(db)
		authService := authmodule.NewService(authRepo, cfg.JWTSecret)
		authHandler := authmodule.NewHandler(authService, cfg.JWTSecret)
		r.Mount("/auth", authHandler.Routes())
		r.Mount("/users", authHandler.UserRoutes())

		mealRepo := mealmodule.NewRepository(db)
		mealService := mealmodule.NewService(mealRepo, redisClient)
		mealHandler := mealmodule.NewHandler(mealService)
		r.Mount("/meal-logs", mealHandler.Routes(cfg.JWTSecret))

		workoutRepo := workoutmodule.NewRepository(db)
		workoutService := workoutmodule.NewService(workoutRepo)
		workoutHandler := workoutmodule.NewHandler(workoutService)
		r.Mount("/workout-plans", workoutHandler.PlanRoutes(cfg.JWTSecret))
		r.Mount("/exercises", workoutHandler.ExerciseRoutes(cfg.JWTSecret))
		r.Mount("/workout-sessions", workoutHandler.SessionRoutes(cfg.JWTSecret))

		progressRepo := progressmodule.NewRepository(db)
		progressService := progressmodule.NewService(progressRepo, redisClient)
		progressHandler := progressmodule.NewHandler(progressService)
		r.Mount("/progress", progressHandler.Routes(cfg.JWTSecret))
		r.Mount("/body-measurements", progressHandler.BodyMeasurementRoutes(cfg.JWTSecret))

		reportRepo := reportmodule.NewRepository(db)
		reportService := reportmodule.NewService(reportRepo)
		reportHandler := reportmodule.NewHandler(reportService)
		r.Mount("/reports", reportHandler.Routes(cfg.JWTSecret))
	})

	server := &http.Server{
		Addr:              cfg.HTTPAddr,
		Handler:           router,
		ReadHeaderTimeout: 5 * time.Second,
	}

	go func() {
		logger.Info("api server started", "addr", cfg.HTTPAddr, "env", cfg.Env)
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			logger.Error("api server failed", "error", err)
			stop()
		}
	}()

	<-ctx.Done()

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		logger.Error("api server shutdown failed", "error", err)
		os.Exit(1)
	}

	logger.Info("api server stopped")
}

func databaseLogFields(databaseURL string) []any {
	parsed, err := url.Parse(databaseURL)
	if err != nil {
		return []any{"database_url", "invalid"}
	}

	return []any{
		"database_host", parsed.Hostname(),
		"database_port", parsed.Port(),
		"database_name", parsed.EscapedPath(),
		"database_query", parsed.RawQuery,
	}
}
