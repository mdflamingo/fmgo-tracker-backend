package handler

import (
	"log"

	_ "github.com/mdflamingo/fgo-tracker-backend/api/swagger"

	"github.com/go-chi/chi/v5"
	"github.com/mdflamingo/fgo-tracker-backend/internal/apiresponse"
	"github.com/mdflamingo/fgo-tracker-backend/internal/auth"
	"github.com/mdflamingo/fgo-tracker-backend/internal/config"
	"github.com/mdflamingo/fgo-tracker-backend/internal/logger"
	pg "github.com/mdflamingo/fgo-tracker-backend/internal/repository/postgres"
	"github.com/mdflamingo/fgo-tracker-backend/internal/service"
	httpSwagger "github.com/swaggo/http-swagger"
)

func NewRouter(conf *config.Config, storage *pg.DBStorage) *chi.Mux {
	r := chi.NewRouter()

	oidcValidator, err := auth.NewOIDCValidator(conf.Keycloak.KeycloakURL, conf.Keycloak.ClientID)
	if err != nil {
		log.Fatalf("Failed to initialize OIDC validator: %v", err)
	}

	taskService := service.NewTaskService(storage)
	userService := service.NewUserService(storage)
	projectService := service.NewProjectService(storage)

	taskHandler := NewTaskHandler(taskService)
	userHandler := NewUserHandler(userService)
	projectHandler := NewProjectHandler(projectService)

	r.Use(logger.RequestLogger)
	r.Use(apiresponse.Recoverer)

	// // Swagger documentation
	r.Get("/swagger/*", httpSwagger.Handler(
		httpSwagger.URL("/swagger/doc.json"),
		httpSwagger.PersistAuthorization(true),
	))

	r.Route("/api", func(r chi.Router) {
		r.Use(oidcValidator.AuthMiddleware())
		r.Route("/task", func(r chi.Router) {
			r.Post("/", taskHandler.CreateTask)
			r.Get("/list", taskHandler.GetList)
			r.Get("/{id}", taskHandler.GetTask)
			r.Put("/{id}", taskHandler.UpdateTask)
			r.Delete("/{id}", taskHandler.DeleteTask)
		})

		r.Route("/user", func(r chi.Router) {
			r.Get("/list", userHandler.GetList)
		})

		r.Route("/project", func(r chi.Router) {
			r.Get("/list", projectHandler.GetList)
			r.Post("/", projectHandler.CreateProject)
		})

	})

	return r
}
