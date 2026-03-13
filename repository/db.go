package repository

import (
	"github.com/hdget/sdk"
	"github.com/rfancn/prism/autogen/db"
)

// New 创建查询器（使用默认连接）
func New() *db.Queries {
	return db.New(sdk.Db().My())
}