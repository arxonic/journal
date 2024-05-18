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

func (s *Storage) SaveCourse(course *scheme.CourseCreation) (int64, error) {
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
	const fn = "storage.sqlite.RemoveStudents"

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

// FIXME Simplify methods

func (s *Storage) TeacherCourses(teacherID int64) (scheme.Courses, error) {
	const fn = "storage.sqlite.TeacherCourses"

	assignments, err := s.AssignmentsByFK(teacherID, "teacher_id")
	if err != nil {
		return scheme.Courses{}, fmt.Errorf("%s:%w", fn, err)
	}

	var courses scheme.Courses

	for _, ass := range assignments.Assignments {
		course, err := s.Course(ass.CourseID)
		if err != nil {
			return scheme.Courses{}, fmt.Errorf("%s:%w", fn, err)
		}

		courses.Courses = append(courses.Courses, course)
	}

	return courses, nil
}

func (s *Storage) StudentCourses(studentID int64) (scheme.Courses, error) {
	const fn = "storage.sqlite.StudentCourses"

	enrolls, err := s.EntollmentsByFK(studentID, "student_id")
	if err != nil {
		return scheme.Courses{}, fmt.Errorf("%s:%w", fn, err)
	}

	var courses scheme.Courses

	for _, enroll := range enrolls.Enrollments {
		course, err := s.Course(enroll.CourseID)
		if err != nil {
			return scheme.Courses{}, fmt.Errorf("%s:%w", fn, err)
		}

		courses.Courses = append(courses.Courses, course)
	}

	return courses, nil
}

// Get the Course by ID
func (s *Storage) Course(courseID int64) (scheme.Course, error) {
	const fn = "storage.sqlite.Course"

	stmt, err := s.db.Prepare("SELECT id, name, num FROM courses WHERE id = ?")
	if err != nil {
		return scheme.Course{}, fmt.Errorf("%s:%w", fn, err)
	}
	defer stmt.Close()

	var course scheme.Course

	err = stmt.QueryRow(courseID).Scan(&course.ID, &course.Name, &course.Number)
	if err != nil {
		return scheme.Course{}, fmt.Errorf("%s:%w", fn, err)
	}

	return course, nil
}

func (s *Storage) AssignmentsByFK(fk int64, fieldName string) (scheme.Assignments, error) {
	const fn = "storage.sqlite.AssignmentsByFK"

	req := fmt.Sprintf("SELECT id, course_id, discipline_id, teacher_id FROM assignments WHERE %s = ?", fieldName)
	stmt, err := s.db.Prepare(req)
	if err != nil {
		return scheme.Assignments{}, fmt.Errorf("%s:%w", fn, err)
	}
	defer stmt.Close()

	rows, err := stmt.Query(fk)
	if err != nil {
		return scheme.Assignments{}, fmt.Errorf("%s:%w", fn, err)
	}

	var assignments scheme.Assignments

	for rows.Next() {
		var ass scheme.Assignment
		if err := rows.Scan(&ass.ID, &ass.CourseID, &ass.DisciplineID, &ass.TeacherID); err != nil {
			return scheme.Assignments{}, fmt.Errorf("%s:%w", fn, err)
		}

		assignments.Assignments = append(assignments.Assignments, ass)
	}

	return assignments, nil
}

func (s *Storage) EntollmentsByFK(fk int64, fieldName string) (scheme.Enrollments, error) {
	const fn = "storage.sqlite.EntollmentsByFK"

	req := fmt.Sprintf("SELECT id, course_id, student_id FROM enrollments WHERE %s = ?", fieldName)
	stmt, err := s.db.Prepare(req)
	if err != nil {
		return scheme.Enrollments{}, fmt.Errorf("%s:%w", fn, err)
	}
	defer stmt.Close()

	rows, err := stmt.Query(fk)
	if err != nil {
		return scheme.Enrollments{}, fmt.Errorf("%s:%w", fn, err)
	}

	var enrolls scheme.Enrollments

	for rows.Next() {
		var enroll scheme.Enrollment
		if err := rows.Scan(&enroll.ID, &enroll.CourseID, &enroll.StudentID); err != nil {
			return scheme.Enrollments{}, fmt.Errorf("%s:%w", fn, err)
		}

		enrolls.Enrollments = append(enrolls.Enrollments, enroll)
	}

	return enrolls, nil
}
