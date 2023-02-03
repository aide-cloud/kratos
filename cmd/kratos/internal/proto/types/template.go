package types

import (
	"bytes"
	"html/template"
)

//nolint:lll
var messageTemplate = `
{{- /* delete empty line */ -}}
package {{ .Package }}

import (
	{{- range .Packages }}
	"{{ . }}"
	{{- end }}
)

type (
	{{ range .Messages }}
	{{ .Name }}Type struct {
		{{- range .Fields }}
		{{ .Name }} {{ if .Repeated }}[]{{ end }}{{ .Type }}
		{{- end }}
	}
	{{- end }}
)
`

type Message struct {
	Name   string
	Fields []*Field
}

type Field struct {
	Name     string
	Type     string
	Repeated bool
}

type Msg struct {
	Package  string
	Packages []string
	Messages []*Message
}

func (l *Msg) execute() ([]byte, error) {
	buf := new(bytes.Buffer)
	tmpl, err := template.New("messages").Parse(messageTemplate)
	if err != nil {
		return nil, err
	}
	if err := tmpl.Execute(buf, l); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// Packages去重复
func (l *Msg) addPackage(p string) {
	if p == "" {
		return
	}

	for _, v := range l.Packages {
		if v == p {
			return
		}
	}
	l.Packages = append(l.Packages, p)
}
