package main

import (
	router "core/api/router"
	config "core/internal/config"
	database "core/internal/database"
	gmail "core/pkg/gmail"
	logger "core/pkg/log"
	openai "core/pkg/openai"
	"net/http"
	"time"

	"github.com/ory/graceful"
)

//	@title			Motif API
//	@version		1.0
//	@description	Acquisition sourcing: campaigns and ICP, leads, AI research, inbound deals and drafts.
//	@BasePath		/api/v1
func main() {
	logger.Start()

	cfg, err := config.Load()
	logger.Fatal(err, "Failed to load env")

	openai.Configure(&cfg.OpenAI)
	gmail.Configure(&cfg.Gmail)

	gormDB := database.Init(&cfg.Database)
	sqlDB, err := gormDB.DB()
	logger.Fatal(err, "Failed to get sql db")
	defer sqlDB.Close()

	server := graceful.WithDefaults(&http.Server{
		Addr:         cfg.Server.Port,
		Handler:      router.Config(gormDB, cfg),
		IdleTimeout:  time.Minute,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 30 * time.Second,
	})

	logger.Info("Starting Motif API on %s", cfg.Server.Port)
	err = graceful.Graceful(server.ListenAndServe, server.Shutdown)
	logger.Fatal(err, "Server shutdown error")
}
