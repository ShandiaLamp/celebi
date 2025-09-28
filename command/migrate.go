package command

import (
	"context"
	"database/sql"
	"fmt"
	"io/ioutil"
	"path/filepath"
	"sort"

	"github.com/shandialamp/celebi/config"
	"github.com/urfave/cli/v3"
)

func UpMigrate(ctx context.Context, cmd *cli.Command) error {
	conn, err := config.GetConnection(cmd.String("name"))
	if err != nil {
		return err
	}

	db, err := sql.Open(conn.Driver, conn.DSN)
	if err != nil {
		return fmt.Errorf("打开连接报错: %w", err)
	}
	defer db.Close()

	// 创建 migrations 表
	_, err = db.Exec(`
        CREATE TABLE IF NOT EXISTS migrations (
            id SERIAL PRIMARY KEY,
            filename VARCHAR(255) NOT NULL UNIQUE,
            executed_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
        )
    `)
	if err != nil {
		return fmt.Errorf("创建迁移表报错: %w", err)
	}

	// 获取 migration 文件列表
	files, err := filepath.Glob(cmd.String("path") + "/*.sql")
	if err != nil {
		return fmt.Errorf("获取迁移文件报错: %w", err)
	}
	sort.Strings(files)

	for _, file := range files {
		var exists int
		err := db.QueryRow("SELECT COUNT(1) FROM migrations WHERE filename = ?", filepath.Base(file)).Scan(&exists)
		if err != nil {
			return fmt.Errorf("查询迁移记录%s报错: %w", file, err)
		}
		if exists > 0 {
			fmt.Println("跳过已经迁移的文件:", file)
			continue
		}
		content, err := ioutil.ReadFile(file)
		if err != nil {
			return fmt.Errorf("读取%s出错: %w", file, err)
		}

		_, err = db.Exec(string(content))
		if err != nil {
			return fmt.Errorf("执行%s出错: %w", file, err)
		}
		_, err = db.Exec("INSERT INTO migrations(filename) VALUES(?)", filepath.Base(file))
		if err != nil {
			return fmt.Errorf("添加迁移记录%s报错: %w", file, err)
		}

		fmt.Println("执行成功:", file)
	}

	return nil
}
