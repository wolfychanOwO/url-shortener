package delete

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/render"
	"log/slog"

	resp "main/internal/lib/api/response"
	"main/internal/lib/logger/sl"
	"main/internal/storage"
)

type URLDeleter interface {
	DeleteURL(alias string) error
}

func New(log *slog.Logger, deleter URLDeleter) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "handlers.url.delete.New"

		log := log.With(
			slog.String("op", op),
			slog.String("request_id", middleware.GetReqID(r.Context())),
		)

		alias := chi.URLParam(r, "alias")
		if alias == "" {
			log.Error("alias param is empty")
			render.JSON(w, r, resp.ERROR("alias param is required"))
			return 
		}

		if err := deleter.DeleteURL(alias); err != nil {
			if err == storage.ErrURLNotFound {
				log.Info("alias not found", slog.String("alias", alias))
				render.JSON(w, r, resp.ERROR("alias not found"))
				return
			}

			log.Error("failed to delete url", sl.Err(err))
			render.JSON(w, r, resp.ERROR("failed to delete url"))
			return 
		}

		log.Info("url deleted", slog.String("alias", alias))

		render.JSON(w, r, resp.OK())
	}
}

