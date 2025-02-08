package restTemplate

import (
	"net/http"
)

type RawResponse struct {
	Status     int
	Headers    http.Header
	RawResBody []byte
}

type Status int

type Header http.Header
