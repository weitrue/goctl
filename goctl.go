package main

import (
	"fmt"
	"os"
	"runtime"

	"github.com/logrusorgru/aurora"
	"github.com/urfave/cli"

	"github.com/zeromicro/go-zero/core/load"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stat"

	"github.com/weitrue/goctl/api/apigen"
	"github.com/weitrue/goctl/api/dartgen"
	"github.com/weitrue/goctl/api/docgen"
	"github.com/weitrue/goctl/api/format"
	"github.com/weitrue/goctl/api/gogen"
	"github.com/weitrue/goctl/api/javagen"
	"github.com/weitrue/goctl/api/ktgen"
	"github.com/weitrue/goctl/api/new"
	"github.com/weitrue/goctl/api/tsgen"
	"github.com/weitrue/goctl/api/validate"
	"github.com/weitrue/goctl/docker"
	"github.com/weitrue/goctl/internal/errorx"
	"github.com/weitrue/goctl/internal/version"
	"github.com/weitrue/goctl/kube"
	"github.com/weitrue/goctl/model/mongo"
	model "github.com/weitrue/goctl/model/sql/command"
	"github.com/weitrue/goctl/plugin"
	rpc "github.com/weitrue/goctl/rpc/cli"
	"github.com/weitrue/goctl/tpl"
	"github.com/weitrue/goctl/upgrade"
	"github.com/weitrue/goctl/util"
)

