package graphql

import (
	"bytes"
	"html/template"
)

//nolint:lll
var serviceTemplate = `
{{- /* delete empty line */ -}}
package graph

import (
	{{- if .UseContext }}
	"context"
	{{- end }}
	{{- if .UseIO }}
	"io"
	{{- end }}
	"github.com/go-kratos/kratos/v2/log"
	pb "{{ .Package }}"
	{{- if .GoogleEmpty }}
	"google.golang.org/protobuf/types/known/emptypb"
	{{- end }}
)

type (
	{{ .Service }}GraphqlService struct {
		*service.{{ .Service }}Service
		logger *log.Helper
	}

	{{ range .Methods }}
	{{ .Request }} struct {}
	{{ .Reply }} struct {}
	{{- end }}
)

func New{{ .Service }}GraphqlService(s *service.{{ .Service }}Service, logger log.Logger) *{{ .Service }}GraphqlService {
	return &{{ .Service }}GraphqlService{ {{ .Service }}Service: s, logger: log.NewHelper(log.With(logger, "module", "graph/{{ .Service }}"))}
}

{{- $s1 := "google.protobuf.Empty" }}
{{ range .Methods }}
func (s *{{ .Service }}GraphqlService) {{ .Name }}(ctx context.Context, args struct {
	In *{{ .Request }}
}) (*{{ .Reply }}, error) {
	// TODO use args.In params
	res, err := s.{{ .Service }}Service.{{ .Name }}(ctx, &{{ if eq .Request $s1 }}emptypb.Empty{}{{ else }}pb.{{ .Request }}{}{{ end }})
	if err != nil {
		s.logger.Errorf("{{ .Name }} err: +v%", err)
		return nil, err
	}

	s.logger.Debugf("{{ .Name }} res: +v%", res)

	return &{{ .Reply }}{
		// TODO 
	}, nil
}
{{- end }}
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
