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
			batch INT NOT NULL,
            executed_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
        )
    `)
	if err != nil {
		return fmt.Errorf("创建迁移表报错: %w", err)
	}

	// 获取 migration 文件列表
	files, err := filepath.Glob(cmd.String("path") + "/*.up.sql")
	if err != nil {
		return fmt.Errorf("获取迁移文件报错: %w", err)
	}
	sort.Strings(files)

	nextBatch, err := getNextBatch(db)
	if err != nil {
		return fmt.Errorf("获取下一个批次报错: %w", err)
	}

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
		_, err = db.Exec("INSERT INTO migrations(filename, batch) VALUES(?, ?)", filepath.Base(file), nextBatch)
		if err != nil {
			return fmt.Errorf("添加迁移记录%s报错: %w", file, err)
		}

		fmt.Println("执行成功:", file)
	}

	return nil
}

func DownMigrate(ctx context.Context, cmd *cli.Command) error {
	conn, err := config.GetConnection(cmd.String("name"))
	if err != nil {
		return err
	}

	db, err := sql.Open(conn.Driver, conn.DSN)
	if err != nil {
		return fmt.Errorf("打开连接报错: %w", err)
	}
	defer db.Close()

	lastBatch, err := getLastbatch(db)
	if err != nil {
		return fmt.Errorf("获取最后批次报错: %w", err)
	}

	rows, err := db.Query("SELECT filename FROM migrations where batch = ? ORDER BY executed_at DESC", lastBatch)
	if err != nil {
		return fmt.Errorf("查询迁移记录报错: %w", err)
	}
	defer rows.Close()

	var files []string
	for rows.Next() {
		var f string
		if err := rows.Scan(&f); err != nil {
			return fmt.Errorf("扫描迁移记录报错: %w", err)
		}
		files = append(files, f)
	}

	// 逆序执行 down
	for _, upFile := range files {
		base := upFile[:len(upFile)-len(".up.sql")]
		downFile := filepath.Join(cmd.String("path"), base+".down.sql")

		content, err := ioutil.ReadFile(downFile)
		if err != nil {
			return fmt.Errorf("读取回滚迁移文件%s报错: %w", downFile, err)
		}

		_, err = db.Exec(string(content))
		if err != nil {
			return fmt.Errorf("执行回滚迁移文件%s报错: %w", downFile, err)
		}

		// 删除 migrations 表记录
		_, err = db.Exec("DELETE FROM migrations WHERE filename = ? and batch = ?", upFile, lastBatch)
		if err != nil {
			return fmt.Errorf("删除迁移记录%s报错: %w", upFile, err)
		}

		fmt.Println("回滚成功", upFile)
	}

	return nil
}

func getNextBatch(db *sql.DB) (int, error) {
	batch, err := getLastbatch(db)
	if err != nil {
		return 0, err
	}
	return batch + 1, nil
}

func getLastbatch(db *sql.DB) (int, error) {
	var maxBatch sql.NullInt64
	query := "SELECT MAX(batch) FROM migrations"
	err := db.QueryRow(query).Scan(&maxBatch)
	if err != nil {
		return 0, err
	}

	if !maxBatch.Valid {
		return 0, nil
	}
	return int(maxBatch.Int64), nil
}
