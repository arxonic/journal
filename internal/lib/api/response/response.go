package response

import (
	"fmt"
	"strings"

	"github.com/go-playground/validator/v10"
)

type Responce struct {
	Status string `json:"status"`
	Error  string `json:"error,omitempty"`
}

const (
	StatusOK    = "OK"
	StatusError = "Error"
)

func OK() Responce {
	return Responce{
		Status: StatusOK,
	}
}

func Error(msg string) Responce {
	return Responce{
		Status: StatusError,
		Error:  msg,
	}
}

func ValidationError(errs validator.ValidationErrors) Responce {
	var errMsgs []string

	for _, err := range errs {
		switch err.ActualTag() {
		case "required":
			errMsgs = append(errMsgs, fmt.Sprintf("field %s is a required field", err.Field()))
		case "url":
			errMsgs = append(errMsgs, fmt.Sprintf("field %s is not a valid URL", err.Field()))
		default:
			errMsgs = append(errMsgs, fmt.Sprintf("field %s is not valid", err.Field()))
		}
	}

	return Responce{
		Status: StatusError,
		Error:  strings.Join(errMsgs, ", "),
	}
}
