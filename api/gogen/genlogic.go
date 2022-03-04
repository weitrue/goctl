package gogen

import (
	"fmt"
	"path"
	"strings"

	"github.com/weitrue/goctl/api/spec"
	"github.com/weitrue/goctl/config"
	ctlutil "github.com/weitrue/goctl/util"
	"github.com/weitrue/goctl/util/format"
	"github.com/weitrue/goctl/vars"
)

const logicTemplate = `package {{.pkgName}}

import (
	{{.imports}}
)

type {{.logic}} struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func New{{.logic}}(ctx context.Context, svcCtx *svc.ServiceContext) {{.logic}} {
	return {{.logic}}{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *{{.logic}}) {{.function}}({{.request}}) {{.responseType}} {
	// todo: add your logic here and delete this line

	{{.returnString}}
}
`

func genLogic(dir, rootPkg string, cfg *config.Config, api *spec.ApiSpec) error {
	for _, g := range api.Service.Groups {
		for _, r := range g.Routes {
			err := genLogicByRoute(dir, rootPkg, cfg, g, r)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func genLogicByRoute(dir, rootPkg string, cfg *config.Config, group spec.Group, route spec.Route) error {
	logic := getLogicName(route)
	goFile, err := format.FileNamingFormat(cfg.NamingFormat, logic)
	if err != nil {
		return err
	}

	imports := genLogicImports(route, rootPkg)
	var responseString string
	var returnString string
	var requestString string
	if len(route.ResponseTypeName()) > 0 {
		resp := responseGoTypeName(route, typesPacket)
		responseString = "(" + resp + ", error)"
		if strings.HasPrefix(resp, "*") {
			returnString = fmt.Sprintf("return &%s{}, nil", strings.TrimPrefix(resp, "*"))
		} else {
			returnString = fmt.Sprintf("return %s{}, nil", resp)
		}
	} else {
		responseString = "error"
		returnString = "return nil"
	}
	if len(route.RequestTypeName()) > 0 {
		requestString = "req " + requestGoTypeName(route, typesPacket)
	}
	summary := "业务逻辑"
	if route.AtDoc.Properties != nil {
		summary = strings.TrimSuffix(strings.TrimPrefix(route.AtDoc.Properties["summary"], "\""), "\"")
	}

	subDir := getLogicFolderPath(group, route)
	return genFile(fileGenConfig{
		dir:             dir,
		subdir:          subDir,
		filename:        goFile + ".go",
		templateName:    "logicTemplate",
		category:        category,
		templateFile:    logicTemplateFile,
		builtinTemplate: logicTemplate,
		data: map[string]string{
			"pkgName":      subDir[strings.LastIndex(subDir, "/")+1:],
			"imports":      imports,
			"logic":        strings.Title(logic),
			"function":     strings.Title(strings.TrimSuffix(logic, "Logic")),
			"responseType": responseString,
			"returnString": returnString,
			"request":      requestString,
			"summary":      summary,
		},
	})
}

func getLogicFolderPath(group spec.Group, route spec.Route) string {
	folder := route.GetAnnotation(groupProperty)
	if len(folder) == 0 {
		folder = group.GetAnnotation(groupProperty)
		if len(folder) == 0 {
			return logicDir
		}
	}
	folder = strings.TrimPrefix(folder, "/")
	folder = strings.TrimSuffix(folder, "/")
	return path.Join(logicDir, folder)
}

func genLogicImports(route spec.Route, parentPkg string) string {
	var imports []string
	imports = append(imports, `"context"`+"\n")
	imports = append(imports, fmt.Sprintf("\"%s/core/logx\"\n", vars.ProjectOpenSourceURL))
	imports = append(imports, fmt.Sprintf("\"%s\"", ctlutil.JoinPackages(parentPkg, contextDir)))
	if len(route.ResponseTypeName()) > 0 || len(route.RequestTypeName()) > 0 {
		imports = append(imports, fmt.Sprintf("\"%s\"", ctlutil.JoinPackages(parentPkg, typesDir)))
	}
	return strings.Join(imports, "\n\t")
}
