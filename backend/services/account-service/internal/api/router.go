package api

import (
	"github.com/daariikk/MyHelp/services/account-service/internal/api/rest/handlers"
	"github.com/daariikk/MyHelp/services/account-service/internal/config"
	"github.com/daariikk/MyHelp/services/account-service/internal/repository/postgres"
	"github.com/go-chi/chi/v5"
	"log/slog"
)

func NewRouter(cfg *config.Config, logger *slog.Logger, storage *postgres.Storage) *chi.Mux {
	router := chi.NewRouter()
	router.Use(handlers.CorsMiddleware)

	router.Route("/MyHelp/account", func(r chi.Router) {
		r.Get("/", handlers.GetPatientByIdHandler(logger, storage))
		r.Put("/", handlers.UpdatePatientInfoHandler(logger, storage))
		r.Delete("/", handlers.DeletePatientHandler(logger, storage))
	})

	return router
}
