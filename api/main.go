package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	httpSwagger "github.com/swaggo/http-swagger"

	_ "workflow-code-test/api/docs"
	"workflow-code-test/api/pkg/db"
	"workflow-code-test/api/services/workflow"
)

// @title Workflow Engine API
// @version 1.0
// @description A Go-based API for managing and executing workflow automations.
// @description Provides endpoints to retrieve workflow definitions, execute workflows, and view execution history.

// @contact.name API Support
// @contact.url https://github.com/albert-saclot/workflow-go-challenge

// @license.name MIT
// @license.url https://opensource.org/licenses/MIT

// @host localhost:8080
// @BasePath /api/v1

// @schemes http

func main() {
	ctx := context.Background()
	logHandler := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	})
	slog.SetDefault(slog.New(logHandler))

	dbURL, ok := os.LookupEnv("DATABASE_URL")
	if !ok {
		slog.Error("DATABASE_URL is not set")
		return
	}

	pool, err := db.Connect(ctx, dbURL)
	if err != nil {
		slog.Error("Failed to connect to database", "error", err)
		return
	}
	defer pool.Close()

	// setup router
	mainRouter := mux.NewRouter()

	apiRouter := mainRouter.PathPrefix("/api/v1").Subrouter()

	workflowService, err := workflow.NewService(pool)
	if err != nil {
		slog.Error("Failed to create workflow service", "error", err)
		return
	}

	workflowService.LoadRoutes(apiRouter)

	// Swagger documentation endpoint
	mainRouter.PathPrefix("/swagger/").Handler(httpSwagger.WrapHandler)

	corsHandler := handlers.CORS(
		// Frontend URL
		handlers.AllowedOrigins([]string{"http://localhost:3003"}),
		handlers.AllowedMethods([]string{"GET", "POST", "PUT", "DELETE", "OPTIONS"}),
		handlers.AllowedHeaders([]string{"Content-Type", "Authorization"}),
		handlers.AllowCredentials(),
	)(mainRouter)

	srv := &http.Server{
		Addr:    ":8080",
		Handler: corsHandler,
	}

	serverErrors := make(chan error, 1)

	go func() {
		slog.Info("Starting server on :8080")
		serverErrors <- srv.ListenAndServe()
	}()

	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, os.Interrupt, syscall.SIGTERM)

	select {
	case err := <-serverErrors:
		slog.Error("Server error", "error", err)

	case sig := <-shutdown:
		slog.Info("Shutdown signal received", "signal", sig)

		ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
		defer cancel()

		if err := srv.Shutdown(ctx); err != nil {
			slog.Error("Could not stop server gracefully", "error", err)
			srv.Close()
		}
	}
}
