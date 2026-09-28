package main

import (
	"log"
	"net/http"

	"github.com/mdflamingo/fgo-tracker-backend/internal/config"
	"github.com/mdflamingo/fgo-tracker-backend/internal/handler"
	"github.com/mdflamingo/fgo-tracker-backend/internal/logger"
	pg "github.com/mdflamingo/fgo-tracker-backend/internal/repository/postgres"
	"github.com/mdflamingo/fgo-tracker-backend/internal/validator"
	"go.uber.org/zap"

	_ "github.com/golang-migrate/migrate/v4/source/file"
)

// @title Task Tracker API
// @version 1.0.0
// @description API for task tracker
// @termsOfService http://swagger.io/terms/
// @contact.name API Support
// @contact.email support@example.com
// @host localhost:8080
// @BasePath /
// @securitydefinitions.oauth2.password OAuth2Keycloak
// @tokenUrl http://localhost:8080/auth/realms/fmgo/protocol/openid-connect/token
// @security OAuth2Keycloak
func main() {
	conf := config.GetConfig()
	if err := run(conf); err != nil {
		log.Fatal(err)
	}
	logger.Log.Info("Server shutdown gracefully")
}

func run(conf *config.Config) error {
	if err := logger.Initialize(conf.LogLevel); err != nil {
		return err
	}

	validator.Init()

	logger.Log.Info("Running server", zap.String("address", conf.RunAddr))

	pgStorage, errStorage := pg.ConnectPG(&conf.DataBaseDSN)
	if errStorage != nil {
		logger.Log.Fatal("Failed to create storage", zap.Error(errStorage))
	}
	defer pgStorage.Close()

	r := handler.NewRouter(conf, pgStorage)

	return http.ListenAndServe(conf.RunAddr, r)
}
