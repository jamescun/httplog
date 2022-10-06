package httplog

import (
	"encoding/base64"
	"io"
	"mime"
	"net/http"
	"time"
)

type Server struct {
	// ResponseCode is the HTTP Status Code sent in response to all requests,
	// if not set, HTTP 200 is used.
	ResponseCode int

	// ResponseBody is the contents sent in response to all requests, if not
	// set, no response body is used.
	ResponseBody []byte

	requests chan *Request
}

func NewServer(requestBufferSize int) *Server {
	return &Server{
		requests: make(chan *Request, requestBufferSize),
	}
}

func (s *Server) Requests() <-chan *Request {
	return s.requests
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	req := &Request{
		Method:  r.Method,
		Path:    r.URL.Path,
		Proto:   r.Proto,
		Headers: r.Header,
		Host:    r.Host,
		At:      time.Now(),
	}

	q := r.URL.Query()
	if len(q) > 0 {
		req.Query = q
	}

	body, _ := io.ReadAll(r.Body)
	if len(body) > 0 {
		contentType := r.Header.Get("Content-Type")
		mediaType, _, _ := mime.ParseMediaType(contentType)

		switch mediaType {
		case "application/json", "application/x-www-form-urlencoded":
			req.Body = string(body)

		default:
			req.Body = base64.StdEncoding.EncodeToString(body)
		}
	}

	s.requests <- req

	if s.ResponseCode > 0 {
		w.WriteHeader(s.ResponseCode)
	}

	if len(s.ResponseBody) > 0 {
		w.Write(s.ResponseBody)
	}
}
