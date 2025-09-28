package command

import (
	"context"
	"fmt"

	"github.com/shandialamp/celebi/config"
	"github.com/urfave/cli/v3"
)

func AddConfig(ctx context.Context, cmd *cli.Command) error {
	cfg, _ := config.Load()
	name := cmd.String("name")
	cfg.Connections[name] = config.Connection{
		Driver: cmd.String("driver"),
		DSN:    cmd.String("dsn"),
	}
	if cmd.Bool("default") || cfg.Default == "" {
		cfg.Default = name
	}
	config.Save(cfg)
	fmt.Println("添加连接成功:", name)
	return nil
}
