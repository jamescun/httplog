package httplog

import (
	"encoding/json"
	"io"
	"text/template"
)

func JSONLogger(dst io.Writer, reqs <-chan *Request) {
	for req := range reqs {
		json.NewEncoder(dst).Encode(req)
	}
}

var textBody = template.Must(template.New("").Parse(`
{{ .At.Format "15:04:05.000" }}:
Method: {{ .Method }}  Path: {{ .Path }}  Host: {{ .Host }}  Proto: {{ .Proto }}
{{- if .Query }}
Query:
{{- range $header, $values := .Query }}
  {{ $header }}: {{ range $i, $value := $values }}{{ if $i }},{{ end }}{{ $value }}{{ end -}}
{{ end }}{{ end }}
{{- if .Headers }}
Headers:
{{- range $header, $values := .Headers }}
  {{ $header }}: {{ range $i, $value := $values }}{{ if $i }},{{ end }}{{ $value }}{{ end -}}
{{ end }}{{ end }}
{{- if .Body }}
Body:
  {{ .Body }}{{ end }}
`))

func TextLogger(dst io.Writer, reqs <-chan *Request) {
	for req := range reqs {
		textBody.Execute(dst, req)
	}
}
