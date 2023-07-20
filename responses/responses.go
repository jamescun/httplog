package responses

import (
	"io"
	"net/http"
	"os"

	"github.com/go-chi/chi/v5"
	"gopkg.in/yaml.v3"
)

type File struct {
	// Responses are all the pre-defined responses that HTTPLog will be
	// configured to reply to. If none match the request, Default will be used.
	Responses []*Response `yaml:"responses"`

	// Headers are HTTP Headers that are to be sent in response to all
	// requests.
	Headers http.Header `yaml:"headers"`

	// NotFound is the default response sent by HTTPLog when no Response was
	// matched for a request.
	NotFound *Response `yaml:"notFound"`

	// MethodNotAllowed is the default response sent by HTTPLog when a Response
	// was matched but not for the method of the request.
	MethodNotAllowed *Response `yaml:"methodNotAllowed"`
}

// ReadFile reads and unmarshals a Responsefile from a YAML file on disk.
func ReadFile(path string) (*File, error) {
	file, err := os.OpenFile(path, os.O_RDONLY, 0)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var cfg File

	err = yaml.NewDecoder(file).Decode(&cfg)
	if err != nil {
		return nil, err
	}

	return &cfg, nil
}

// Handler builds the response routes configured on File into a HTTP router.
func (f *File) Handler() http.Handler {
	r := chi.NewRouter()

	if len(f.Headers) > 0 {
		r.Use(func(next http.Handler) http.Handler {
			return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				for key, values := range f.Headers {
					for _, value := range values {
						w.Header().Set(key, value)
					}
				}

				next.ServeHTTP(w, r)
			})
		})
	}

	if f.NotFound != nil {
		r.NotFound(f.NotFound.handlerFunc())
	}

	if f.MethodNotAllowed != nil {
		r.MethodNotAllowed(f.MethodNotAllowed.handlerFunc())
	}

	for _, response := range f.Responses {
		if response.Method != "" {
			r.Method(response.Method, response.Path, response.handlerFunc())
		} else {
			r.Handle(response.Path, response.handlerFunc())
		}
	}

	return r
}

type Response struct {
	// Method is the HTTP Method verb to match this request, or all requests if
	// not configured.
	Method string `yaml:"method"`

	// Path is the URL Path where this response/ will be made available by
	// HTTPLog. Internally, HTTPLog uses the Chi router, which supports
	// parameters and regular expressions. Parameters are currently unused.
	Path string `yaml:"path"`

	// Status is the HTTP Status Code returned by this Response, or HTTP 200
	// if not configured.
	Status int `yaml:"status"`

	// Headers are HTTP Headers that are sent in response to this request only.
	Headers http.Header `yaml:"headers"`

	// Body is the contents of the response body.
	Body string `yaml:"body"`

	// File is like Body, except reads from a file path.
	File string `yaml:"file"`
}

func (res *Response) handlerFunc() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		for key, values := range res.Headers {
			for _, value := range values {
				w.Header().Set(key, value)
			}
		}

		if res.Status > 0 {
			w.WriteHeader(res.Status)
		} else {
			w.WriteHeader(http.StatusOK)
		}

		if res.Body != "" {
			io.WriteString(w, res.Body)
		} else if res.File != "" {
			http.ServeFile(w, r, res.File)
		}
	}
}
