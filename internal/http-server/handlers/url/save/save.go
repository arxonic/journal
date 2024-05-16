package save

import (
	"log/slog"
	"net/http"

	resp "github.com/arxonic/journal/internal/lib/api/response"
	"github.com/arxonic/journal/internal/lib/logger/sl"

	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/render"
	"github.com/go-playground/validator/v10"
)

type Request struct {
	URL string `json:"url" validate:"required,url"`
	//Products string `json:"products,omitempty"`
}

type Responce struct {
	resp.Responce
	Products string `json:"products,omitempty"`
}

type URLSaver interface {
	SaveURL(url string, products []string) error
}

func New(log *slog.Logger, urlSaver URLSaver) http.HandlerFunc {
	log.Info("save.new")
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "http-server.handlers.url.save.New"

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

		// TODO: GET products
		products := "aaaaaa"

		log.Info("url ...")

		render.JSON(w, r, Responce{
			Responce: resp.OK(),
			Products: products,
		})
	}
}
