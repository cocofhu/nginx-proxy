package db

import (
	"log"

	"gorm.io/gorm"
)

// MigrateCertificateTable 迁移证书表，添加新字段
func MigrateCertificateTable(db *gorm.DB) error {
	// 检查是否需要添加新字段
	if !db.Migrator().HasColumn(&Certificate{}, "source") {
		if err := db.Migrator().AddColumn(&Certificate{}, "source"); err != nil {
			log.Printf("Failed to add source column: %v", err)
			return err
		}
		log.Println("Added source column to certificates table")
	}

	if !db.Migrator().HasColumn(&Certificate{}, "source_id") {
		if err := db.Migrator().AddColumn(&Certificate{}, "source_id"); err != nil {
			log.Printf("Failed to add source_id column: %v", err)
			return err
		}
		log.Println("Added source_id column to certificates table")
	}

	// 为现有记录设置默认值
	if err := db.Model(&Certificate{}).Where("source = ? OR source IS NULL", "").Update("source", "upload").Error; err != nil {
		log.Printf("Failed to update existing certificates source: %v", err)
		return err
	}

	return nil
}
