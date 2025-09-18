package api

import (
	"crypto/x509"
	"encoding/pem"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"

	"nginx-proxy/internal/db"
)

// GetCertificates 获取所有证书
func (h *Handler) GetCertificates(c *gin.Context) {
	var certificates []db.Certificate
	if err := h.db.Find(&certificates).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"certificates": certificates})
}

// GetCertificate 获取单个证书
func (h *Handler) GetCertificate(c *gin.Context) {
	id := c.Param("id")

	var certificate db.Certificate
	if err := h.db.First(&certificate, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "Certificate not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, certificate)
}

// UploadCertificateNew 上传新证书（替换原有的UploadCertificate方法）
func (h *Handler) UploadCertificateNew(c *gin.Context) {
	// 确保证书目录存在
	if err := os.MkdirAll(h.certDir, 0755); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create cert directory"})
		return
	}

	// 获取上传的文件
	form, err := c.MultipartForm()
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	certFiles := form.File["cert"]
	keyFiles := form.File["key"]

	if len(certFiles) == 0 || len(keyFiles) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Both cert and key files are required"})
		return
	}

	certFile := certFiles[0]
	keyFile := keyFiles[0]

	// 生成唯一的文件名
	certID := uuid.New().String()
	certPath := filepath.Join(h.certDir, fmt.Sprintf("%s.crt", certID))
	keyPath := filepath.Join(h.certDir, fmt.Sprintf("%s.key", certID))

	// 保存证书文件
	if err := c.SaveUploadedFile(certFile, certPath); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save cert file"})
		return
	}

	// 保存密钥文件
	if err := c.SaveUploadedFile(keyFile, keyPath); err != nil {
		// 如果密钥保存失败，删除已保存的证书文件
		if removeErr := os.Remove(certPath); removeErr != nil {
			log.Printf("Warning: Failed to cleanup cert file after key save failure: %v", removeErr)
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save key file"})
		return
	}

	// 解析证书信息
	certInfo, err := parseCertificateInfo(certPath)
	if err != nil {
		// 如果解析失败，删除已保存的文件
		if removeErr := os.Remove(certPath); removeErr != nil {
			log.Printf("Warning: Failed to cleanup cert file after parse failure: %v", removeErr)
		}
		if removeErr := os.Remove(keyPath); removeErr != nil {
			log.Printf("Warning: Failed to cleanup key file after parse failure: %v", removeErr)
		}
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid certificate file: " + err.Error()})
		return
	}

	// 获取证书名称
	certName := strings.TrimSuffix(certFile.Filename, filepath.Ext(certFile.Filename))
	if nameValues := form.Value["name"]; len(nameValues) > 0 && nameValues[0] != "" {
		certName = nameValues[0]
	}

	// 创建证书记录
	certificate := db.Certificate{
		ID:        certID,
		Name:      certName,
		Domain:    certInfo.Domain,
		CertPath:  certPath,
		KeyPath:   keyPath,
		ExpiresAt: certInfo.ExpiresAt,
		Source:    "upload",
		Status:    "active", // 设置默认状态为活跃
	}

	// 保存到数据库
	if err := h.db.Create(&certificate).Error; err != nil {
		// 数据库保存失败，删除文件
		if removeErr := os.Remove(certPath); removeErr != nil {
			log.Printf("Warning: Failed to cleanup cert file after database save failure: %v", removeErr)
		}
		if removeErr := os.Remove(keyPath); removeErr != nil {
			log.Printf("Warning: Failed to cleanup key file after database save failure: %v", removeErr)
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"certificate": certificate,
		"message":     "Certificate uploaded successfully",
	})
}

// DeleteCertificate 删除证书
func (h *Handler) DeleteCertificate(c *gin.Context) {
	id := c.Param("id")

	var certificate db.Certificate
	if err := h.db.First(&certificate, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "Certificate not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// 检查是否有规则在使用这个证书
	var count int64
	if err := h.db.Model(&db.Rule{}).Where("ssl_cert = ? OR ssl_key = ?", certificate.CertPath, certificate.KeyPath).Count(&count).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to check certificate usage"})
		return
	}
	if count > 0 {
		c.JSON(http.StatusConflict, gin.H{"error": "Certificate is being used by existing rules"})
		return
	}

	// 删除文件
	if err := os.Remove(certificate.CertPath); err != nil && !os.IsNotExist(err) {
		log.Printf("Warning: Failed to delete cert file: %v", err)
	}
	if err := os.Remove(certificate.KeyPath); err != nil && !os.IsNotExist(err) {
		log.Printf("Warning: Failed to delete key file: %v", err)
	}

	// 从数据库删除
	if err := h.db.Delete(&certificate).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Certificate deleted successfully"})
}

// CertificateInfo 证书信息结构
type CertificateInfo struct {
	Domain    string
	ExpiresAt time.Time
}

// parseCertificateInfo 解析证书信息
func parseCertificateInfo(certPath string) (*CertificateInfo, error) {
	certData, err := os.ReadFile(certPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read certificate file: %w", err)
	}

	block, rest := pem.Decode(certData)
	if block == nil {
		return nil, fmt.Errorf("failed to parse certificate PEM")
	}
	// 检查是否有剩余数据（可能包含多个证书）
	if len(rest) > 0 {
		// 记录警告但不阻止处理
		fmt.Printf("Warning: Certificate file contains additional data after first certificate\n")
	}

	cert, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("failed to parse certificate: %w", err)
	}

	domain := ""
	if len(cert.DNSNames) > 0 {
		domain = cert.DNSNames[0]
	} else if cert.Subject.CommonName != "" {
		domain = cert.Subject.CommonName
	}

	return &CertificateInfo{
		Domain:    domain,
		ExpiresAt: cert.NotAfter,
	}, nil
}
