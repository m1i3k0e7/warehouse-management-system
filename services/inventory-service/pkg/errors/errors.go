package errors

import "fmt"

type ValidationError struct {
    Message string
    Cause   error
}

func (e *ValidationError) Error() string {
    if e.Cause != nil {
        return fmt.Sprintf("%s: %v", e.Message, e.Cause)
    }
    return e.Message
}

func NewValidationError(message string, cause error) *ValidationError {
    return &ValidationError{Message: message, Cause: cause}
}

type NotFoundError struct {
    Message string
    Cause   error
}

func (e *NotFoundError) Error() string {
    if e.Cause != nil {
        return fmt.Sprintf("%s: %v", e.Message, e.Cause)
    }
    return e.Message
}

func NewNotFoundError(message string, cause error) *NotFoundError {
    return &NotFoundError{Message: message, Cause: cause}
}

type ConflictError struct {
    Message string
    Cause   error
}

func (e *ConflictError) Error() string {
    if e.Cause != nil {
        return fmt.Sprintf("%s: %v", e.Message, e.Cause)
    }
    return e.Message
}

func NewConflictError(message string, cause error) *ConflictError {
    return &ConflictError{Message: message, Cause: cause}
}

type InternalError struct {
    Message string
    Cause   error
}

func (e *InternalError) Error() string {
    if e.Cause != nil {
        return fmt.Sprintf("%s: %v", e.Message, e.Cause)
    }
    return e.Message
}

func NewInternalError(message string, cause error) *InternalError {
    return &InternalError{Message: message, Cause: cause}
}