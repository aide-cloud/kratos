package server

import (
	"bytes"
	"html/template"
)

//nolint:lll
var serviceTemplate = `
{{- /* delete empty line */ -}}
package service

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
	{{ .Service }}Service struct {
		pb.Unimplemented{{ .Service }}Server
	
        logger *log.Helper
		logic {{ .Service }}LogicInterface
	}
	
	{{ .Service }}GraphqlService struct {
		*{{ .Service }}Service
	}

	{{ .Service }}LogicInterface interface {
		{{- $s1 := "google.protobuf.Empty" }}
		{{ range .Methods }}
		{{- if eq .Type 1 }}
		{{ .Name }}(ctx context.Context, req {{ if eq .Request $s1 }}*emptypb.Empty{{ else }}*pb.{{ .Request }}{{ end }}) ({{ if eq .Reply $s1 }}*emptypb.Empty{{ else }}*pb.{{ .Reply }}{{ end }}, error)
		{{- end }}
		{{- end }}
	}
)

func New{{ .Service }}Service(logic {{ .Service }}LogicInterface, logger log.Logger) *{{ .Service }}Service {
	return &{{ .Service }}Service{logger: log.NewHelper(logger), logic: logic}
}

func New{{ .Service }}GraphqlService(s *{{ .Service }}Service) *{{ .Service }}GraphqlService {
	return &{{ .Service }}GraphqlService{ {{ .Service }}Service: s }
}

{{- $s1 := "google.protobuf.Empty" }}
{{ range .Methods }}
{{- if eq .Type 1 }}
func (s *{{ .Service }}GraphqlService) {{ .Name }}(ctx context.Context, args struct {
	In {{ if eq .Request $s1 }}*emptypb.Empty{{ else }}*pb.{{ .Request }}{{ end }}
}) ({{ if eq .Reply $s1 }}*emptypb.Empty{{ else }}*pb.{{ .Reply }}{{ end }}, error) {
	return s.{{ .Service }}Service.{{ .Name }}(ctx, args.In)
}
{{- end }}
{{- end }}


{{- $s1 := "google.protobuf.Empty" }}
{{ range .Methods }}
{{- if eq .Type 1 }}
func (s *{{ .Service }}Service) {{ .Name }}(ctx context.Context, req {{ if eq .Request $s1 }}*emptypb.Empty{{ else }}*pb.{{ .Request }}{{ end }}) ({{ if eq .Reply $s1 }}*emptypb.Empty{{ else }}*pb.{{ .Reply }}{{ end }}, error) {
	return Validate(ctx, req, s.logic.{{ .Name }})
}

{{- else if eq .Type 2 }}
func (s *{{ .Service }}Service) {{ .Name }}(conn pb.{{ .Service }}_{{ .Name }}Server) error {
	for {
		req, err := conn.Recv()
		if err == io.EOF {
			return nil
		}
		if err != nil {
			return err
		}
		
		err = conn.Send(&pb.{{ .Reply }}{})
		if err != nil {
			return err
		}
	}
}

{{- else if eq .Type 3 }}
func (s *{{ .Service }}Service) {{ .Name }}(conn pb.{{ .Service }}_{{ .Name }}Server) error {
	for {
		req, err := conn.Recv()
		if err == io.EOF {
			return conn.SendAndClose(&pb.{{ .Reply }}{})
		}
		if err != nil {
			return err
		}
	}
}

{{- else if eq .Type 4 }}
func (s *{{ .Service }}Service) {{ .Name }}(req {{ if eq .Request $s1 }}*emptypb.Empty
{{ else }}*pb.{{ .Request }}{{ end }}, conn pb.{{ .Service }}_{{ .Name }}Server) error {
	for {
		err := conn.Send(&pb.{{ .Reply }}{})
		if err != nil {
			return err
		}
	}
}

{{- end }}
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
