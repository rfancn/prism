// Package assets contains embedded assets for the Prism service.
package assets

import "embed"

// MigrationsFS 嵌入数据库迁移文件
//
//go:embed migrations/*.sql
var MigrationsFS embed.FS