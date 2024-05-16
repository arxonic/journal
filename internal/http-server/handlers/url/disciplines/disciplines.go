package disciplines

import (
	"fmt"
	"log/slog"
	"net/http"

	"github.com/arxonic/journal/internal/domain/models"
	"github.com/arxonic/journal/internal/http-server/middleware/auth"
	"github.com/arxonic/journal/internal/storage/sqlite"
	"github.com/go-chi/chi/middleware"
)

func New(log *slog.Logger, storage *sqlite.Storage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "http-server.handlers.url.disciplines.New"

		log = log.With(
			slog.String("op", op),
			slog.String("request_id", middleware.GetReqID(r.Context())),
		)

		user := r.Context().Value(auth.ContextAuthMiddlewareKey)
		fmt.Println("A")
		fmt.Println(user.(*models.Key))
		// var req Request

		// err := render.DecodeJSON(r.Body, &req)
		// if err != nil {
		// 	log.Error("failed to decode request body", sl.Err(err))
		// 	render.JSON(w, r, resp.Error("failed to decode request"))
		// 	return
		// }

		// log.Info("request body decoded", slog.Any("request", req))

		// if err := validator.New().Struct(req); err != nil {
		// 	validateErr := err.(validator.ValidationErrors)
		// 	log.Error("invalid request", sl.Err(err))
		// 	render.JSON(w, r, resp.ValidationError(validateErr))
		// 	return
		// }

		// // Response
		// render.JSON(w, r, Responce{
		// 	Responce: resp.OK(),
		// 	Products: resp.Products{},
		// })
		// log.Info("responce was sent to the client")
	}
}
