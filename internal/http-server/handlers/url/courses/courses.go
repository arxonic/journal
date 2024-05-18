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

type CoursesGetter interface {
	TeacherCourses(int64) (scheme.Courses, error)
	StudentCourses(int64) (scheme.Courses, error)
}

type GetCoursesResponse struct {
	resp.Responce
	scheme.Courses
	scheme.Enrollments
}

func Get(url string, log *slog.Logger, s CoursesGetter, ac *policy.AccessControl) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const fn = "http-server.handlers.url.cources.Get"

		log = log.With(
			slog.String("fn", fn),
		)

		// User Role check
		userAuthData := r.Context().Value(auth.ContextAuthMiddlewareKey).(*models.Key)
		if !ac.Contains(url, userAuthData.Role) {
			log.Error("unauthorized operation", sl.Err(policy.ErrUnauthorized))
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		log = log.With(
			slog.Int64("user_id", userAuthData.ID),
		)

		// Get Courses
		courses, err := get(s, userAuthData.ID, userAuthData.Role)
		if err != nil {
			log.Error("failed to decode request body", sl.Err(err))
			render.JSON(w, r, resp.Error("failed to decode request"))
			return
		}

		// Response
		render.JSON(w, r, GetCoursesResponse{
			Responce: resp.OK(),
			Courses:  courses,
		})
	}
}

func get(s CoursesGetter, id int64, role string) (scheme.Courses, error) {
	var courses scheme.Courses
	var err error
	switch role {
	case "teacher":
		courses, err = s.TeacherCourses(id)
	case "student":
		courses, err = s.StudentCourses(id)
	case "admin":

	default:
		err = policy.ErrUnauthorized
	}
	return courses, err
}

type CourseSaver interface {
	SaveCourse(*scheme.CourseCreation) (int64, error)
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
			slog.Int64("user_id", userAuthData.ID),
		)

		var req scheme.CourseCreation

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

		_courseID, err := strconv.Atoi(courseIDString)
		if err != nil {
			log.Info("unknown courseID")
			render.JSON(w, r, resp.Error("course not found"))
			return
		}

		courseID := int64(_courseID)

		log = log.With(
			slog.Int64("user_id", userAuthData.ID),
			slog.Int64("course_id", courseID),
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
		render.JSON(w, r, EnrollStudentsResponse{
			Responce: resp.OK(),
		})

		log.Info("students enrolled")
	}
}

type StudentsRemover interface {
	RemoveStudents(*scheme.Enrollments) error
}

type RemoveStudentsResponse struct {
	resp.Responce
}

func RemoveStudents(url string, log *slog.Logger, s StudentsRemover, ac *policy.AccessControl) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const fn = "http-server.handlers.url.cources.RemoveStudents"

		log = log.With(
			slog.String("fn", fn),
		)

		// User role check
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
			log.Info("unknown courseID")
			render.JSON(w, r, resp.Error("course not found"))
			return
		}

		log = log.With(
			slog.Int64("user_id", userAuthData.ID),
			slog.Int("course_id", courseID),
		)

		var req scheme.Enrollments

		err = render.DecodeJSON(r.Body, &req)
		if err != nil {
			log.Error("failed to decode request body", sl.Err(err))
			render.JSON(w, r, resp.Error("failed to decode request"))
			return
		}

		err = s.RemoveStudents(&req)
		if err != nil {
			log.Error("failed to remove students", sl.Err(err))
			render.JSON(w, r, resp.Error("failed to remove students"))
			return
		}

		// Response
		render.JSON(w, r, RemoveStudentsResponse{
			Responce: resp.OK(),
		})

		log.Info("students removed")
	}
}
