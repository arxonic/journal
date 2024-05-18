package courses

import (
	"log/slog"
	"net/http"
	"strconv"

	"github.com/arxonic/journal/internal/domain/models"
	"github.com/arxonic/journal/internal/domain/scheme"
	"github.com/arxonic/journal/internal/http-server/middleware/auth"
	resp "github.com/arxonic/journal/internal/lib/api/response"
	"github.com/arxonic/journal/internal/lib/logger/sl"
	"github.com/arxonic/journal/internal/services/policy"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/render"
)

type CourseSaver interface {
	SaveCourse(*scheme.Course) (int64, error)
}

type CreateCourseResponse struct {
	CourseID int64 `json:"course_id"`
	resp.Responce
}

func Create(url string, log *slog.Logger, s CourseSaver, ac *policy.AccessControl) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const fn = "http-server.handlers.url.cources.Create"

		log = log.With(
			slog.String("fn", fn),
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
			render.JSON(w, r, resp.Error("failed to save course"))
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

type StudentsEnroller interface {
	EnrollStudents(*scheme.Enrollments) error
}

type EnrollStudentsResponse struct {
	resp.Responce
}

func EnrollStudents(url string, log *slog.Logger, s StudentsEnroller, ac *policy.AccessControl) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const fn = "http-server.handlers.url.cources.EnrollStudents"

		log = log.With(
			slog.String("fn", fn),
		)

		// Role check
		userAuthData := r.Context().Value(auth.ContextAuthMiddlewareKey).(*models.Key)
		if !ac.Contains(url, userAuthData.Role) {
			log.Error("unauthorized operation", sl.Err(policy.ErrUnauthorized))
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		// Get courseID from URL
		courseIDString := chi.URLParam(r, "courseID")

		courseID, err := strconv.Atoi(courseIDString)
		if err != nil {
			log.Info("unknown course_id")
			render.JSON(w, r, resp.Error("course not found"))
			return
		}

		log = log.With(
			slog.Int("user_id", userAuthData.ID),
			slog.Int("course_id", courseID),
		)

		var req scheme.Enrollments

		err = render.DecodeJSON(r.Body, &req)
		if err != nil {
			log.Error("failed to decode request body", sl.Err(err))
			render.JSON(w, r, resp.Error("failed to decode request"))
			return
		}

		err = s.EnrollStudents(&req)
		if err != nil {
			log.Error("failed to enroll students", sl.Err(err))
			render.JSON(w, r, resp.Error("failed to enroll students"))
			return
		}

		// Response
		render.JSON(w, r, CreateCourseResponse{
			Responce: resp.OK(),
		})

		log.Info("students enrolled")
	}
}
