package api

import (
	"github.com/daariikk/MyHelp/services/polyclinic-service/internal/api/rest/handlers"
	"github.com/daariikk/MyHelp/services/polyclinic-service/internal/config"
	"github.com/daariikk/MyHelp/services/polyclinic-service/internal/repository/postgres"
	"github.com/daariikk/MyHelp/services/polyclinic-service/internal/use_cases"
	"github.com/go-chi/chi/v5"
	"log/slog"
)

func NewRouter(cfg *config.Config, logger *slog.Logger, storage *postgres.Storage, scheduleUseCase use_cases.NewScheduleWrapper) *chi.Mux {
	router := chi.NewRouter()
	router.Use(handlers.CorsMiddleware)

	router.Route("/MyHelp/specializations", func(r chi.Router) {
		// Получить список специализаций
		r.Get("/", handlers.GetPolyclinicInfoHandler(logger, storage))

		// Получить список врачей определенной специализации
		r.Get("/{specializationID}", handlers.GetSpecializationDoctorHandler(logger, storage))

		// Создать новую специализацию
		r.Post("/", handlers.CreateNewSpecializationHandler(logger, storage))

		// Удалить специализацию (и всех врачей этой специализации)
		r.Delete("/{specializationID}", handlers.DeleteSpecializationHandler(logger, storage))

	})

	router.Route("/MyHelp/doctors", func(r chi.Router) {
		// Создать нового врача
		r.Post("/", handlers.NewDoctorHandler(logger, storage))

		// Удалить врача
		r.Delete("/{doctorID}", handlers.DeleteDoctorHandler(logger, storage))

	})

	router.Route("/MyHelp/schedule/doctors", func(r chi.Router) {
		// Получить расписание врача
		r.Get("/{doctorID}", handlers.GetScheduleDoctorByIdHandler(logger, storage))

		// Добавить расписание
		r.Post("/{doctorID}", handlers.NewScheduleHandler(logger, storage, scheduleUseCase))
	})

	return router
}
