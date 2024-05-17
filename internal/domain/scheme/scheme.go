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
