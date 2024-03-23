package main

import (
	"L0WB/internal/cache"
	"L0WB/internal/config"
	"L0WB/internal/http-server/handlers/order"
	"L0WB/internal/http-server/middleware/logger"
	"L0WB/internal/stan"
	"L0WB/internal/storage/postgres"
	"context"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	ostan "github.com/nats-io/stan.go"
	"github.com/rs/cors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

const (
	envDev     = "dev"                // уровень логирования
	envProd    = "prod"               // уровень логирования
	configPath = "config/config.yaml" //путь до конфиг файла
)

func main() {
	// инициализация конфига
	cfg := config.MustLoad(configPath)

	// инициализация логгера
	log := SetupLogger(cfg.Env)

	log.Info("App started", slog.String("env", cfg.Env))
	log.Debug("Debugging started")

	// инициализация БД
	storage, err := postgres.New(
		cfg.Postgres.Host,
		cfg.Postgres.Port,
		cfg.Postgres.User,
		cfg.Postgres.Password,
		cfg.Postgres.DBName,
	)
	if err != nil {
		log.Error("failed to init storage: ", err)
		os.Exit(1)
	}
	defer storage.Close()

	log.Info("storage successfully initialized")

	client, err := stan.NewClient(
		cfg.Nats.ClusterId,
		cfg.Nats.ClientId,
		cfg.Nats.Url,
	)
	if err != nil {
		log.Error("failed to init connection to nats: ", err)
		os.Exit(1)
	}
	defer client.Close()

	// подписка на канал
	go func() {
		_, err = client.Subscribe("orders", func(m *ostan.Msg) {
			log.Info("New order to NATS!")
			stan.OrderMessage(log, storage, m)
		})
		if err != nil {
			log.Error("failed to subscribe to nats chanel: ", err)
			os.Exit(1)
		}
	}()

	log.Info("nats successfully connected and listening")

	err = cache.Init(log, storage)
	if err != nil {
		log.Error("failed to init cache: ", err)
	}

	log.Info("cache successfully init")

	// инициализация роутера и настройка мидлвейров
	router := chi.NewRouter()

	// CORS для локального запуска и тестирования
	corsHandler := cors.New(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowedMethods:   []string{"GET", "POST", "DELETE", "PUT", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: true,
		MaxAge:           300,
	})

	router.Use(middleware.RequestID)
	router.Use(middleware.Logger)
	router.Use(logger.New(log))
	router.Use(middleware.Recoverer)
	router.Use(middleware.URLFormat)
	router.Use(corsHandler.Handler)

	// REST маршруты
	router.Get("/api/order/{id}", order.GetOrder(log))

	log.Info("starting server", slog.String("address", cfg.Address))

	// настройка и запуск сервера
	srv := &http.Server{
		Addr:         cfg.Address,
		Handler:      router,
		ReadTimeout:  cfg.HTTPServer.Timeout,
		WriteTimeout: cfg.HTTPServer.Timeout,
		IdleTimeout:  cfg.HTTPServer.IdleTimeout,
	}

	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Error("failed to start server: ", err)
		}
	}()

	// Graceful shutdown
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)

	<-stop

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Error("shutdown error: ", err)
	} else {
		log.Info("server stopped gracefully")
	}
}

func SetupLogger(env string) *slog.Logger {
	var log *slog.Logger

	switch env {
	case envDev:
		log = slog.New(
			slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}),
		)

	case envProd:
		log = slog.New(
			slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}),
		)
	}

	return log
}
