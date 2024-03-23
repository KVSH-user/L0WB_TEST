package order

import (
	"L0WB/internal/cache"
	resp "L0WB/internal/lib/api/response"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/render"
	"log/slog"
	"net/http"
)

func GetOrder(log *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "handlers.order.GetOrder"

		log := log.With(
			slog.String("op", op),
			slog.String("request_id", middleware.GetReqID(r.Context())),
		)

		id := chi.URLParam(r, "id")
		if id == "" {
			log.Info("id is empty")
			w.WriteHeader(http.StatusBadRequest)
			render.JSON(w, r, resp.Error("id parameter is required"))
			return
		}

		response, err := cache.GetOrder(id)
		if err != nil {
			w.WriteHeader(http.StatusNotFound)
			render.JSON(w, r, resp.Error("id not found"))

			return
		}

		log.Info("order geted ")

		w.WriteHeader(http.StatusOK)
		render.JSON(w, r, response)
	}
}
