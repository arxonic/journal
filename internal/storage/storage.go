package storage

import (
	"errors"
)

var (
	ErrUserNotFound   = errors.New("user not found")
	ErrCourseNotFound = errors.New("course not found")
)