var commands = []cli.Command{
	{
		Name:   "upgrade",
		Usage:  "upgrade goctl to latest version",
		Action: upgrade.Upgrade,
	},
	{
		Name:  "api",
		Usage: "generate api related files",
		Flags: []cli.Flag{
			cli.StringFlag{
				Name:  "o",
				Usage: "the output api file",
			},
			cli.StringFlag{
				Name:  "home",
				Usage: "the goctl home path of the template",
			},
		},
		Action: apigen.ApiCommand,
		Subcommands: []cli.Command{
			{
				Name:   "new",
				Usage:  "fast create api service",
				Action: new.CreateServiceCommand,
			},
			{
				Name:  "format",
				Usage: "format api files",
				Flags: []cli.Flag{
					cli.StringFlag{
						Name:  "dir",
						Usage: "the format target dir",
					},
					cli.BoolFlag{
						Name:  "iu",
						Usage: "ignore update",
					},
					cli.BoolFlag{
						Name:  "stdin",
						Usage: "use stdin to input api doc content, press \"ctrl + d\" to send EOF",
					},
				},
				Action: format.GoFormatApi,
			},
			{
				Name:  "validate",
				Usage: "validate api file",
				Flags: []cli.Flag{
					cli.StringFlag{
						Name:  "api",
						Usage: "validate target api file",
					},
				},
				Action: validate.GoValidateApi,
			},
			{
				Name:  "doc",
				Usage: "generate doc files",
				Flags: []cli.Flag{
					cli.StringFlag{
						Name:  "dir",
						Usage: "the target dir",
					},
					cli.StringFlag{
						Name:     "o",
						Required: false,
						Usage:    "the output markdown directory",
					},
				},
				Action: docgen.DocCommand,
			},
			{
				Name:  "go",
				Usage: "generate go files for provided api in yaml file",
				Flags: []cli.Flag{
					cli.StringFlag{
						Name:  "dir",
						Usage: "the target dir",
					},
					cli.StringFlag{
						Name:  "api",
						Usage: "the api file",
					},
					cli.StringFlag{
						Name:  "style",
						Usage: "the file naming format, see [https://github.com/zeromicro/go-zero/tree/master/tools/goctl/config/readme.md]",
					},
					cli.StringFlag{
						Name:  "home",
						Usage: "the goctl home path of the template",
					},
				},
				Action: gogen.GoCommand,
			},
			{
				Name:  "java",
				Usage: "generate java files for provided api in api file",
				Flags: []cli.Flag{
					cli.StringFlag{
						Name:  "dir",
						Usage: "the target dir",
					},
					cli.StringFlag{
						Name:  "api",
						Usage: "the api file",
					},
				},
				Action: javagen.JavaCommand,
			},
			{
				Name:  "ts",
				Usage: "generate ts files for provided api in api file",
				Flags: []cli.Flag{
					cli.StringFlag{
						Name:  "dir",
						Usage: "the target dir",
					},
					cli.StringFlag{
						Name:  "api",
						Usage: "the api file",
					},
					cli.StringFlag{
						Name:  "webapi",
						Usage: "the web api file path",
					},
					cli.StringFlag{
						Name:  "caller",
						Usage: "the web api caller",
					},
					cli.BoolFlag{
						Name:  "unwrap",
						Usage: "unwrap the webapi caller for import",
					},
				},
				Action: tsgen.TsCommand,
			},
			{
				Name:  "dart",
				Usage: "generate dart files for provided api in api file",
				Flags: []cli.Flag{
					cli.StringFlag{
						Name:  "dir",
						Usage: "the target dir",
					},
					cli.StringFlag{
						Name:  "api",
						Usage: "the api file",
					},
				},
				Action: dartgen.DartCommand,
			},
			{
				Name:  "kt",
				Usage: "generate kotlin code for provided api file",
				Flags: []cli.Flag{
					cli.StringFlag{
						Name:  "dir",
						Usage: "the target directory",
					},
					cli.StringFlag{
						Name:  "api",
						Usage: "the api file",
					},
					cli.StringFlag{
						Name:  "pkg",
						Usage: "define package name for kotlin file",
					},
				},
				Action: ktgen.KtCommand,
			},
			{
				Name:  "plugin",
				Usage: "custom file generator",
				Flags: []cli.Flag{
					cli.StringFlag{
						Name:  "plugin, p",
						Usage: "the plugin file",
					},
					cli.StringFlag{
						Name:  "dir",
						Usage: "the target directory",
					},
					cli.StringFlag{
						Name:  "api",
						Usage: "the api file",
					},
					cli.StringFlag{
						Name:  "style",
						Usage: "the file naming format, see [https://github.com/zeromicro/go-zero/tree/master/tools/goctl/config/readme.md]",
					},
				},
				Action: plugin.PluginCommand,
			},
		},
	},
	{
		Name:  "docker",
		Usage: "generate Dockerfile",
		Flags: []cli.Flag{
			cli.StringFlag{
				Name:  "go",
				Usage: "the file that contains main function",
			},
			cli.IntFlag{
				Name:  "port",
				Usage: "the port to expose, default none",
				Value: 0,
			},
			cli.StringFlag{
				Name:  "home",
				Usage: "the goctl home path of the template",
			},
		},
		Action: docker.DockerCommand,
	},
	{
		Name:  "kube",
		Usage: "generate kubernetes files",
		Subcommands: []cli.Command{
			{
				Name:  "deploy",
				Usage: "generate deployment yaml file",
				Flags: []cli.Flag{
					cli.StringFlag{
						Name:     "name",
						Usage:    "the name of deployment",
						Required: true,
					},
					cli.StringFlag{
						Name:     "namespace",
						Usage:    "the namespace of deployment",
						Required: true,
					},
					cli.StringFlag{
						Name:     "image",
						Usage:    "the docker image of deployment",
						Required: true,
					},
					cli.StringFlag{
						Name:  "secret",
						Usage: "the secret to image pull from registry",
					},
					cli.IntFlag{
						Name:  "requestCpu",
						Usage: "the request cpu to deploy",
						Value: 500,
					},
					cli.IntFlag{
						Name:  "requestMem",
						Usage: "the request memory to deploy",
						Value: 512,
					},
					cli.IntFlag{
						Name:  "limitCpu",
						Usage: "the limit cpu to deploy",
						Value: 1000,
					},
					cli.IntFlag{
						Name:  "limitMem",
						Usage: "the limit memory to deploy",
						Value: 1024,
					},
					cli.StringFlag{
						Name:     "o",
						Usage:    "the output yaml file",
						Required: true,
					},
					cli.IntFlag{
						Name:  "replicas",
						Usage: "the number of replicas to deploy",
						Value: 3,
					},
					cli.IntFlag{
						Name:  "revisions",
						Usage: "the number of revision history to limit",
						Value: 5,
					},
					cli.IntFlag{
						Name:     "port",
						Usage:    "the port of the deployment to listen on pod",
						Required: true,
					},
					cli.IntFlag{
						Name:  "nodePort",
						Usage: "the nodePort of the deployment to expose",
						Value: 0,
					},
					cli.IntFlag{
						Name:  "minReplicas",
						Usage: "the min replicas to deploy",
						Value: 3,
					},
					cli.IntFlag{
						Name:  "maxReplicas",
						Usage: "the max replicas of deploy",
						Value: 10,
					},
					cli.StringFlag{
						Name:  "home",
						Usage: "the goctl home path of the template",
					},
				},
				Action: kube.DeploymentCommand,
			},
		},
	},
	{
		Name:  "rpc",
		Usage: "generate rpc code",
		Subcommands: []cli.Command{
			{
				Name:  "new",
				Usage: `generate rpc demo service`,
				Flags: []cli.Flag{
					cli.StringFlag{
						Name:  "style",
						Usage: "the file naming format, see [https://github.com/zeromicro/go-zero/tree/master/tools/goctl/config/readme.md]",
					},
					cli.BoolFlag{
						Name:  "idea",
						Usage: "whether the command execution environment is from idea plugin. [optional]",
					},
					cli.StringFlag{
						Name:  "home",
						Usage: "the goctl home path of the template",
					},
				},
				Action: rpc.RPCNew,
			},
			{
				Name:  "template",
				Usage: `generate proto template`,
				Flags: []cli.Flag{
					cli.StringFlag{
						Name:  "out, o",
						Usage: "the target path of proto",
					},
					cli.StringFlag{
						Name:  "home",
						Usage: "the goctl home path of the template",
					},
				},
				Action: rpc.RPCTemplate,
			},
			{
				Name:  "proto",
				Usage: `generate rpc from proto`,
				Flags: []cli.Flag{
					cli.StringFlag{
						Name:  "src, s",
						Usage: "the file path of the proto source file",
					},
					cli.StringSliceFlag{
						Name:  "proto_path, I",
						Usage: `native command of protoc, specify the directory in which to search for imports. [optional]`,
					},
					cli.StringSliceFlag{
						Name:  "go_opt",
						Usage: `native command of protoc-gen-go, specify the mapping from proto to go, eg --go_opt=proto_import=go_package_import. [optional]`,
					},
					cli.StringFlag{
						Name:  "dir, d",
						Usage: `the target path of the code`,
					},
					cli.StringFlag{
						Name:  "style",
						Usage: "the file naming format, see [https://github.com/zeromicro/go-zero/tree/master/tools/goctl/config/readme.md]",
					},
					cli.BoolFlag{
						Name:  "idea",
						Usage: "whether the command execution environment is from idea plugin. [optional]",
					},
					cli.StringFlag{
						Name:  "home",
						Usage: "the goctl home path of the template",
					},
				},
				Action: rpc.RPC,
			},
		},
	},
	{
		Name:  "model",
		Usage: "generate model code",
		Subcommands: []cli.Command{
			{
				Name:  "mysql",
				Usage: `generate mysql model`,
				Subcommands: []cli.Command{
					{
						Name:  "ddl",
						Usage: `generate mysql model from ddl`,
						Flags: []cli.Flag{
							cli.StringFlag{
								Name:  "src, s",
								Usage: "the path or path globbing patterns of the ddl",
							},
							cli.StringFlag{
								Name:  "dir, d",
								Usage: "the target dir",
							},
							cli.StringFlag{
								Name:  "style",
								Usage: "the file naming format, see [https://github.com/zeromicro/go-zero/tree/master/tools/goctl/config/readme.md]",
							},
							cli.BoolFlag{
								Name:  "cache, c",
								Usage: "generate code with cache [optional]",
							},
							cli.BoolFlag{
								Name:  "idea",
								Usage: "for idea plugin [optional]",
							},
							cli.StringFlag{
								Name:  "database, db",
								Usage: "the name of database [optional]",
							},
							cli.StringFlag{
								Name:  "home",
								Usage: "the goctl home path of the template",
							},
						},
						Action: model.MysqlDDL,
					},
					{
						Name:  "datasource",
						Usage: `generate model from datasource`,
						Flags: []cli.Flag{
							cli.StringFlag{
								Name:  "url",
								Usage: `the data source of database,like "root:password@tcp(127.0.0.1:3306)/database"`,
							},
							cli.StringFlag{
								Name:  "table, t",
								Usage: `the table or table globbing patterns in the database`,
							},
							cli.BoolFlag{
								Name:  "cache, c",
								Usage: "generate code with cache [optional]",
							},
							cli.StringFlag{
								Name:  "dir, d",
								Usage: "the target dir",
							},
							cli.StringFlag{
								Name:  "style",
								Usage: "the file naming format, see [https://github.com/zeromicro/go-zero/tree/master/tools/goctl/config/readme.md]",
							},
							cli.BoolFlag{
								Name:  "idea",
								Usage: "for idea plugin [optional]",
							},
							cli.StringFlag{
								Name:  "home",
								Usage: "the goctl home path of the template",
							},
						},
						Action: model.MySqlDataSource,
					},
				},
			},
			{
				Name:  "pg",
				Usage: `generate postgresql model`,
				Subcommands: []cli.Command{
					{
						Name:  "datasource",
						Usage: `generate model from datasource`,
						Flags: []cli.Flag{
							cli.StringFlag{
								Name:  "url",
								Usage: `the data source of database,like "postgres://root:password@127.0.0.1:54332/database?sslmode=disable"`,
							},
							cli.StringFlag{
								Name:  "table, t",
								Usage: `the table or table globbing patterns in the database`,
							},
							cli.StringFlag{
								Name:  "schema, s",
								Usage: `the table schema, default is [public]`,
							},
							cli.BoolFlag{
								Name:  "cache, c",
								Usage: "generate code with cache [optional]",
							},
							cli.StringFlag{
								Name:  "dir, d",
								Usage: "the target dir",
							},
							cli.StringFlag{
								Name:  "style",
								Usage: "the file naming format, see [https://github.com/zeromicro/go-zero/tree/master/tools/goctl/config/readme.md]",
							},
							cli.BoolFlag{
								Name:  "idea",
								Usage: "for idea plugin [optional]",
							},
							cli.StringFlag{
								Name:  "home",
								Usage: "the goctl home path of the template",
							},
						},
						Action: model.PostgreSqlDataSource,
					},
				},
			},
			{
				Name:  "mongo",
				Usage: `generate mongo model`,
				Flags: []cli.Flag{
					cli.StringSliceFlag{
						Name:  "type, t",
						Usage: "specified model type name",
					},
					cli.BoolFlag{
						Name:  "cache, c",
						Usage: "generate code with cache [optional]",
					},
					cli.StringFlag{
						Name:  "dir, d",
						Usage: "the target dir",
					},
					cli.StringFlag{
						Name:  "style",
						Usage: "the file naming format, see [https://github.com/zeromicro/go-zero/tree/master/tools/goctl/config/readme.md]",
					},
					cli.StringFlag{
						Name:  "home",
						Usage: "the goctl home path of the template",
					},
				},
				Action: mongo.Action,
			},
		},
	},
	{
		Name:  "template",
		Usage: "template operation",
		Subcommands: []cli.Command{
			{
				Name:  "init",
				Usage: "initialize the all templates(force update)",
				Flags: []cli.Flag{
					cli.StringFlag{
						Name:  "home",
						Usage: "the goctl home path of the template",
					},
				},
				Action: tpl.GenTemplates,
			},
			{
				Name:  "clean",
				Usage: "clean the all cache templates",
				Flags: []cli.Flag{
					cli.StringFlag{
						Name:  "home",
						Usage: "the goctl home path of the template",
					},
				},
				Action: tpl.CleanTemplates,
			},
			{
				Name:  "update",
				Usage: "update template of the target category to the latest",
				Flags: []cli.Flag{
					cli.StringFlag{
						Name:  "category,c",
						Usage: "the category of template, enum [api,rpc,model,docker,kube]",
					},
					cli.StringFlag{
						Name:  "home",
						Usage: "the goctl home path of the template",
					},
				},
				Action: tpl.UpdateTemplates,
			},
			{
				Name:  "revert",
				Usage: "revert the target template to the latest",
				Flags: []cli.Flag{
					cli.StringFlag{
						Name:  "category,c",
						Usage: "the category of template, enum [api,rpc,model,docker,kube]",
					},
					cli.StringFlag{
						Name:  "name,n",
						Usage: "the target file name of template",
					},
					cli.StringFlag{
						Name:  "home",
						Usage: "the goctl home path of the template",
					},
				},
				Action: tpl.RevertTemplates,
			},
		},
	},
}

func main() {
	logx.Disable()
	load.Disable()
	stat.DisableLog()

	app := cli.NewApp()
	app.Usage = "a cli tool to generate code"
	app.Version = fmt.Sprintf("%s %s/%s", version.BuildVersion, runtime.GOOS, runtime.GOARCH)
	app.Commands = commands
	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:  "template-folder, tf",
			Usage: "the goctl template folder",
		},
	}
	app.Before = func(c *cli.Context) error {
		tf := c.GlobalString("template-folder")
		if tf != "" {
			fmt.Printf("using goctl templates: %s\n", tf)
			util.TemplateFolder.Store(tf)
		}
		return nil
	}
	// cli already print error messages
	if err := app.Run(os.Args); err != nil {
		fmt.Println(aurora.Red(errorx.Wrap(err).Error()))
	}
}
