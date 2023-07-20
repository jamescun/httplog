package httplog

import (
	"encoding/base64"
	"io"
	"mime"
	"net/http"
	"strings"
	"time"
)

type Server struct {
	// Handler is the HTTP Handler executed for all requests. If not
	// configured, an HTTP 200 will be sent in response to all requests.
	Handler http.Handler

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

	if s.Handler != nil {
		s.Handler.ServeHTTP(w, r)
	} else {
		w.WriteHeader(http.StatusOK)
	}
}

// ParseHeaders unpacks a slice of KEY=VALUE formatted strings into a set of
// HTTP Headers. If a value cannot be split into KEY=VALUE, it is ignored.
func ParseHeaders(headers []string) http.Header {
	if len(headers) < 1 {
		return nil
	}

	res := http.Header{}

	for _, header := range headers {
		key, value := splitKeyValue(header, '=')
		if value != "" {
			res.Add(key, value)
		}
	}

	return res
}

func splitKeyValue(s string, sep byte) (key, value string) {
	i := strings.IndexByte(s, sep)
	if i > -1 {
		key = s[:i]
		value = s[i+1:]
	} else {
		key = s
	}

	return
}
