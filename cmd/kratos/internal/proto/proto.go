package proto

import (
	"github.com/aide-cloud/kratos/cmd/kratos/v2/internal/proto/add"
	"github.com/aide-cloud/kratos/cmd/kratos/v2/internal/proto/biz"
	"github.com/aide-cloud/kratos/cmd/kratos/v2/internal/proto/client"
	"github.com/aide-cloud/kratos/cmd/kratos/v2/internal/proto/data"
	"github.com/aide-cloud/kratos/cmd/kratos/v2/internal/proto/gateway"
	"github.com/aide-cloud/kratos/cmd/kratos/v2/internal/proto/graphql"
	"github.com/aide-cloud/kratos/cmd/kratos/v2/internal/proto/server"
	"github.com/aide-cloud/kratos/cmd/kratos/v2/internal/proto/types"
	"github.com/spf13/cobra"
)

// CmdProto represents the proto command.
var CmdProto = &cobra.Command{
	Use:   "proto",
	Short: "Generate the proto files",
	Long:  "Generate the proto files.",
}

func init() {
	CmdProto.AddCommand(add.CmdAdd)
	CmdProto.AddCommand(client.CmdClient)
	CmdProto.AddCommand(server.CmdServer)
	CmdProto.AddCommand(biz.CmdServer)
	CmdProto.AddCommand(data.CmdServer)
	CmdProto.AddCommand(graphql.CmdServer)
	CmdProto.AddCommand(gateway.CmdServer)
	CmdProto.AddCommand(types.CmdServer)
}
