package response

import (
	"fmt"
	"strings"

	"github.com/go-playground/validator/v10"
)

type Product struct {
	Type    string `json:"type"`
	Name    string `json:"name"`
	Id      string `json:"id"`
	Referer string `json:"referer"`
	ImgSrc  string `json:"img"`
	URL     string `json:"url"`
}

type Products struct {
	Products []Product `json:"products"`
}

type Responce struct {
	Status string `json:"status"`
	Error  string `json:"error,omitempty"`
	//Products Products
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

// 	log.Error("fa
