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

func ListConfig(ctx context.Context, cmd *cli.Command) error {
	cfg, _ := config.Load()
	for name, conn := range cfg.Connections {
		fmt.Printf("- %s: %s (%s)\n", name, conn.DSN, conn.Driver)
	}
	return nil
}

func PingConfig(ctx context.Context, cmd *cli.Command) error {
	conn, err := config.GetConnection(cmd.String("name"))
	if err != nil {
		return err
	}
	err = config.TestConnection(conn.Driver, conn.DSN)
	if err != nil {
		fmt.Println("连接失败:", err)
	} else {
		fmt.Println("连接成功")
	}
	return nil
}
