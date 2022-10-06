package httplog

import (
	"time"
)

type Request struct {
	Method  string              `json:"method"`
	Path    string              `json:"path"`
	Query   map[string][]string `json:"query,omitempty"`
	Proto   string              `json:"proto"`
	Host    string              `json:"host"`
	Headers map[string][]string `json:"headers,omitempty"`
	Body    string              `json:"body,omitempty"`
	At      time.Time           `json:"at"`
}
