package main

import (
	"context"
	"log"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"service/internal/config"
	"service/internal/handler"
	"service/internal/middleware"
	"service/internal/store"
	"syscall"
	"time"

	"github.com/go-chi/chi"
	"github.com/joho/godotenv"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func main() {
	// env load
	if err := godotenv.Load(); err != nil {
		log.Println("Файл конфигурации не обнаружен")
	}
	// cfg load
	cfg, err := config.Load()
	if err != nil {
		log.Printf("ошибка загрузки конфига %s", err)
		os.Exit(1)
	}

	middleware.InitMetrics()
	ctx := context.Background()
	// db load
	pool, err := store.NewPool(ctx, cfg)
	if err != nil {
		log.Printf("Ошибка подключения к БД: %v", err)
		os.Exit(1)
	}
	defer pool.Close()

	// db migrate
	if err := store.Migrate(ctx, pool); err != nil {
		log.Printf("Ошибка миграции: %v", err)
		os.Exit(1)
	}

	// redis load

	redisClient, err := store.NewRedisClient(ctx, cfg)

	if err != nil {
		slog.Error("Редиска не законнектилась", "error", err)
		os.Exit(1)
	}

	defer redisClient.Close()

	noteStore := store.NewNoteStore(pool)
	cachedStore := store.NewCachedStore(noteStore, redisClient)
	noteHandler := handler.NewNoteHandler(cachedStore)

	r := chi.NewRouter()

	r.Use(middleware.Logger)
	r.Use(middleware.MetricsMiddleware)

	r.Post("/notes", noteHandler.Create)
	r.Get("/notes", noteHandler.GetAll)
	r.Get("/notes/{id}", noteHandler.GetByID)
	r.Get("/metrics", promhttp.Handler().ServeHTTP)

	srv := &http.Server{
		Addr:         ":8080",
		Handler:      r,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  15 * time.Second,
	}

	go func() {
		slog.Info("server starting", "port", ":8080")
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			slog.Error("listen error", "error", err)
			os.Exit(1)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGABRT)
	sig := <-quit
	slog.Info("shutting down", "sygnal", sig.String())

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		slog.Error("server down error", "error", err)
	}

	pool.Close()

	slog.Info("server gracefully stopped")
}
