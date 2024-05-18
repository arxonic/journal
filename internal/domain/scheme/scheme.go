package scheme

type Course struct {
	Name     string    `json:"name"`
	Number   int       `json:"number"`
	Subjects []Subject `json:"subjects"`
}

type Subject struct {
	TeacherID    int `json:"teacher_id"`
	DisciplineID int `json:"discipline_id"`
}

type Enrollments struct {
	Enrollments []Enrollment `json:"enrollments"`
}

type Enrollment struct {
	CourseID  int `json:"course_id"`
	StudentID int `json:"student_id"`
}
