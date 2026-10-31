// @title       Personal API
// @version     1.0
// @description Сервис профилей и интересов проекта «Авантюра». JWT проверяет API Gateway.
// @BasePath    /
// @securityDefinitions.apikey XUserID
// @in          header
// @name        X-User-ID
package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/VsemoGosi-Adventure/personal/internal/config"
	"github.com/VsemoGosi-Adventure/personal/internal/db"
	"github.com/VsemoGosi-Adventure/personal/internal/handler"
	"github.com/VsemoGosi-Adventure/personal/internal/migrate"
	"github.com/VsemoGosi-Adventure/personal/internal/service"
	"github.com/jackc/pgx/v5/pgxpool"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	if err := run(ctx); err != nil {
		log.Print(err)
		os.Exit(1)
	}
}

func run(ctx context.Context) error {
	cfg, err := config.Load()
	if err != nil {
		return err
	}
	pool, err := pgxpool.New(ctx, cfg.DatabaseURL)
	if err != nil {
		return fmt.Errorf("create database pool: %w", err)
	}
	defer pool.Close()
	if err := pool.Ping(ctx); err != nil {
		return fmt.Errorf("connect to database: %w", err)
	}
	count, err := migrate.Up(ctx, pool)
	if err != nil {
		return err
	}
	log.Printf("database migrations applied: %d", count)

	queries := db.New(pool)
	server := &http.Server{
		Addr:              cfg.HTTPAddress,
		Handler:           handler.NewRouter(service.NewProfileService(queries), service.NewInterestService(queries)),
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       15 * time.Second,
		WriteTimeout:      30 * time.Second,
		IdleTimeout:       60 * time.Second,
	}
	serverErr := make(chan error, 1)
	go func() { serverErr <- server.ListenAndServe() }()
	log.Printf("personal listening on %s", cfg.HTTPAddress)

	select {
	case err := <-serverErr:
		if errors.Is(err, http.ErrServerClosed) {
			return nil
		}
		return fmt.Errorf("serve HTTP: %w", err)
	case <-ctx.Done():
		shutdownCtx, cancel := context.WithTimeout(context.Background(), cfg.ShutdownTimeout)
		defer cancel()
		if err := server.Shutdown(shutdownCtx); err != nil {
			_ = server.Close()
			return fmt.Errorf("shut down HTTP server: %w", err)
		}
		if err := <-serverErr; err != nil && !errors.Is(err, http.ErrServerClosed) {
			return fmt.Errorf("serve HTTP: %w", err)
		}
		return nil
	}
}
