package cmd

import (
	"database/sql"
	"fmt"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/sqlite3"
	"github.com/golang-migrate/migrate/v4/source/iofs"
	"github.com/pkg/errors"
	"github.com/rfancn/prism/assets"
)

// runMigrations runs the database schema migrations using golang-migrate.
func runMigrations(db *sql.DB) error {
	// 从嵌入的文件系统创建迁移源
	source, err := iofs.New(assets.MigrationsFS, "migrations")
	if err != nil {
		return fmt.Errorf("创建迁移源失败: %w", err)
	}

	// 创建 sqlite3 驱动
	driver, err := sqlite3.WithInstance(db, &sqlite3.Config{})
	if err != nil {
		return fmt.Errorf("创建数据库驱动失败: %w", err)
	}

	// 创建 migrate 实例
	m, err := migrate.NewWithInstance("iofs", source, "sqlite3", driver)
	if err != nil {
		return fmt.Errorf("创建迁移实例失败: %w", err)
	}

	// 执行迁移
	if err = m.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return fmt.Errorf("执行迁移失败: %w", err)
	}

	return nil
}
