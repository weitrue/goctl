package tsgen

import (
	"errors"
	"fmt"

	"github.com/logrusorgru/aurora"
	"github.com/urfave/cli"
	"github.com/weitrue/goctl/api/parser"
	"github.com/weitrue/goctl/util"
	"github.com/zeromicro/go-zero/core/logx"
)

// TsCommand provides the entry to generate typescript codes
func TsCommand(c *cli.Context) error {
	apiFile := c.String("api")
	dir := c.String("dir")
	webAPI := c.String("webapi")
	caller := c.String("caller")
	unwrapAPI := c.Bool("unwrap")
	if len(apiFile) == 0 {
		return errors.New("missing -api")
	}

	if len(dir) == 0 {
		return errors.New("missing -dir")
	}

	api, err := parser.Parse(apiFile)
	if err != nil {
		fmt.Println(aurora.Red("Failed"))
		return err
	}

	logx.Must(util.MkdirIfNotExist(dir))
	logx.Must(genHandler(dir, webAPI, caller, api, unwrapAPI))
	logx.Must(genComponents(dir, api))

	fmt.Println(aurora.Green("Done."))
	return nil
}
