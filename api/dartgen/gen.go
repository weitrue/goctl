package dartgen

import (
	"errors"
	"strings"

	"github.com/urfave/cli"
	"github.com/weitrue/goctl/api/parser"
	"github.com/zeromicro/go-zero/core/logx"
)

// DartCommand create dart network request code
func DartCommand(c *cli.Context) error {
	apiFile := c.String("api")
	dir := c.String("dir")
	if len(apiFile) == 0 {
		return errors.New("missing -api")
	}
	if len(dir) == 0 {
		return errors.New("missing -dir")
	}

	api, err := parser.Parse(apiFile)
	if err != nil {
		return err
	}

	if !strings.HasSuffix(dir, "/") {
		dir = dir + "/"
	}
	api.Info.Title = strings.Replace(apiFile, ".api", "", -1)
	logx.Must(genData(dir+"data/", api))
	logx.Must(genApi(dir+"api/", api))
	logx.Must(genVars(dir + "vars/"))
	return nil
}
