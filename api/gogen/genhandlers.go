package gogen

import (
	"fmt"
	"path"
	"strings"

	"github.com/weitrue/goctl/api/spec"
	"github.com/weitrue/goctl/config"
	"github.com/weitrue/goctl/internal/version"
	"github.com/weitrue/goctl/util"
	"github.com/weitrue/goctl/util/format"
	"github.com/weitrue/goctl/vars"
)

const (
	defaultLogicPackage = "logic"
	handlerTemplate     = `package {{.PkgName}}

import (
	"net/http"

	{{if .After1_1_10}}"github.com/zeromicro/go-zero/rest/httpx"{{end}}
	{{.ImportPackages}}
)

func {{.HandlerName}}(ctx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		{{if .HasRequest}}var req types.{{.RequestType}}
		if err := httpx.Parse(r, &req); err != nil {
			httpx.Error(w, err)
			return
		}

		{{end}}l := {{.LogicName}}.New{{.LogicType}}(r.Context(), ctx)
		{{if .HasResp}}resp, {{end}}err := l.{{.Call}}({{if .HasRequest}}req{{end}})
		if err != nil {
			httpx.Error(w, err)
		} else {
			{{if .HasResp}}httpx.OkJson(w, resp){{else}}httpx.Ok(w){{end}}
		}
	}
}
`
)

type handlerInfo struct {
	PkgName        string
	ImportPackages string
	HandlerName    string
	PathName       string
	MethodName     string
	Tag            string
	Summary        string
	ResponseType   string
	RequestType    string
	LogicName      string
	LogicType      string
	Call           string
	HasResp        bool
	HasRequest     bool
	HasSecurity    bool
	After1_1_10    bool
}

func genHandler(dir, rootPkg string, cfg *config.Config, group spec.Group, route spec.Route) error {
	handler := getHandlerName(route)
	handlerPath := getHandlerFolderPath(group, route)
	pkgName := handlerPath[strings.LastIndex(handlerPath, "/")+1:]
	logicName := defaultLogicPackage
	if handlerPath != handlerDir {
		handler = strings.Title(handler)
		logicName = pkgName
	}
	parentPkg, err := getParentPackage(dir)
	if err != nil {
		return err
	}
	tag := "Tag"
	if a := group.GetAnnotation("tag"); a != "" {
		tag = strings.TrimSuffix(strings.TrimPrefix(a, "\""), "\"")
	}
	summary := "Summary"
	if route.AtDoc.Properties != nil {
		summary = strings.TrimSuffix(strings.TrimPrefix(route.AtDoc.Properties["summary"], "\""), "\"")
	}
	hasSecurity := false
	if ja := group.GetAnnotation("jwt"); ja != "" {
		hasSecurity = true
	} else if ma := strings.ToLower(group.GetAnnotation("middleware")); ma != "" {
		if strings.Contains(ma, "jwt") || strings.Contains(ma, "auth") {
			hasSecurity = true
		}
	}

	goctlVersion := version.GetGoctlVersion()
	// todo(anqiansong): This will be removed after a certain number of production versions of goctl (probably 5)
	after1_1_10 := version.IsVersionGreaterThan(goctlVersion, "1.1.10")
	return doGenToFile(dir, handler, cfg, group, route, handlerInfo{
		PkgName:        pkgName,
		ImportPackages: genHandlerImports(group, route, parentPkg),
		HandlerName:    handler,
		PathName:       strings.TrimSpace(route.Path),
		MethodName:     strings.ToLower(strings.TrimSpace(route.Method)),
		Tag:            tag,
		Summary:        summary,
		ResponseType:   util.Title(route.ResponseTypeName()),
		RequestType:    util.Title(route.RequestTypeName()),
		LogicName:      logicName,
		LogicType:      strings.Title(getLogicName(route)),
		Call:           strings.Title(strings.TrimSuffix(handler, "Handler")),
		HasResp:        len(route.ResponseTypeName()) > 0,
		HasRequest:     len(route.RequestTypeName()) > 0,
		HasSecurity:    hasSecurity,
		After1_1_10:    after1_1_10,
	})
}

func doGenToFile(dir, handler string, cfg *config.Config, group spec.Group,
	route spec.Route, handleObj handlerInfo) error {
	filename, err := format.FileNamingFormat(cfg.NamingFormat, handler)
	if err != nil {
		return err
	}

	return genFile(fileGenConfig{
		dir:             dir,
		subdir:          getHandlerFolderPath(group, route),
		filename:        filename + ".go",
		templateName:    "handlerTemplate",
		category:        category,
		templateFile:    handlerTemplateFile,
		builtinTemplate: handlerTemplate,
		data:            handleObj,
	})
}

func genHandlers(dir, rootPkg string, cfg *config.Config, api *spec.ApiSpec) error {
	for _, group := range api.Service.Groups {
		for _, route := range group.Routes {
			if err := genHandler(dir, rootPkg, cfg, group, route); err != nil {
				return err
			}
		}
	}

	return nil
}

func genHandlerImports(group spec.Group, route spec.Route, parentPkg string) string {
	var imports []string
	imports = append(imports, fmt.Sprintf("\"%s\"",
		util.JoinPackages(parentPkg, getLogicFolderPath(group, route))))
	imports = append(imports, fmt.Sprintf("\"%s\"", util.JoinPackages(parentPkg, contextDir)))
	if len(route.RequestTypeName()) > 0 {
		imports = append(imports, fmt.Sprintf("\"%s\"\n", util.JoinPackages(parentPkg, typesDir)))
	}

	currentVersion := version.GetGoctlVersion()
	// todo(anqiansong): This will be removed after a certain number of production versions of goctl (probably 5)
	if !version.IsVersionGreaterThan(currentVersion, "1.1.10") {
		imports = append(imports, fmt.Sprintf("\"%s/rest/httpx\"", vars.ProjectOpenSourceURL))
	}

	return strings.Join(imports, "\n\t")
}

func getHandlerBaseName(route spec.Route) (string, error) {
	handler := route.Handler
	handler = strings.TrimSpace(handler)
	handler = strings.TrimSuffix(handler, "handler")
	handler = strings.TrimSuffix(handler, "Handler")
	return handler, nil
}

func getHandlerFolderPath(group spec.Group, route spec.Route) string {
	folder := route.GetAnnotation(groupProperty)
	if len(folder) == 0 {
		folder = group.GetAnnotation(groupProperty)
		if len(folder) == 0 {
			return handlerDir
		}
	}

	folder = strings.TrimPrefix(folder, "/")
	folder = strings.TrimSuffix(folder, "/")
	return path.Join(handlerDir, folder)
}

func getHandlerName(route spec.Route) string {
	handler, err := getHandlerBaseName(route)
	if err != nil {
		panic(err)
	}

	return handler + "Handler"
}

func getLogicName(route spec.Route) string {
	handler, err := getHandlerBaseName(route)
	if err != nil {
		panic(err)
	}

	return handler + "Logic"
}
