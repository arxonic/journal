-- Таблица Users
CREATE TABLE IF NOT EXISTS users
(
    id          INTEGER PRIMARY KEY AUTOINCREMENT,
    email       VARCHAR(100) NOT NULL UNIQUE,
    role        TEXT CHECK(role IN ('student', 'teacher', 'admin', 'unknown')) NOT NULL,
    last_name   VARCHAR(50) NOT NULL,
    first_name  VARCHAR(50) NOT NULL,
    patronymic  VARCHAR(50),
    phone       VARCHAR(50) NOT NULL
);
CREATE INDEX IF NOT EXISTS idx_email ON users (email);

-- Таблица Students
CREATE TABLE IF NOT EXISTS students(
    id              INTEGER PRIMARY KEY AUTOINCREMENT,
    city            VARCHAR(50),
    record_book_id  VARCHAR(50),
    user_id         INTEGER NOT NULL,
    FOREIGN KEY (user_id) REFERENCES users(id)
);

-- Таблица Courses
CREATE TABLE IF NOT EXISTS courses(
    id      INTEGER PRIMARY KEY AUTOINCREMENT,
    num     INTEGER NOT NULL,
    name    VARCHAR(100) NOT NULL
);

-- Таблица Disciplines
CREATE TABLE IF NOT EXISTS disciplines(
    id      INTEGER PRIMARY KEY AUTOINCREMENT,
    name    VARCHAR(100) NOT NULL
);

-- Таблица Assignments
CREATE TABLE IF NOT EXISTS assignments(
    id              INTEGER PRIMARY KEY AUTOINCREMENT,
    course_id       INTEGER NOT NULL,
    discipline_id   INTEGER NOT NULL,
    teacher_id      INTEGER NOT NULL,
    FOREIGN KEY (teacher_id) REFERENCES users(id),
    FOREIGN KEY (discipline_id) REFERENCES disciplines(id),
    FOREIGN KEY (course_id) REFERENCES courses(id)
);

-- Таблица Enrollments
CREATE TABLE IF NOT EXISTS enrollments(
    id              INTEGER PRIMARY KEY AUTOINCREMENT,
    course_id       INTEGER NOT NULL,
    student_id      INTEGER NOT NULL,
    FOREIGN KEY (course_id) REFERENCES courses(id),
    FOREIGN KEY (student_id) REFERENCES users(id)
);

-- Таблица Grades
CREATE TABLE IF NOT EXISTS grades(
    id          INTEGER PRIMARY KEY AUTOINCREMENT,
    exam_id     INTEGER NOT NULL,
    teacher_id  INTEGER NOT NULL,
    grade       INTEGER NOT NULL CHECK(grade BETWEEN 1 AND 5),
    grade_date  DATE,
    FOREIGN KEY (exam_id) REFERENCES exams(id),
    FOREIGN KEY (teacher_id) REFERENCES users(id)
);

-- Таблица Exams
CREATE TABLE IF NOT EXISTS exams(
    id              INTEGER PRIMARY KEY AUTOINCREMENT,
    student_id      INTEGER NOT NULL,
    assignment_id   INTEGER NOT NULL,
    exam_date       DATE,
    FOREIGN KEY (student_id) REFERENCES users(id),
    FOREIGN KEY (assignment_id) REFERENCES assignments(id)
);