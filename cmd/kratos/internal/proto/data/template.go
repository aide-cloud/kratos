package data

import (
	"bytes"
	"html/template"
)

//nolint:lll
var serviceTemplate = `
{{- /* delete empty line */ -}}
package data

import (
	{{- if .UseContext }}
	"context"
	{{- end }}
	"github.com/go-kratos/kratos/v2/log"
)

type (
	{{ .Service }}Repo struct {
        logger *log.Helper
		data   *Data
	}
)

func New{{ .Service }}Repo(data *Data, logger log.Logger) *{{ .Service }}Repo {
	return &{{ .Service }}Repo{data: data, logger: log.NewHelper(log.With(logger, "module", "data/{{ .Service }}"))}
}
`

type MethodType uint8

const (
	unaryType          MethodType = 1
	twoWayStreamsType  MethodType = 2
	requestStreamsType MethodType = 3
	returnsStreamsType MethodType = 4
)

// Service is a proto service.
type Service struct {
	Package     string
	Service     string
	Methods     []*Method
	GoogleEmpty bool

	UseIO      bool
	UseContext bool
}

// Method is a proto method.
type Method struct {
	Service string
	Name    string
	Request string
	Reply   string

	// type: unary or stream
	Type MethodType
}

func (s *Service) execute() ([]byte, error) {
	const empty = "google.protobuf.Empty"
	buf := new(bytes.Buffer)
	for _, method := range s.Methods {
		if (method.Type == unaryType && (method.Request == empty || method.Reply == empty)) ||
			(method.Type == returnsStreamsType && method.Request == empty) {
			s.GoogleEmpty = true
		}
		if method.Type == twoWayStreamsType || method.Type == requestStreamsType {
			s.UseIO = true
		}
		if method.Type == unaryType {
			s.UseContext = true
		}
	}
	tmpl, err := template.New("service").Parse(serviceTemplate)
	if err != nil {
		return nil, err
	}
	if err := tmpl.Execute(buf, s); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}
