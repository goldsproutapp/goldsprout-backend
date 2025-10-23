package exceptions

import (
	"errors"
	"fmt"
)

var UserForbiddenBase = errors.New("Forbidden")
var InvalidRequestBase = errors.New("Invalid request")
var ConflictBase = errors.New("Conflict")

func UserForbidden(message string) error {
	return fmt.Errorf("%v %w", message, UserForbiddenBase)
}

func InvalidRequest(message string) error {
	return fmt.Errorf("%v %w", message, InvalidRequestBase)
}

func Conflict(message string) error {
	return fmt.Errorf("%v %w", message, ConflictBase)
}
