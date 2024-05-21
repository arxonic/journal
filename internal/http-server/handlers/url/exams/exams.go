package exams

import (
	"log/slog"
	"net/http"
	"time"

	"github.com/arxonic/journal/internal/domain/models"
	"github.com/arxonic/journal/internal/domain/scheme"
	"github.com/arxonic/journal/internal/http-server/middleware/auth"
	resp "github.com/arxonic/journal/internal/lib/api/response"
	"github.com/arxonic/journal/internal/lib/logger/sl"
	"github.com/arxonic/journal/internal/services/policy"
	"github.com/go-chi/render"
)

type ExamSignUper interface {
	ExamSignUp(int64, int64, time.Time) error
	AssignmentID(int64, int64, int64) (int64, error)
}

type ExamSignUpResponse struct {
	resp.Responce
}

type ExamSignUpRequest struct {
	scheme.Assignment
	ExamDate time.Time `json:"exam_date"`
}

func ExamSignUp(url string, log *slog.Logger, s ExamSignUper, ac *policy.AccessControl) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const fn = "http-server.handlers.url.exams.ExamSignUp"

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

		var req ExamSignUpRequest

		err := render.DecodeJSON(r.Body, &req)
		if err != nil {
			log.Error("failed to decode request body", sl.Err(err))
			render.JSON(w, r, resp.Error("failed to decode request"))
			return
		}

		assignmentID, err := s.AssignmentID(req.CourseID, req.DisciplineID, req.TeacherID)
		if err != nil {
			log.Error("failed to decode get assignmentID", sl.Err(err))
			render.JSON(w, r, resp.Error("failed to decode request"))
			return
		}

		// Exam sign up
		err = s.ExamSignUp(userAuthData.ID, assignmentID, req.ExamDate)
		if err != nil {
			log.Error("failed to sign up for the exam", sl.Err(err))
			render.JSON(w, r, resp.Error("failed to sign up for the exam"))
			return
		}

		// Response
		render.JSON(w, r, ExamSignUpResponse{
			Responce: resp.OK(),
		})
	}
}

type ExamGrader interface {
	AssignmentID(int64, int64, int64) (int64, error)
	ExamID(int64, int64, time.Time) (int64, error)
	ExamGrade(int64, int64, int, time.Time) error
}

type ExamGradeResponse struct {
	resp.Responce
}

type ExamGradeRequest struct {
	scheme.Assignment
	StudentID int64     `json:"student_id"`
	Grade     int       `json:"grade"`
	ExamDate  time.Time `json:"grade_date"`
}

func ExamGrade(url string, log *slog.Logger, s ExamGrader, ac *policy.AccessControl) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const fn = "http-server.handlers.url.exams.ExamGrade"

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

		var req ExamGradeRequest

		err := render.DecodeJSON(r.Body, &req)
		if err != nil {
			log.Error("failed to decode request body", sl.Err(err))
			render.JSON(w, r, resp.Error("failed to decode request"))
			return
		}

		// Get AssignmentID
		assignmentID, err := s.AssignmentID(req.CourseID, req.DisciplineID, userAuthData.ID)
		if err != nil {
			log.Error("failed to decode get assignmentID", sl.Err(err))
			render.JSON(w, r, resp.Error("failed to decode request"))
			return
		}

		// Get ExamID
		examID, err := s.ExamID(req.StudentID, assignmentID, req.ExamDate)
		if err != nil {
			log.Error("failed to get exam", sl.Err(err))
			render.JSON(w, r, resp.Error("failed to get exam"))
			return
		}

		// Grading
		err = s.ExamGrade(examID, userAuthData.ID, req.Grade, req.ExamDate)
		if err != nil {
			log.Error("failed to set exam grade", sl.Err(err))
			render.JSON(w, r, resp.Error("error in rating"))
			return
		}

		// Response
		render.JSON(w, r, ExamSignUpResponse{
			Responce: resp.OK(),
		})
	}
}

type Grader interface {
	ExamByStudentID(int64) (scheme.Exam, error)
	GradesByExamID(int64) (scheme.Grade, error)
	Assignment(int64) (scheme.Assignment, error)
	ExamGrade(int64, int64, int, time.Time) error
}

type GradesResponse struct {
	resp.Responce
}

type GradesRequest struct {
	scheme.Assignment
	StudentID int64     `json:"student_id"`
	Grade     int       `json:"grade"`
	ExamDate  time.Time `json:"grade_date"`
}
