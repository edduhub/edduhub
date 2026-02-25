// File: cmd/main.go
package main

import (
	"context"
	"eduhub/server/api/app"
	"eduhub/server/logger"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/joho/godotenv"
)

// @title           EduHub API
// @version         1.0
// @description     API for the EduHub platform.
// @termsOfService  http://swagger.io/terms/

// @contact.name   API Support
// @contact.url    http://www.swagger.io/support
// @contact.email  support@swagger.io

// @license.name  Apache 2.0
// @license.url   http://www.apache.org/licenses/LICENSE-2.0.html

// @host      localhost:8080 // Change if needed
// @BasePath  /

// @securityDefinitions.apikey ApiKeyAuth
// @in header
// @name Authorization

func main() {
	// Load .env file FIRST
	err := godotenv.Load()
	log := logger.NewZeroLogger(true)

	if err != nil {
		log.Logger.Warn().Msg("unable to load  env variables")
	}

	// Create the app instance (which loads config, logger, db, etc.)
	setup, err := app.New()
	if err != nil {
		log.Logger.Fatal().Err(err).Msg("failed to create app instance")
		return
	}
	log.Logger.Debug().Msg("app instance created")

	// Setup graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// Start the application in a goroutine
	errChan := make(chan error, 1)
	go func() {
		if err := setup.Start(); err != nil {
			errChan <- err
		}
	}()

	// Wait for shutdown signal or error
	select {
	case <-sigChan:
		log.Logger.Info().Msg("shutdown signal received, gracefully shutting down...")
	case err := <-errChan:
		log.Logger.Fatal().Err(err).Msg("failed to start application")
		return
	}

	// Shutdown the application
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer shutdownCancel()

	if err := setup.Shutdown(shutdownCtx); err != nil {
		log.Logger.Error().Err(err).Msg("error during shutdown")
	} else {
		log.Logger.Info().Msg("server stopped successfully")
	}
}
