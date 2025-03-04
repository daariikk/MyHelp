package api

import (
	"github.com/daariikk/MyHelp/services/api-gateway/internal/api/rest/handlers"
	"github.com/daariikk/MyHelp/services/api-gateway/internal/config"
	"github.com/daariikk/MyHelp/services/api-gateway/internal/repository/postgres"
	"github.com/go-chi/chi/v5"
	"log/slog"
)

func NewRouter(cfg *config.Config, logger *slog.Logger, storage *postgres.Storage) *chi.Mux {
	router := chi.NewRouter()
	router.Use(handlers.CorsMiddleware)

	router.Route("/api/v1/auth", func(r chi.Router) {
		r.Post("/signin", handlers.RegisterHandler(logger, storage))
		r.Post("/signup", handlers.LoginHandler(logger, storage, cfg))
		r.Post("/refresh", handlers.RefreshHandler(logger, cfg))
		// r.Post("/reset-password", handlers.ResetHandler(logger, cfg))
	})

	return router
}
