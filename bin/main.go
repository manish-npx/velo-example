package main

import (
	"github.com/manish-npx/velo/config"
	"github.com/manish-npx/velo/core"
	http "github.com/manish-npx/velo/http"
	"github.com/manish-npx/velo/logger"
)

func main() {
	// Load configuration
	cfg := config.NewConfig(".env")
	appConfig := cfg.GetAppConfig()

	// Initialize logger
	log := logger.NewSimpleLogger(logger.LevelInfo)
	defer log.Close()

	// Create app instance
	configMap := map[string]interface{}{
		"name": appConfig.Name,
		"env":  appConfig.Env,
		"port": appConfig.Port,
	}
	app := core.New(configMap, log)

	// Setup routes
	router := http.NewRouter(app)
	router.Get("/", func(c *http.Context) error {
		return c.JSON(map[string]string{
			"message": "Welcome to Velo Framework!",
		})
	})

	// Start server
	if err := app.Start(appConfig.Port); err != nil {
		log.Error("Failed to start server", err)
	}
}
