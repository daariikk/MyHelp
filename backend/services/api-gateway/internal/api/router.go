package api

import (
	"github.com/daariikk/MyHelp/services/api-gateway/internal/api/rest/handlers"
	account_service "github.com/daariikk/MyHelp/services/api-gateway/internal/api/rest/handlers/account-service"
	appointment_service "github.com/daariikk/MyHelp/services/api-gateway/internal/api/rest/handlers/appointment-service"
	polyclinic_service "github.com/daariikk/MyHelp/services/api-gateway/internal/api/rest/handlers/polyclinic-service"
	"github.com/daariikk/MyHelp/services/api-gateway/internal/config"
	"github.com/daariikk/MyHelp/services/api-gateway/internal/repository/postgres"
	"github.com/go-chi/chi/v5"
	"log/slog"
)

func NewRouter(cfg *config.Config, logger *slog.Logger, storage *postgres.Storage) *chi.Mux {
	router := chi.NewRouter()
	router.Use(handlers.CorsMiddleware)

	router.Route("/MyHelp/auth", func(r chi.Router) {
		// Регистрация
		r.Post("/signin", handlers.RegisterHandler(logger, storage))
		// Авторизация
		r.Post("/signup", handlers.LoginHandler(logger, storage, cfg))
		// Авторизация админа
		r.Post("/signup/admin", handlers.LoginAdminHandler(logger, storage, cfg))
		// Обновление токенов
		r.Post("/refresh", handlers.RefreshHandler(logger, cfg))
		// Получить данные user-а
		r.Get("/get-user", handlers.GetUserHandler(logger, storage))
		// Получить данные admin-а
		r.Get("/get-admin", handlers.GetAdminHandler(logger, storage))
	})

	// Public specializations routes
	router.Route("/MyHelp/specializations", func(r chi.Router) {
		r.Get("/", polyclinic_service.GetPolyclinic(logger, cfg))
		r.Get("/{specializationID}", polyclinic_service.GetDoctorsBySpecialization(logger, cfg))
	})

	// Public schedule routes
	router.Route("/MyHelp/schedule/doctors", func(r chi.Router) {
		r.Get("/{doctorID}", polyclinic_service.GetSchedule(logger, cfg))
	})

	// Protected routes
	router.Group(func(r chi.Router) {
		r.Use(handlers.AuthMiddleware(logger, cfg))

		r.Route("/MyHelp/account", func(r chi.Router) {
			r.Get("/", account_service.GetPatient(logger, cfg))
			r.Put("/", account_service.UpdatePatient(logger, cfg))
			r.Delete("/", account_service.DeletePatient(logger, cfg))
		})

		r.Route("/MyHelp/schedule/appointments", func(r chi.Router) {
			r.Post("/", appointment_service.NewAppointment(logger, cfg))
			r.Patch("/{appointmentID}", appointment_service.UpdateAppointment(logger, cfg))
			r.Delete("/{appointmentID}", appointment_service.DeleteAppointment(logger, cfg))
		})

		// Add admin-only operations under a different path or as subroutes
		r.Route("/MyHelp/admin/specializations", func(r chi.Router) {
			r.Post("/", polyclinic_service.NewSpecialization(logger, cfg))
			r.Delete("/{specializationID}", polyclinic_service.DeleteSpecialization(logger, cfg))
		})

		r.Route("/MyHelp/admin/doctors", func(r chi.Router) {
			r.Post("/", polyclinic_service.NewDoctor(logger, cfg))
			r.Delete("/{doctorID}", polyclinic_service.DeleteDoctor(logger, cfg))
		})

		r.Route("/MyHelp/admin/schedule/doctors", func(r chi.Router) {
			r.Post("/{doctorID}", polyclinic_service.NewSchedule(logger, cfg))
		})
	})

	return router
}
