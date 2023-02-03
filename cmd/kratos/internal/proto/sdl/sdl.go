package sdl

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
	Use:   "sdl",
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
		proto.WithMessage(func(m *proto.Message) {
			message := &Message{
				Name: m.Name,
			}
			for _, e := range m.Elements {
				field, ok := e.(*proto.NormalField)
				if ok {
					// 自定义类型
					if field.InlineComment != nil {
						customTypeName := parseSdlType(field.InlineComment.Message())
						if customTypeName != "" {
							field.Type = customTypeName
						}
					}
					message.addField(&Field{
						Name:     field.Name,
						Type:     parseFieldType(field.Type),
						Repeated: field.Repeated,
						Required: true,
					})
					continue
				}

				mapField, ok := e.(*proto.MapField)
				if ok {
					// 自定义类型
					if mapField.InlineComment != nil {
						customTypeName := parseSdlType(mapField.InlineComment.Message())
						if customTypeName != "" {
							mapField.Type = customTypeName
						}
					}
					message.addField(&Field{
						Name:     mapField.Name,
						Type:     parseFieldType(mapField.Type),
						Repeated: false,
						Required: true,
					})
					continue
				}
			}
			if m.Comment != nil {
				msg.addMessage(parseFieldMode(m.Comment.Message()), message)
			}
		}),
		proto.WithRPC(func(r *proto.RPC) {
			if r.Comment != nil {
				msg.addMethod(parseMethodMode(r.Comment.Message()), &Method{
					Name:    r.Name,
					Request: serviceName(r.RequestType),
					Reply:   r.ReturnsType,
				})
			}
		}),
	)

	if _, err := os.Stat(targetDir); os.IsNotExist(err) {
		fmt.Printf("Target directory: %s does not exsit\n", targetDir)
		return
	}

	{
		moduleName := strings.ToLower(parseFileName(args[0]))
		fileName := strings.ToLower(moduleName + ".sdl.graphql")
		to := path.Join(targetDir, fileName)
		//if _, err := os.Stat(to); !os.IsNotExist(err) {
		//	fmt.Fprintf(os.Stderr, "%s already exists: %s\n", fileName, to)
		//}
		msg.Source = moduleName + ".proto"
		msg.Module = toUpperCamelCase(moduleName)
		b, err := msg.execute()
		if err != nil {
			log.Fatal(err)
		}
		if err := os.WriteFile(to, b, 0o644); err != nil {
			log.Println("write file error: ", to)
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
