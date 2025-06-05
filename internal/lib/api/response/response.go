package response

import (
	"fmt"
	"strings"

    "github.com/go-playground/validator/v10"
)

type Response struct {
    Status string `json:"status"`
    Error string `json:"error,omitempty"`
}

const (
    StatusOK = "OK"
    StatusError = "ERROR"
)

func OK() Response {
    return Response{
        Status: StatusOK,
    }
}

func ERROR(msg string) Response {
    return Response{
        Status: StatusError,
        Error: msg,
    }
}

func ValidatationError(errs validator.ValidationErrors) Response {
    var errMsgs []string

    for _, err := range errs {
        switch err.ActualTag() {
        case "required":
            errMsgs = append(errMsgs, fmt.Sprintf("field %s is required", err.Field()))
        case "url":
            errMsgs = append(errMsgs, fmt.Sprintf("field %s is not valid URL", err.Field()))
        default:
            errMsgs = append(errMsgs, fmt.Sprintf("field %s is invalid", err.Field()))
        }
    }

    return Response{
        Status: StatusError,
        Error: strings.Join(errMsgs, ", "),
    }
}
