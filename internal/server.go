package internal

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/Koenigseder/aoc-leaderboard-scraper/internal/clients/aoc"
)

const (
	envVarMainPort   = "MAIN_PORT"
	envVarHealthPort = "HEALTH_PORT"
)

func SetupAndRunServer() {
	// Generic health response
	healthRes := struct {
		Msg string `json:"msg"`
	}{
		Msg: "ok",
	}

	// ------------------- Main router -------------------
	mainMux := http.NewServeMux()

	mainMux.HandleFunc("GET /getLocalScoreHistory/{eventYear}/{leaderboardId}", aoc.GetHistoricScoreFile)
	mainMux.HandleFunc("POST /persistCurrentLocalScores/{eventYear}/{leaderboardId}", aoc.PersistCurrentLocalScores)

	mainPort, ok := os.LookupEnv(envVarMainPort)
	if !ok {
		slog.Error(fmt.Sprintf("failed looking up env var '%s'", envVarMainPort))
		os.Exit(1)
	}

	mainServer := &http.Server{
		Addr:              ":" + mainPort,
		Handler:           mainMux,
		ReadTimeout:       5 * time.Second,  //nolint:mnd
		ReadHeaderTimeout: 5 * time.Second,  //nolint:mnd
		WriteTimeout:      10 * time.Second, //nolint:mnd
		IdleTimeout:       15 * time.Second, //nolint:mnd
	}

	// Graceful shutdown logic
	mainDone := make(chan bool)
	mainQuit := make(chan os.Signal, 1)
	signal.Notify(mainQuit, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-mainQuit
		slog.Info("main server is shutting down")

		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second) //nolint:mnd
		defer cancel()

		mainServer.SetKeepAlivesEnabled(false)

		if err := mainServer.Shutdown(ctx); err != nil {
			slog.Error("could not gracefully shutdown the main server", "error", err)
			os.Exit(1)
		}

		close(mainDone)
	}()

	// ------------------- Health router -------------------
	healthMux := http.NewServeMux()

	healthMux.HandleFunc("GET /healthz", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(healthRes)
	})

	healthPort, ok := os.LookupEnv(envVarHealthPort)
	if !ok {
		slog.Error(fmt.Sprintf("failed looking up env var '%s'", envVarHealthPort))
		os.Exit(1)
	}

	healthServer := &http.Server{
		Addr:              ":" + healthPort,
		Handler:           healthMux,
		ReadTimeout:       5 * time.Second,  //nolint:mnd
		ReadHeaderTimeout: 5 * time.Second,  //nolint:mnd
		WriteTimeout:      10 * time.Second, //nolint:mnd
		IdleTimeout:       15 * time.Second, //nolint:mnd
	}

	// Graceful shutdown logic
	healthDone := make(chan bool)
	healthQuit := make(chan os.Signal, 1)
	signal.Notify(healthQuit, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-healthQuit
		slog.Info("health server is shutting down")

		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second) //nolint:mnd
		defer cancel()

		healthServer.SetKeepAlivesEnabled(false)

		if err := healthServer.Shutdown(ctx); err != nil {
			slog.Error("could not gracefully shutdown the health server", "error", err)
			os.Exit(1)
		}

		close(healthDone)
	}()

	// ------------------- Start servers -------------------
	var wg sync.WaitGroup

	wg.Add(2) //nolint:mnd

	// Start main server
	go func() {
		defer wg.Done()

		if err := mainServer.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			slog.Error("failed starting main server", "error", err)
			os.Exit(1)
		}

		<-mainDone
		slog.Info("main server gracefully shut down")
	}()

	// Start health sever
	go func() {
		defer wg.Done()

		if err := healthServer.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			slog.Error("failed starting health server", "error", err)
			os.Exit(1)
		}

		<-healthDone
		slog.Info("health server gracefully shut down")
	}()

	slog.Info("application started")

	wg.Wait()
}
