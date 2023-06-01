package client

import (
	"fmt"
	"net/http"
)

type Error struct {
	Code    int
	Message string
}

func (e *Error) Error() string {
	return fmt.Sprintf("%s: %s", http.StatusText(e.Code), e.Message)
}
