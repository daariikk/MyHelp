package helper

import (
	"io"
	"log/slog"
	"net/http"
)

func CopyHeaders(dst, src http.Header) {
	for key, values := range src {
		for _, value := range values {
			dst.Set(key, value)
		}
	}
}

func ForwardRequest(logger *slog.Logger, w http.ResponseWriter, r *http.Request, url string, method string) {
	// Создаем новый запрос с указанным методом и телом

	req, err := http.NewRequest(method, url, r.Body)
	if err != nil {
		logger.Error("Failed to create request", slog.String("error", err.Error()))
		http.Error(w, "Failed to create request", http.StatusInternalServerError)
		return
	}
	logger.Debug("Новый запрос с указанным методом и телом успешно создан")

	CopyHeaders(req.Header, r.Header)
	logger.Debug("Заголовки успешно скопированы")

	req.URL.RawQuery = r.URL.RawQuery
	logger.Debug("Query-параметры успешно скопированы")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		logger.Error("Failed to forward request", slog.String("error", err.Error()))
		http.Error(w, "Failed to forward request", http.StatusInternalServerError)
		return
	}
	logger.Debug("Запрос прошел успешно")
	defer resp.Body.Close()

	CopyHeaders(w.Header(), resp.Header)
	w.WriteHeader(resp.StatusCode)

	_, err = io.Copy(w, resp.Body)
	if err != nil {
		logger.Error("Failed to copy response", slog.String("error", err.Error()))
		http.Error(w, "Failed to copy response", http.StatusInternalServerError)
		return
	}
}
