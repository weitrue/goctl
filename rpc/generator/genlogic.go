package generator

import (
	"fmt"
	"path/filepath"
	"strings"

	conf "github.com/weitrue/goctl/config"
	"github.com/weitrue/goctl/rpc/parser"
	"github.com/weitrue/goctl/util"
	"github.com/weitrue/goctl/util/format"
	"github.com/weitrue/goctl/util/stringx"
	"github.com/zeromicro/go-zero/core/collection"
)

const (
	logicTemplate = `package logic

import (
	"context"

	{{.imports}}

	"github.com/zeromicro/go-zero/core/logx"
)

type {{.logicName}} struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func New{{.logicName}}(ctx context.Context,svcCtx *svc.ServiceContext) *{{.logicName}} {
	return &{{.logicName}}{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}
{{.functions}}
`
	logicFunctionTemplate = `{{if .hasComment}}{{.comment}}{{end}}
func (l *{{.logicName}}) {{.method}} ({{if .hasReq}}in {{.request}}{{if .stream}},stream {{.streamBody}}{{end}}{{else}}stream {{.streamBody}}{{end}}) ({{if .hasReply}}{{.response}},{{end}} error) {
	// todo: add your logic here and delete this line
	
	return {{if .hasReply}}&{{.responseType}}{},{{end}} nil
}
`
)

// GenLogic generates the logic file of the rpc service, which corresponds to the RPC definition items in proto.
func (g *DefaultGenerator) GenLogic(ctx DirContext, proto parser.Proto, cfg *conf.Config) error {
	dir := ctx.GetLogic()
	service := proto.Service.Service.Name
	for _, rpc := range proto.Service.RPC {
		logicFilename, err := format.FileNamingFormat(cfg.NamingFormat, rpc.Name+"_logic")
		if err != nil {
			return err
		}

		filename := filepath.Join(dir.Filename, logicFilename+".go")
		functions, err := g.genLogicFunction(service, proto.PbPackage, rpc)
		if err != nil {
			return err
		}

		comment := "业务逻辑"
		if fs := strings.Fields(parser.GetComment(rpc.Doc())); len(fs) > 2 {
			comment = fs[2]
		}

		imports := collection.NewSet()
		imports.AddStr(fmt.Sprintf(`"%v"`, ctx.GetSvc().Package))
		imports.AddStr(fmt.Sprintf(`"%v"`, ctx.GetPb().Package))
		text, err := util.LoadTemplate(category, logicTemplateFileFile, logicTemplate)
		if err != nil {
			return err
		}
		err = util.With("logic").GoFmt(true).Parse(text).SaveTo(map[string]interface{}{
			"logicName": fmt.Sprintf("%sLogic", stringx.From(rpc.Name).ToCamel()),
			"functions": functions,
			"imports":   strings.Join(imports.KeysStr(), util.NL),
			"comment":   comment,
		}, filename, false)
		if err != nil {
			return err
		}
	}
	return nil
}

func (g *DefaultGenerator) genLogicFunction(serviceName, goPackage string, rpc *parser.RPC) (string, error) {
	functions := make([]string, 0)
	text, err := util.LoadTemplate(category, logicFuncTemplateFileFile, logicFunctionTemplate)
	if err != nil {
		return "", err
	}

	logicName := stringx.From(rpc.Name + "_logic").ToCamel()
	comment := parser.GetComment(rpc.Doc())
	streamServer := fmt.Sprintf("%s.%s_%s%s", goPackage, parser.CamelCase(serviceName), parser.CamelCase(rpc.Name), "Server")
	buffer, err := util.With("fun").Parse(text).Execute(map[string]interface{}{
		"logicName":    logicName,
		"method":       parser.CamelCase(rpc.Name),
		"hasReq":       !rpc.StreamsRequest,
		"request":      fmt.Sprintf("*%s.%s", goPackage, parser.CamelCase(rpc.RequestType)),
		"hasReply":     !rpc.StreamsRequest && !rpc.StreamsReturns,
		"response":     fmt.Sprintf("*%s.%s", goPackage, parser.CamelCase(rpc.ReturnsType)),
		"responseType": fmt.Sprintf("%s.%s", goPackage, parser.CamelCase(rpc.ReturnsType)),
		"stream":       rpc.StreamsRequest || rpc.StreamsReturns,
		"streamBody":   streamServer,
		"hasComment":   len(comment) > 0,
		"comment":      comment,
	})
	if err != nil {
		return "", err
	}

	functions = append(functions, buffer.String())
	return strings.Join(functions, util.NL), nil
}
