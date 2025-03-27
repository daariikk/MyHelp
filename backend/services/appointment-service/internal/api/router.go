package api

import (
	"github.com/daariikk/MyHelp/services/appointment-service/internal/api/rest/handlers"
	"github.com/daariikk/MyHelp/services/appointment-service/internal/repository/postgres"
	"github.com/go-chi/chi/v5"
	"log/slog"
)

func NewRouter(logger *slog.Logger, storage *postgres.Storage) *chi.Mux {
	router := chi.NewRouter()
	router.Use(handlers.CorsMiddleware)

	router.Route("/MyHelp/schedule/appointments", func(r chi.Router) {
		// Создать запись к врачу
		r.Post("/", handlers.CreateAppointmentHandler(logger, storage))

		// Получить записи к врачу по id пациента patientID
		//r.Get("/", handlers.GetAppointmentsHandler(logger, storage))

		// Изменить запись к врачу
		r.Patch("/{appointmentID}", handlers.UpdateAppointmentHandler(logger, storage))

		// Удалить/отменить запись к врачу
		r.Delete("/{appointmentID}", handlers.CancelAppointmentHandler(logger, storage))
	})

	return router
}
