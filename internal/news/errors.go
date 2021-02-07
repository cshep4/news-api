package news

import (
	"errors"
	"fmt"
)

var (
	ErrProviderNotFound = errors.New("provider not found")
	ErrCategoryNotFound = errors.New("category not found")
)

// InvalidParameterError is returned when a parameter is invalid.
type InvalidParameterError struct {
	Parameter string
}

func (i InvalidParameterError) Error() string {
	return fmt.Sprintf("invalid parameter: %s", i.Parameter)
}
