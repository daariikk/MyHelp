package api

import (
	"github.com/daariikk/MyHelp/services/api-gateway/internal/api/rest/handlers"
	account_service "github.com/daariikk/MyHelp/services/api-gateway/internal/api/rest/handlers/account-service"
	appointment_service "github.com/daariikk/MyHelp/services/api-gateway/internal/api/rest/handlers/appointment-service"
	polyclinic_service "github.com/daariikk/MyHelp/services/api-gateway/internal/api/rest/handlers/polyclinic-service"
	"github.com/daariikk/MyHelp/services/api-gateway/internal/config"
	"github.com/daariikk/MyHelp/services/api-gateway/internal/repository/postgres"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"log/slog"
)

func NewRouter(cfg *config.Config, logger *slog.Logger, storage *postgres.Storage) *chi.Mux {
	router := chi.NewRouter()
	router.Use(handlers.CorsMiddleware)
	router.Use(middleware.Logger)
	router.Use(middleware.Recoverer)

	// Auth routes (public)
	router.Route("/MyHelp/auth", func(r chi.Router) {
		r.Post("/signin", handlers.RegisterHandler(logger, storage))
		r.Post("/signup", handlers.LoginHandler(logger, storage, cfg))
		r.Post("/signup/admin", handlers.LoginAdminHandler(logger, storage, cfg))
		r.Post("/refresh", handlers.RefreshHandler(logger, cfg))
		r.Get("/get-user", handlers.GetUserHandler(logger, storage))
		r.Get("/get-admin", handlers.GetAdminHandler(logger, storage))
	})

	// Specializations routes
	router.Route("/MyHelp/specializations", func(r chi.Router) {
		// Public
		r.Get("/", polyclinic_service.GetPolyclinic(logger, cfg))
		r.Get("/{specializationID}", polyclinic_service.GetDoctorsBySpecialization(logger, cfg))

		// Protected (require auth)
		r.Group(func(r chi.Router) {
			r.Use(handlers.AuthMiddleware(logger, cfg))
			r.Post("/", polyclinic_service.NewSpecialization(logger, cfg))
			r.Delete("/{specializationID}", polyclinic_service.DeleteSpecialization(logger, cfg))
		})
	})

	// Doctors routes
	router.Route("/MyHelp/doctors", func(r chi.Router) {
		// Protected only (no public methods)
		r.Group(func(r chi.Router) {
			r.Use(handlers.AuthMiddleware(logger, cfg))
			r.Post("/", polyclinic_service.NewDoctor(logger, cfg))
			r.Delete("/{doctorID}", polyclinic_service.DeleteDoctor(logger, cfg))
		})
	})

	// Schedule routes
	router.Route("/MyHelp/schedule", func(r chi.Router) {
		// Public
		r.Route("/doctors", func(r chi.Router) {
			r.Get("/{doctorID}", polyclinic_service.GetSchedule(logger, cfg))
		})

		// Protected appointments
		r.Route("/appointments", func(r chi.Router) {
			r.Group(func(r chi.Router) {
				r.Use(handlers.AuthMiddleware(logger, cfg))
				r.Post("/", appointment_service.NewAppointment(logger, cfg))
				r.Patch("/{appointmentID}", appointment_service.UpdateAppointment(logger, cfg))
				r.Delete("/{appointmentID}", appointment_service.DeleteAppointment(logger, cfg))
			})
		})

		// Protected admin schedule management
		r.Route("/admin/doctors", func(r chi.Router) {
			r.Group(func(r chi.Router) {
				r.Use(handlers.AuthMiddleware(logger, cfg))
				r.Post("/{doctorID}", polyclinic_service.NewSchedule(logger, cfg))
			})
		})
	})

	// Account routes (protected only)
	router.Route("/MyHelp/account", func(r chi.Router) {
		r.Group(func(r chi.Router) {
			r.Use(handlers.AuthMiddleware(logger, cfg))
			r.Get("/", account_service.GetPatient(logger, cfg))
			r.Put("/", account_service.UpdatePatient(logger, cfg))
			r.Delete("/", account_service.DeletePatient(logger, cfg))
		})
	})

	return router
}
