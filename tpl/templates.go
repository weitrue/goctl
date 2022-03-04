package tpl

import (
	"fmt"
	"path/filepath"

	"github.com/logrusorgru/aurora"
	"github.com/urfave/cli"
	"github.com/weitrue/goctl/api/apigen"
	"github.com/weitrue/goctl/api/gogen"
	"github.com/weitrue/goctl/docker"
	"github.com/weitrue/goctl/kube"
	mongogen "github.com/weitrue/goctl/model/mongo/generate"
	modelgen "github.com/weitrue/goctl/model/sql/gen"
	rpcgen "github.com/weitrue/goctl/rpc/generator"
	"github.com/weitrue/goctl/util"
	"github.com/zeromicro/go-zero/core/errorx"
)

const templateParentPath = "/"

// GenTemplates writes the latest template text into file which is not exists
func GenTemplates(ctx *cli.Context) error {
	path := ctx.String("home")
	if len(path) != 0 {
		util.RegisterGoctlHome(path)
	}

	if err := errorx.Chain(
		func() error {
			return gogen.GenTemplates(ctx)
		},
		func() error {
			return modelgen.GenTemplates(ctx)
		},
		func() error {
			return rpcgen.GenTemplates(ctx)
		},
		func() error {
			return docker.GenTemplates(ctx)
		},
		func() error {
			return kube.GenTemplates(ctx)
		},
		func() error {
			return mongogen.Templates(ctx)
		},
		func() error {
			return apigen.GenTemplates(ctx)
		},
	); err != nil {
		return err
	}

	dir, err := util.GetTemplateDir(templateParentPath)
	if err != nil {
		return err
	}

	abs, err := filepath.Abs(dir)
	if err != nil {
		return err
	}

	fmt.Printf("Templates are generated in %s, %s\n", aurora.Green(abs),
		aurora.Red("edit on your risk!"))

	return nil
}

// CleanTemplates deletes all templates
func CleanTemplates(ctx *cli.Context) error {
	path := ctx.String("home")
	if len(path) != 0 {
		util.RegisterGoctlHome(path)
	}

	err := errorx.Chain(
		func() error {
			return gogen.Clean()
		},
		func() error {
			return modelgen.Clean()
		},
		func() error {
			return rpcgen.Clean()
		},
		func() error {
			return docker.Clean()
		},
		func() error {
			return kube.Clean()
		},
		func() error {
			return mongogen.Clean()
		},
		func() error {
			return apigen.Clean()
		},
	)
	if err != nil {
		return err
	}

	fmt.Printf("%s\n", aurora.Green("template are clean!"))
	return nil
}

// UpdateTemplates writes the latest template text into file,
// it will delete the older templates if there are exists
func UpdateTemplates(ctx *cli.Context) (err error) {
	path := ctx.String("home")
	category := ctx.String("category")
	if len(path) != 0 {
		util.RegisterGoctlHome(path)
	}

	defer func() {
		if err == nil {
			fmt.Println(aurora.Green(fmt.Sprintf("%s template are update!", category)).String())
		}
	}()
	switch category {
	case docker.Category():
		return docker.Update()
	case gogen.Category():
		return gogen.Update()
	case kube.Category():
		return kube.Update()
	case rpcgen.Category():
		return rpcgen.Update()
	case modelgen.Category():
		return modelgen.Update()
	case mongogen.Category():
		return mongogen.Update()
	case apigen.Category():
		return apigen.Update()
	default:
		err = fmt.Errorf("unexpected category: %s", category)
		return
	}
}

// RevertTemplates will overwrite the old template content with the new template
func RevertTemplates(ctx *cli.Context) (err error) {
	path := ctx.String("home")
	category := ctx.String("category")
	filename := ctx.String("name")
	if len(path) != 0 {
		util.RegisterGoctlHome(path)
	}

	defer func() {
		if err == nil {
			fmt.Println(aurora.Green(fmt.Sprintf("%s template are reverted!", filename)).String())
		}
	}()
	switch category {
	case docker.Category():
		return docker.RevertTemplate(filename)
	case kube.Category():
		return kube.RevertTemplate(filename)
	case gogen.Category():
		return gogen.RevertTemplate(filename)
	case rpcgen.Category():
		return rpcgen.RevertTemplate(filename)
	case modelgen.Category():
		return modelgen.RevertTemplate(filename)
	case mongogen.Category():
		return mongogen.RevertTemplate(filename)
	case apigen.Category():
		return apigen.RevertTemplate(filename)
	default:
		err = fmt.Errorf("unexpected category: %s", category)
		return
	}
}
