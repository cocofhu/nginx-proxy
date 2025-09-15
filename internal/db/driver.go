package db

import (
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// InitDB 初始化数据库连接
// 使用纯 Go SQLite 驱动，避免 CGO 编译问题
func InitDB(dbPath string) (*gorm.DB, error) {
	// 使用纯 Go SQLite 驱动
	db, err := gorm.Open(sqlite.Open(dbPath), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		return nil, err
	}

	// 自动迁移数据表
	err = db.AutoMigrate(&Rule{}, &Certificate{})
	if err != nil {
		return nil, err
	}

	return db, nil
}
