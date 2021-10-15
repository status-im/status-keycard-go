package statuskeycardgo

import "fmt"

type keycardError struct {
	message string
}

func (e *keycardError) Error() string {
	return fmt.Sprintf("keycard-error: %s", e.message)
}

func newKeycardError(message string) *keycardError {
	return &keycardError{message}
}
