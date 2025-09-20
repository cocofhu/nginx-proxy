package db

import (
	"log"

	"gorm.io/gorm"
	_ "modernc.org/sqlite"
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

// MigrateCertificateTableV2 迁移证书表，添加续期相关字段
func MigrateCertificateTableV2(db *gorm.DB) error {
	// 检查是否需要添加新字段
	if !db.Migrator().HasColumn(&Certificate{}, "status") {
		if err := db.Migrator().AddColumn(&Certificate{}, "status"); err != nil {
			return err
		}
		// 为现有证书设置默认状态
		db.Model(&Certificate{}).Where("status = '' OR status IS NULL").Update("status", "active")
	}

	if !db.Migrator().HasColumn(&Certificate{}, "renewal_source_id") {
		if err := db.Migrator().AddColumn(&Certificate{}, "renewal_source_id"); err != nil {
			return err
		}
	}

	if !db.Migrator().HasColumn(&Certificate{}, "original_source_id") {
		if err := db.Migrator().AddColumn(&Certificate{}, "original_source_id"); err != nil {
			return err
		}
	}

	return nil
}

// MigrateCertificateTableV3 迁移证书表，确保所有字段都存在
func MigrateCertificateTableV3(db *gorm.DB) error {
	// 确保所有必要的字段都存在
	if !db.Migrator().HasColumn(&Certificate{}, "source") {
		if err := db.Migrator().AddColumn(&Certificate{}, "source"); err != nil {
			return err
		}
	}

	if !db.Migrator().HasColumn(&Certificate{}, "source_id") {
		if err := db.Migrator().AddColumn(&Certificate{}, "source_id"); err != nil {
			return err
		}
	}

	// 为现有的上传证书设置正确的状态
	db.Model(&Certificate{}).Where("source = 'upload' AND (status = '' OR status IS NULL)").Update("status", "active")

	return nil
}
