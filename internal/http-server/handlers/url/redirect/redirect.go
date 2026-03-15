package redirect

import (
	"errors"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/render"
	"log/slog"
	"net/http"
	resp "shorter/internal/lib/api/response"
	"shorter/internal/lib/logger/sl"
	"shorter/internal/storage"
)

type Response struct {
	resp.Response
	Url string `json:"url"`
}

//go:generate go run github.com/vektra/mockery/v2@latest --name=UrlGetter
type UrlGetter interface {
	GetUrl(alias string) (string, error)
}

func New(logger *slog.Logger, getter UrlGetter) http.HandlerFunc {
	const op = "handlers.url.redirect.New"
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		logger = logger.With(
			slog.String("op", op),
			slog.String("request_id", middleware.GetReqID(r.Context())),
		)
		alias := chi.URLParam(r, "alias")
		if alias == "" {
			logger.Warn("Query alias is empty")
			w.WriteHeader(http.StatusBadRequest)
			render.JSON(w, r, resp.Error("alias is required field"))
			return
		}
		url, err := getter.GetUrl(alias)
		if errors.Is(err, storage.ErrURLNotFound) {
			logger.Warn("Alias is not found", slog.String("alias", alias))
			w.WriteHeader(http.StatusNotFound)
			render.JSON(w, r, resp.Error("alias is not found"))
			return
		}
		if err != nil {
			logger.Error("Error to get alias", slog.String("alias", alias), sl.Err(err))
			w.WriteHeader(http.StatusInternalServerError)
			render.JSON(w, r, resp.Error("Internal Server Error"))
		}
		logger.Info("Url by alias found", slog.String("alias", alias), slog.String("url", url))
		http.Redirect(w, r, url, http.StatusFound)
	}
}
