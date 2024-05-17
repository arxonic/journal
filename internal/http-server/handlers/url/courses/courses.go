package courses

import (
	"log/slog"
	"net/http"

	"github.com/arxonic/journal/internal/domain/models"
	"github.com/arxonic/journal/internal/domain/scheme"
	"github.com/arxonic/journal/internal/http-server/middleware/auth"
	resp "github.com/arxonic/journal/internal/lib/api/response"
	"github.com/arxonic/journal/internal/lib/logger/sl"
	"github.com/arxonic/journal/internal/services/policy"
	"github.com/arxonic/journal/internal/storage/sqlite"
	"github.com/go-chi/chi/middleware"
	"github.com/go-chi/render"
)

type CreateCourseResponse struct {
	CourseID int64 `json:"course_id"`
	resp.Responce
}

func Create(url string, log *slog.Logger, s *sqlite.Storage, ac *policy.AccessControl) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const fn = "http-server.handlers.url.cources.Create"

		log = log.With(
			slog.String("fn", fn),
			slog.String("request_id", middleware.GetReqID(r.Context())),
		)

		// Role check
		userAuthData := r.Context().Value(auth.ContextAuthMiddlewareKey).(*models.Key)
		if !ac.Contains(url, userAuthData.Role) {
			log.Error("unauthorized operation", sl.Err(policy.ErrUnauthorized))
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		log = log.With(
			slog.Int("user_id", userAuthData.ID),
		)

		var req scheme.Course

		err := render.DecodeJSON(r.Body, &req)
		if err != nil {
			log.Error("failed to decode request body", sl.Err(err))
			render.JSON(w, r, resp.Error("failed to decode request"))
			return
		}

		id, err := s.SaveCourse(&req)
		if err != nil {
			log.Error("failed to save course", sl.Err(err))
			render.JSON(w, r, resp.Error("ffailed to save course"))
			return
		}

		// Response
		render.JSON(w, r, CreateCourseResponse{
			Responce: resp.OK(),
			CourseID: id,
		})

		log.Info("course created")
	}
}
