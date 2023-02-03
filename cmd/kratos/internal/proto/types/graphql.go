package types

import (
	"fmt"
	"log"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/emicklei/proto"
	"github.com/spf13/cobra"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

// CmdServer the service command.
var CmdServer = &cobra.Command{
	Use:   "types",
	Short: "Generate the proto Server implementations",
	Long:  "Generate the proto Server implementations. Example: kratos proto types api/xxx.proto -target-dir=internal/service",
	Run:   run,
}
var targetDir string

func init() {
	CmdServer.Flags().StringVarP(&targetDir, "target-dir", "t", "internal/service/graph", "generate target directory")
}

func run(cmd *cobra.Command, args []string) {
	if len(args) == 0 {
		fmt.Fprintln(os.Stderr, "Please specify the proto file. Example: kratos proto server api/xxx.proto")
		return
	}
	reader, err := os.Open(args[0])
	if err != nil {
		log.Fatal(err)
	}
	defer reader.Close()

	parser := proto.NewParser(reader)
	definition, err := parser.Parse()
	if err != nil {
		log.Fatal(err)
	}

	var (
		msg Msg
	)
	proto.Walk(definition,
		proto.WithOption(func(o *proto.Option) {
			if o.Name == "go_package" && msg.Package == "" {
				nameSlice := strings.Split(o.Constant.Source, ";")
				msg.Package = nameSlice[len(nameSlice)-1]
			}
		}),
		proto.WithPackage(func(p *proto.Package) {
			if msg.Package == "" {
				nameSlice := strings.Split(p.Name, ".")
				msg.Package = nameSlice[len(nameSlice)-1]
			}
		}),
		proto.WithMessage(func(m *proto.Message) {
			var ms = &Message{
				Name: m.Name,
			}

			for _, e := range m.Elements {
				field, ok := e.(*proto.NormalField)

				if ok {
					// 自定义类型
					if field.InlineComment != nil {
						customTypeName := parseType(field.InlineComment.Message())
						if customTypeName != "" {
							field.Type = customTypeName
						}
						msg.addPackage(parsePackage(field.InlineComment.Message()))
					}

					ms.Fields = append(ms.Fields, &Field{
						Name:     toUpperCamelCaseField(field.Name),
						Type:     field.Type,
						Repeated: field.Repeated,
					})
				}
			}
			msg.Messages = append(msg.Messages, ms)
		}),
	)

	if _, err := os.Stat(targetDir); os.IsNotExist(err) {
		fmt.Printf("Target directory: %s does not exsit\n", targetDir)
		return
	}

	{
		moduleName := strings.ToLower(parseFileName(args[0]))
		fileName := strings.ToLower(moduleName + ".pb.types.go")
		to := path.Join(targetDir, fileName)
		if _, err := os.Stat(to); !os.IsNotExist(err) {
			fmt.Fprintf(os.Stderr, "%s already exists: %s\n", fileName, to)
		}
		msg.Source = moduleName + ".proto"
		b, err := msg.execute()
		if err != nil {
			log.Fatal(err)
		}
		if err := os.WriteFile(to, b, 0o644); err != nil {
			log.Fatal(err)
		}
		fmt.Println(to)
	}
}

func serviceName(name string) string {
	return toUpperCamelCase(strings.Split(name, ".")[0])
}

func toUpperCamelCase(s string) string {
	s = strings.ReplaceAll(s, "_", " ")
	s = cases.Title(language.Und, cases.NoLower).String(s)
	return strings.ReplaceAll(s, " ", "")
}

// 解析// @gotags: uri:"id" @type: "int" 得到@type的value
func parseType(s string) string {
	return parseInlineComment(s, "@type")
}

func parsePackage(s string) string {
	return parseInlineComment(s, "@pkg")
}

// 解析路径文件名称并去掉后缀
func parseFileName(path string) string {
	_, file := filepath.Split(path)
	return strings.TrimSuffix(file, filepath.Ext(file))
}

func parseInlineComment(s, tag string) string {
	tagStr := tag + ":"
	if strings.Contains(s, tagStr) {
		ss := strings.Split(s, " ")
		for index, v := range ss {
			if v == tagStr && index+1 < len(ss) {
				res := ss[index+1]
				// 去掉双引号
				res = strings.ReplaceAll(res, "\"", "")
				return res
			}
		}
	}
	return ""
}

// field name转首字母大写的驼峰
func toUpperCamelCaseField(s string) string {
	s = strings.ReplaceAll(s, "_", " ")
	s = cases.Title(language.Und, cases.NoLower).String(s)
	return strings.ReplaceAll(s, " ", "")
}
