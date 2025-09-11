// File: cmd/main.go
package main

import (
	"eduhub/server/api/app"
	"eduhub/server/logger"

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
// @name Authorization // Or X-Session-Token depending on your auth mechanism

func main() {
	// Load .env file FIRST
	err := godotenv.Load()
	logger := logger.NewZeroLogger(true)

	if err != nil {
		logger.Logger.Warn().Msg("unable to load  env variables")
	}

	// Create the app instance (which loads config, logger, db, etc.)
	setup := app.New()
	logger.Logger.Debug().Msg("app instance created")
	// Start the application
	err = setup.Start()
	if err != nil {
		logger.Logger.Error().Msg("failed to start application")
	}

	logger.Logger.Debug().Msg("server stopped successfully")
}
