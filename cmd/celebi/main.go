package main

import (
	"context"
	"log"
	"os"

	"github.com/shandialamp/celebi/command"
	"github.com/urfave/cli/v3"
)

func main() {
	cmd := &cli.Command{
		Name:  "celebi",
		Usage: "一个简易数据库迁移工具",
		Commands: []*cli.Command{
			{
				Name:  "config:add",
				Usage: "配置数据库连接",
				Flags: []cli.Flag{
					&cli.StringFlag{Name: "name", Usage: "连接名称", Required: true},
					&cli.StringFlag{Name: "driver", Usage: "mysql或者postgres", Required: true, DefaultText: "mysql"},
					&cli.StringFlag{Name: "dsn", Usage: "DSN", Required: true},
					&cli.BoolFlag{Name: "default", Usage: "是否默认"},
				},
				Action: command.AddConfig,
			},
		},
	}
	if err := cmd.Run(context.Background(), os.Args); err != nil {
		log.Fatal(err)
	}
}
