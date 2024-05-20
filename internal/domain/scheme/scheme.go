package scheme

// User
type User struct {
	ID         int64  `json:"user_id"`
	LastName   string `json:"last_name"`
	FirstName  string `json:"first_name"`
	Patronymic string `json:"patronymic"`
}

// Test
type CourseCreation struct {
	Name     string    `json:"name"`
	Number   int       `json:"number"`
	Subjects []Subject `json:"subjects,omitempty"`
}

type Subject struct {
	TeacherID    int64 `json:"teacher_id"`
	DisciplineID int64 `json:"discipline_id"`
}

// Enrollments
type Enrollments struct {
	Enrollments []Enrollment `json:"enrolls,omitempty"`
}

type Enrollment struct {
	ID        int64 `json:"enroll_id"`
	CourseID  int64 `json:"course_id"`
	StudentID int64 `json:"student_id"`
}

// Courses
type Courses struct {
	Courses []Course `json:"courses"`
}

type Course struct {
	ID     int64  `json:"course_id"`
	Name   string `json:"course_name"`
	Number int    `json:"course_number"`
	Disciplines
}

// Disciplines
type Disciplines struct {
	Disciplines []Discipline `json:"disciplines"`
}

type Discipline struct {
	ID       int64  `json:"discipline_id"`
	Name     string `json:"discipline_name"`
	Teachers []User `json:"teachers,omitempty"`
}

// Assignments
type Assignments struct {
	Assignments []Assignment `json:"assignments"`
}

type Assignment struct {
	ID           int64 `json:"assignment_id"`
	CourseID     int64 `json:"course_id"`
	DisciplineID int64 `json:"discipline_id"`
	TeacherID    int64 `json:"teacher_id"`
}
