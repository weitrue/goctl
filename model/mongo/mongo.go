package mongo

import (
	"errors"
	"path/filepath"
	"strings"

	"github.com/urfave/cli"
	"github.com/weitrue/goctl/config"
	"github.com/weitrue/goctl/model/mongo/generate"
	file "github.com/weitrue/goctl/util"
)

// Action provides the entry for goctl mongo code generation.
func Action(ctx *cli.Context) error {
	tp := ctx.StringSlice("type")
	c := ctx.Bool("cache")
	o := strings.TrimSpace(ctx.String("dir"))
	s := ctx.String("style")
	home := ctx.String("home")

	if len(home) > 0 {
		file.RegisterGoctlHome(home)
	}

	if len(tp) == 0 {
		return errors.New("missing type")
	}

	cfg, err := config.NewConfig(s)
	if err != nil {
		return err
	}

	a, err := filepath.Abs(o)
	if err != nil {
		return err
	}

	return generate.Do(&generate.Context{
		Types:  tp,
		Cache:  c,
		Output: a,
		Cfg:    cfg,
	})
}
