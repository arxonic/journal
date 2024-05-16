package analysis

import (
	// "fmt"

	resp "github.com/arxonic/journal/internal/lib/api/response"
	"github.com/arxonic/journal/internal/lib/logger/sl"

	"log/slog"
	"net/http"

	"github.com/go-chi/chi/middleware"
	"github.com/go-chi/render"
	"github.com/go-playground/validator/v10"
)

type Request struct {
	URL string `json:"url" validate:"required,url"`
	//Products string `json:"products,omitempty"`
}

type R struct {
	Yes string `json:"yes"`
}

type Responce struct {
	resp.Responce
	resp.Products
	//Products Products `json:"products,omitempty"`
}

func New(log *slog.Logger, ytDlpPath, yandexSecretPath string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "http-server.handlers.url.analysis.New"

		log = log.With(
			slog.String("op", op),
			slog.String("request_id", middleware.GetReqID(r.Context())),
		)

		var req Request

		err := render.DecodeJSON(r.Body, &req)
		if err != nil {
			log.Error("failed to decode request body", sl.Err(err))
			render.JSON(w, r, resp.Error("failed to decode request"))
			return
		}

		log.Info("request body decoded", slog.Any("request", req))

		if err := validator.New().Struct(req); err != nil {
			validateErr := err.(validator.ValidationErrors)
			log.Error("invalid request", sl.Err(err))
			render.JSON(w, r, resp.ValidationError(validateErr))
			return
		}

		// Response
		render.JSON(w, r, Responce{
			Responce: resp.OK(),
			Products: resp.Products{},
		})
		log.Info("responce was sent to the client")
	}
}
