package sqlite

import (
	"database/sql"
	"errors"
	"fmt"

	"github.com/arxonic/journal/internal/domain/models"
	"github.com/arxonic/journal/internal/domain/scheme"
	store "github.com/arxonic/journal/internal/storage"
	_ "github.com/mattn/go-sqlite3"
)

type Storage struct {
	db *sql.DB
}

func New(storagePath string) (*Storage, error) {
	const op = "storage.sqlite.New"

	db, err := sql.Open("sqlite3", storagePath)
	if err != nil {
		return nil, fmt.Errorf("%s:%w", op, err)
	}

	return &Storage{db: db}, nil
}

func (s *Storage) UserRole(email string) (models.Key, error) {
	const fn = "storage.sqlite.UserRole"

	stmt, err := s.db.Prepare("SELECT id, email, role FROM users WHERE email = ?")
	if err != nil {
		return models.Key{}, err
	}

	var key models.Key
	err = stmt.QueryRow(email).Scan(&key.ID, &key.Email, &key.Role)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return models.Key{}, store.ErrUserNotFound
		}
		return models.Key{}, fmt.Errorf("%s:%w", fn, err)
	}

	return key, nil
}

func (s *Storage) SaveCourse(course *scheme.Course) (int64, error) {
	const fn = "storage.sqlite.SaveCourse"

	stmt, err := s.db.Prepare("INSERT INTO courses (num, name) VALUES (?, ?)")
	if err != nil {
		return 0, fmt.Errorf("%s:%w", fn, err)
	}
	defer stmt.Close()

	res, err := stmt.Exec(course.Number, course.Name)
	if err != nil {
		return 0, fmt.Errorf("%s:%w", fn, err)
	}

	courseID, err := res.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("%s:%w", fn, err)
	}

	stmt, err = s.db.Prepare("INSERT INTO assignments (course_id, discipline_id, teacher_id) VALUES (?, ?, ?)")
	if err != nil {
		return 0, fmt.Errorf("%s:%w", fn, err)
	}

	for _, subject := range course.Subjects {
		_, err = stmt.Exec(courseID, subject.DisciplineID, subject.TeacherID)
		if err != nil {
			return 0, fmt.Errorf("%s:%w", fn, err)
		}
	}

	return courseID, nil
}

func (s *Storage) EnrollStudents(enrollments *scheme.Enrollments) error {
	const fn = "storage.sqlite.EnrollStudents"

	stmt, err := s.db.Prepare("INSERT INTO enrollments (course_id, student_id) VALUES (?, ?)")
	if err != nil {
		return fmt.Errorf("%s:%w", fn, err)
	}
	defer stmt.Close()

	for _, enroll := range enrollments.Enrollments {
		_, err = stmt.Exec(enroll.CourseID, enroll.StudentID)
		if err != nil {
			return fmt.Errorf("%s:%w", fn, err)
		}
	}

	return nil
}

func (s *Storage) RemoveStudents(enrollments *scheme.Enrollments) error {
	const fn = "storage.sqlite.EnrollStudents"

	stmt, err := s.db.Prepare("DELETE FROM enrollments WHERE course_id = ? AND student_id = ?")
	if err != nil {
		return fmt.Errorf("%s:%w", fn, err)
	}
	defer stmt.Close()

	for _, enroll := range enrollments.Enrollments {
		_, err = stmt.Exec(enroll.CourseID, enroll.StudentID)
		if err != nil {
			return fmt.Errorf("%s:%w", fn, err)
		}
	}

	return nil
}
