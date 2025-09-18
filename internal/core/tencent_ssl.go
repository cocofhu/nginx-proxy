package core

import (
	"archive/zip"
	"bytes"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common/errors"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common/profile"
	ssl "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/ssl/v20191205"
	"gorm.io/gorm"

	"nginx-proxy/internal/db"
)

// TencentSSLService 腾讯云SSL证书服务
type TencentSSLService struct {
	sslClient *ssl.Client
	db        *gorm.DB
	certDir   string
}

// ApplyCertificateRequest 申请证书请求
type ApplyCertificateRequest struct {
	Domain       string `json:"domain" binding:"required"`
	ValidateType string `json:"validate_type" binding:"required"` // DNS_AUTO, DNS, FILE_VALIDATION
	CertAlias    string `json:"cert_alias"`
}

// ApplyCertificateResponse 申请证书响应
type ApplyCertificateResponse struct {
	CertificateID string                 `json:"certificate_id"`
	Status        string                 `json:"status"`
	ValidateInfo  map[string]interface{} `json:"validate_info,omitempty"`
}

// CertificateStatusResponse 证书状态响应
type CertificateStatusResponse struct {
	CertificateID string `json:"certificate_id"`
	Status        string `json:"status"`
	Domain        string `json:"domain"`
	ExpiresAt     string `json:"expires_at,omitempty"`
}

// RenewCertificateResponse 续期证书响应
type RenewCertificateResponse struct {
	NewCertificateID string                 `json:"new_certificate_id"`
	Status           string                 `json:"status"`
	ValidateInfo     map[string]interface{} `json:"validate_info,omitempty"`
}

// NewTencentSSLService 创建腾讯云SSL服务实例
func NewTencentSSLService(config *TencentCloudConfig, database *gorm.DB, certDir string) *TencentSSLService {
	credential := common.NewCredential(config.SecretId, config.SecretKey)
	cpf := profile.NewClientProfile()
	cpf.HttpProfile.Endpoint = "ssl.tencentcloudapi.com"

	client, err := ssl.NewClient(credential, config.Region, cpf)
	if err != nil {
		log.Printf("Failed to create Tencent Cloud SSL client: %v", err)
		return nil
	}

	service := &TencentSSLService{
		sslClient: client,
		db:        database,
		certDir:   certDir,
	}

	return service
}

// ApplyCertificate 申请免费证书
func (s *TencentSSLService) ApplyCertificate(req *ApplyCertificateRequest) (*ApplyCertificateResponse, error) {
	log.Printf("Applying certificate for domain: %s", req.Domain)

	// 创建申请证书请求
	request := ssl.NewApplyCertificateRequest()
	request.DvAuthMethod = common.StringPtr(req.ValidateType)
	request.DomainName = common.StringPtr(req.Domain)

	if req.CertAlias != "" {
		request.Alias = common.StringPtr(req.CertAlias)
	}

	// 调用腾讯云API
	response, err := s.sslClient.ApplyCertificate(request)
	if err != nil {
		if sdkErr, ok := err.(*errors.TencentCloudSDKError); ok {
			return nil, fmt.Errorf("腾讯云API错误: %s - %s", sdkErr.Code, sdkErr.Message)
		}
		return nil, fmt.Errorf("申请证书失败: %w", err)
	}

	certID := *response.Response.CertificateId

	// 检查是否已存在相同的证书记录（避免重复创建）
	var existingCert db.Certificate
	err = s.db.Where("source_id = ? AND source = ?", certID, "tencent_cloud").First(&existingCert).Error
	if err == nil {
		// 如果已存在，直接返回
		log.Printf("Certificate record already exists: %s", certID)
	} else if err == gorm.ErrRecordNotFound {
		// 记录不存在，创建新的数据库记录（这是正常情况）
		log.Printf("Creating new certificate record: %s", certID)
		certificate := db.Certificate{
			ID:       certID,
			Name:     req.CertAlias,
			Domain:   req.Domain,
			Source:   "tencent_cloud",
			SourceID: certID,
			CertPath: "",       // 申请阶段暂时为空
			KeyPath:  "",       // 申请阶段暂时为空
			Status:   "active", // 默认状态为活跃
		}

		if err := s.db.Create(&certificate).Error; err != nil {
			return nil, fmt.Errorf("failed to save certificate record: %w", err)
		}
	} else {
		// 其他数据库错误
		return nil, fmt.Errorf("failed to check existing certificate: %w", err)
	}

	result := &ApplyCertificateResponse{
		CertificateID: certID,
		Status:        "申请中",
	}

	// 获取验证信息
	if req.ValidateType == "DNS" {
		// 查询DNS验证信息
		descRequest := ssl.NewDescribeCertificateDetailRequest()
		descRequest.CertificateId = common.StringPtr(certID)

		descResponse, err := s.sslClient.DescribeCertificateDetail(descRequest)
		if err == nil && descResponse.Response.DvAuthDetail != nil {
			result.ValidateInfo = map[string]interface{}{
				"type":   "DNS",
				"record": *descResponse.Response.DvAuthDetail.DvAuthKey,
				"value":  *descResponse.Response.DvAuthDetail.DvAuthValue,
			}
		}
	}

	return result, nil
}

// CheckCertificateStatus 检查证书状态
func (s *TencentSSLService) CheckCertificateStatus(certificateID string) (*CertificateStatusResponse, error) {
	var certificate db.Certificate
	if err := s.db.First(&certificate, "source_id = ? AND source = ?", certificateID, "tencent_cloud").Error; err != nil {
		return nil, fmt.Errorf("certificate not found: %w", err)
	}

	// 调用腾讯云API查询证书详情
	request := ssl.NewDescribeCertificateDetailRequest()
	request.CertificateId = common.StringPtr(certificateID)

	response, err := s.sslClient.DescribeCertificateDetail(request)
	if err != nil {
		if sdkErr, ok := err.(*errors.TencentCloudSDKError); ok {
			return nil, fmt.Errorf("腾讯云API错误: %s - %s", sdkErr.Code, sdkErr.Message)
		}
		return nil, fmt.Errorf("查询证书状态失败: %w", err)
	}

	status := *response.Response.Status
	statusMap := map[uint64]string{
		0:  "审核中",
		1:  "已通过",
		2:  "审核失败",
		3:  "已过期",
		4:  "DNS记录添加中",
		5:  "企业证书，待提交",
		6:  "订单取消中",
		7:  "已取消",
		8:  "已提交资料，待上传确认函",
		9:  "证书吊销中",
		10: "已吊销",
		11: "重颁发中",
		12: "待上传吊销确认函",
	}

	statusText := statusMap[status]
	if statusText == "" {
		statusText = "未知状态"
	}

	result := &CertificateStatusResponse{
		CertificateID: certificateID,
		Status:        statusText,
		Domain:        certificate.Domain,
	}

	// 更新本地数据库中的证书状态信息
	needUpdate := false

	// 如果证书已通过，添加过期时间并更新数据库
	if status == 1 && response.Response.CertEndTime != nil {
		result.ExpiresAt = *response.Response.CertEndTime

		// 解析过期时间并更新到数据库
		if expTime, parseErr := time.Parse("2006-01-02 15:04:05", *response.Response.CertEndTime); parseErr == nil {
			if certificate.ExpiresAt.IsZero() || !certificate.ExpiresAt.Equal(expTime) {
				certificate.ExpiresAt = expTime
				needUpdate = true
			}
		}

		// 如果证书已通过但本地没有证书文件，自动下载证书
		if certificate.CertPath == "" || certificate.KeyPath == "" {
			log.Printf("Certificate %s is approved but files not found locally, downloading...", certificateID)
			if downloadErr := s.DownloadCertificate(certificateID); downloadErr != nil {
				log.Printf("Warning: Failed to auto-download certificate %s: %v", certificateID, downloadErr)
			} else {
				log.Printf("Successfully auto-downloaded certificate %s", certificateID)
				// 重新读取更新后的证书信息
				s.db.First(&certificate, "source_id = ? AND source = ?", certificateID, "tencent_cloud")
			}
		}

		// 检查是否有续期的新证书ID，如果有且新证书已通过，则处理续期完成逻辑
		if certificate.RenewalSourceID != "" && certificate.Status == "renewing" {
			if err := s.handleRenewalCompletion(certificateID, certificate.RenewalSourceID); err != nil {
				log.Printf("Warning: Failed to handle renewal completion: %v", err)
			}
		}
	}

	// 更新证书别名（如果有变化）
	if response.Response.Alias != nil && *response.Response.Alias != certificate.Name {
		certificate.Name = *response.Response.Alias
		needUpdate = true
	}

	// 如果有信息需要更新，保存到数据库
	if needUpdate {
		if err := s.db.Save(&certificate).Error; err != nil {
			log.Printf("Warning: Failed to update certificate info: %v", err)
		} else {
			log.Printf("Updated certificate info for: %s", certificateID)
		}
	}

	return result, nil
}

// DownloadCertificate 下载证书
func (s *TencentSSLService) DownloadCertificate(certificateID string) error {
	var certificate db.Certificate
	if err := s.db.First(&certificate, "source_id = ? AND source = ?", certificateID, "tencent_cloud").Error; err != nil {
		return fmt.Errorf("certificate not found: %w", err)
	}
	return s.downloadCertificateFromAPI(certificateID)
}

// RenewTencentCertificate 续期腾讯云证书（在原记录上更新，不创建新记录）
func (s *TencentSSLService) RenewTencentCertificate(oldCertificateID string) (*RenewCertificateResponse, error) {
	// 获取旧证书信息
	var certificate db.Certificate
	if err := s.db.First(&certificate, "source_id = ? AND source = ?", oldCertificateID, "tencent_cloud").Error; err != nil {
		return nil, fmt.Errorf("证书不存在: %w", err)
	}

	// 检查是否已经在续期中
	if certificate.Status == "renewing" {
		return nil, fmt.Errorf("证书已在续期中，请稍后再试")
	}

	log.Printf("Renewing certificate for domain: %s (cert: %s)", certificate.Domain, oldCertificateID)

	// 更新证书状态为"续期中"
	certificate.Status = "renewing"
	if err := s.db.Save(&certificate).Error; err != nil {
		return nil, fmt.Errorf("更新证书状态失败: %w", err)
	}

	// 申请新证书（腾讯云API）
	request := ssl.NewApplyCertificateRequest()
	request.DvAuthMethod = common.StringPtr("DNS_AUTO") // 默认使用自动DNS验证
	request.DomainName = common.StringPtr(certificate.Domain)
	request.Alias = common.StringPtr(certificate.Name + "_renewed_" + time.Now().Format("20060102"))

	// 调用腾讯云API申请新证书
	response, err := s.sslClient.ApplyCertificate(request)
	if err != nil {
		// 如果申请失败，恢复证书状态
		certificate.Status = "active"
		s.db.Save(&certificate)

		if sdkErr, ok := err.(*errors.TencentCloudSDKError); ok {
			return nil, fmt.Errorf("腾讯云API错误: %s - %s", sdkErr.Code, sdkErr.Message)
		}
		return nil, fmt.Errorf("申请新证书失败: %w", err)
	}

	newCertID := *response.Response.CertificateId

	// 记录新证书ID到续期字段
	certificate.RenewalSourceID = newCertID
	if err := s.db.Save(&certificate).Error; err != nil {
		log.Printf("Warning: Failed to update renewal certificate ID: %v", err)
	}

	// 构建响应
	renewResponse := &RenewCertificateResponse{
		NewCertificateID: newCertID,
		Status:           "申请中",
	}

	// 获取验证信息（如果需要手动验证）
	descRequest := ssl.NewDescribeCertificateDetailRequest()
	descRequest.CertificateId = common.StringPtr(newCertID)

	if descResponse, err := s.sslClient.DescribeCertificateDetail(descRequest); err == nil && descResponse.Response.DvAuthDetail != nil {
		renewResponse.ValidateInfo = map[string]interface{}{
			"type":   "DNS",
			"record": *descResponse.Response.DvAuthDetail.DvAuthKey,
			"value":  *descResponse.Response.DvAuthDetail.DvAuthValue,
		}
	}

	log.Printf("Certificate renewal initiated: cert=%s, new_cert_id=%s", oldCertificateID, newCertID)
	return renewResponse, nil
}

// handleRenewalCompletion 处理续期完成逻辑
func (s *TencentSSLService) handleRenewalCompletion(originalCertID, newCertID string) error {
	log.Printf("Handling renewal completion: original=%s, new=%s", originalCertID, newCertID)

	// 获取原始证书信息
	var originalCert db.Certificate
	if err := s.db.First(&originalCert, "source_id = ? AND source = ?", originalCertID, "tencent_cloud").Error; err != nil {
		return fmt.Errorf("failed to find original certificate: %w", err)
	}

	// 直接通过腾讯云API检查新证书状态，而不是查找数据库记录
	request := ssl.NewDescribeCertificateDetailRequest()
	request.CertificateId = common.StringPtr(newCertID)

	response, err := s.sslClient.DescribeCertificateDetail(request)
	if err != nil {
		if sdkErr, ok := err.(*errors.TencentCloudSDKError); ok {
			return fmt.Errorf("腾讯云API错误: %s - %s", sdkErr.Code, sdkErr.Message)
		}
		return fmt.Errorf("查询新证书状态失败: %w", err)
	}

	status := *response.Response.Status

	// 只有当新证书已通过时才进行切换
	if status != 1 { // 1表示已通过
		statusMap := map[uint64]string{
			0:  "审核中",
			1:  "已通过",
			2:  "审核失败",
			3:  "已过期",
			4:  "DNS记录添加中",
			5:  "企业证书，待提交",
			6:  "订单取消中",
			7:  "已取消",
			8:  "已提交资料，待上传确认函",
			9:  "证书吊销中",
			10: "已吊销",
			11: "重颁发中",
			12: "待上传吊销确认函",
		}
		statusText := statusMap[status]
		if statusText == "" {
			statusText = "未知状态"
		}
		log.Printf("New certificate %s not ready yet, status: %s", newCertID, statusText)
		return nil
	}

	// 下载新证书
	if err := s.downloadCertificateFromAPI(newCertID); err != nil {
		log.Printf("Warning: Failed to download new certificate %s: %v", newCertID, err)
		return err
	}

	// 保存老证书文件路径，用于后续删除
	oldCertPath := originalCert.CertPath
	oldKeyPath := originalCert.KeyPath

	// 生成新证书的文件路径
	newCertPath := filepath.Join(s.certDir, newCertID+".crt")
	newKeyPath := filepath.Join(s.certDir, newCertID+".key")

	// 更新原始证书记录，使用新证书的内容和路径，但保持原有的ID结构
	// 不更改SourceID，保持证书记录的连续性
	originalCert.CertPath = newCertPath
	originalCert.KeyPath = newKeyPath
	originalCert.Status = "active"
	originalCert.RenewalSourceID = "" // 清除续期标记

	// 记录新证书ID用于追踪，但不作为主要标识
	log.Printf("Certificate %s renewed with new Tencent Cloud cert %s", originalCertID, newCertID)

	// 解析新证书的过期时间
	if response.Response.CertEndTime != nil {
		if expTime, parseErr := time.Parse("2006-01-02 15:04:05", *response.Response.CertEndTime); parseErr == nil {
			originalCert.ExpiresAt = expTime
		}
	}

	// 保存更新后的证书记录
	if err := s.db.Save(&originalCert).Error; err != nil {
		return fmt.Errorf("failed to update certificate record: %w", err)
	}

	// 删除老证书文件
	if oldCertPath != "" && oldCertPath != newCertPath {
		if err := os.Remove(oldCertPath); err != nil {
			log.Printf("Warning: Failed to delete old certificate file %s: %v", oldCertPath, err)
		} else {
			log.Printf("Successfully deleted old certificate file: %s", oldCertPath)
		}
	}

	if oldKeyPath != "" && oldKeyPath != newKeyPath {
		if err := os.Remove(oldKeyPath); err != nil {
			log.Printf("Warning: Failed to delete old key file %s: %v", oldKeyPath, err)
		} else {
			log.Printf("Successfully deleted old key file: %s", oldKeyPath)
		}
	}

	// 注意：不删除腾讯云端的老证书，因为我们要保持证书记录的连续性
	// 老证书会在腾讯云端自然过期，这样可以避免证书ID失效的问题
	log.Printf("Keeping old certificate %s in Tencent Cloud for reference, it will expire naturally", originalCertID)

	// 更新使用此证书的nginx配置（如果有的话）
	if err := s.updateNginxConfigForRenewal(originalCert, originalCert); err != nil {
		log.Printf("Warning: Failed to update nginx config: %v", err)
	}

	log.Printf("Renewal completion handled successfully: %s (now using new cert %s)", originalCertID, newCertID)
	return nil
}

// updateNginxConfigForRenewal 更新nginx配置中的证书路径
func (s *TencentSSLService) updateNginxConfigForRenewal(originalCert, newCert db.Certificate) error {
	// 查找使用原始证书的所有规则
	var rules []db.Rule
	if err := s.db.Where("ssl_cert = ? OR ssl_key = ?", originalCert.CertPath, originalCert.KeyPath).Find(&rules).Error; err != nil {
		return fmt.Errorf("failed to find rules using original certificate: %w", err)
	}

	if len(rules) == 0 {
		log.Printf("No nginx rules found using certificate %s", originalCert.SourceID)
		return nil
	}

	// 更新每个规则的证书路径
	for _, rule := range rules {
		needUpdate := false

		if rule.SSLCert == originalCert.CertPath {
			rule.SSLCert = newCert.CertPath
			needUpdate = true
		}

		if rule.SSLKey == originalCert.KeyPath {
			rule.SSLKey = newCert.KeyPath
			needUpdate = true
		}

		if needUpdate {
			if err := s.db.Save(&rule).Error; err != nil {
				log.Printf("Warning: Failed to update rule %s: %v", rule.ID, err)
				continue
			}
			log.Printf("Updated nginx rule %s to use new certificate", rule.ID)
		}
	}

	log.Printf("Updated %d nginx rules to use new certificate %s", len(rules), newCert.SourceID)
	return nil
}

// UpdateCertificateName 更新证书名称
func (s *TencentSSLService) UpdateCertificateName(certificateID, newName string) error {
	var certificate db.Certificate
	if err := s.db.First(&certificate, "source_id = ? AND source = ?", certificateID, "tencent_cloud").Error; err != nil {
		return fmt.Errorf("certificate not found: %w", err)
	}

	// 更新证书名称
	certificate.Name = newName
	if err := s.db.Save(&certificate).Error; err != nil {
		return fmt.Errorf("failed to update certificate name: %w", err)
	}

	log.Printf("Certificate name updated: %s -> %s", certificateID, newName)
	return nil
}

// revokeTencentCloudCertificate 删除腾讯云端的证书
func (s *TencentSSLService) revokeTencentCloudCertificate(certificateID string) error {
	// 创建删除证书请求
	request := ssl.NewRevokeCertificateRequest()
	request.CertificateId = common.StringPtr(certificateID)
	request.Reason = common.StringPtr("nginx proxy")
	// 调用腾讯云API删除证书
	_, err := s.sslClient.RevokeCertificate(request)
	if err != nil {
		if sdkErr, ok := err.(*errors.TencentCloudSDKError); ok {
			// 如果证书不存在或已被删除，不视为错误
			if sdkErr.Code == "InvalidParameter.CertificateNotFound" ||
				sdkErr.Code == "InvalidParameter.CertificateStatusInvalid" {
				log.Printf("Certificate %s not found or already deleted in Tencent Cloud", certificateID)
				return nil
			}
			return fmt.Errorf("腾讯云API错误: %s - %s", sdkErr.Code, sdkErr.Message)
		}
		return fmt.Errorf("删除腾讯云证书失败: %w", err)
	}

	log.Printf("Successfully deleted certificate %s from Tencent Cloud", certificateID)
	return nil
}

// DeleteTencentCertificate 删除腾讯云证书（同时删除腾讯云端证书）
func (s *TencentSSLService) DeleteTencentCertificate(certificateID string) error {
	var certificate db.Certificate
	if err := s.db.First(&certificate, "source_id = ? AND source = ?", certificateID, "tencent_cloud").Error; err != nil {
		return fmt.Errorf("certificate not found: %w", err)
	}

	// 先删除腾讯云端的证书
	if err := s.revokeTencentCloudCertificate(certificateID); err != nil {
		log.Printf("Warning: Failed to delete certificate from Tencent Cloud: %v", err)
		// 不阻止本地删除，继续执行
	}

	// 删除本地文件
	if certificate.CertPath != "" {
		if err := os.Remove(certificate.CertPath); err != nil && !os.IsNotExist(err) {
			log.Printf("Warning: Failed to delete cert file: %v", err)
		}
	}
	if certificate.KeyPath != "" {
		if err := os.Remove(certificate.KeyPath); err != nil && !os.IsNotExist(err) {
			log.Printf("Warning: Failed to delete key file: %v", err)
		}
	}

	// 从数据库删除
	if err := s.db.Delete(&certificate).Error; err != nil {
		return fmt.Errorf("failed to delete certificate: %w", err)
	}

	log.Printf("Certificate record deleted: %s", certificateID)
	return nil
}

// downloadCertificateFromAPI 直接从腾讯云API下载证书文件
func (s *TencentSSLService) downloadCertificateFromAPI(certificateID string) error {
	// 确保证书目录存在
	if err := os.MkdirAll(s.certDir, 0755); err != nil {
		return fmt.Errorf("failed to create cert directory: %w", err)
	}

	// 调用腾讯云API下载证书
	request := ssl.NewDownloadCertificateRequest()
	request.CertificateId = common.StringPtr(certificateID)

	response, err := s.sslClient.DownloadCertificate(request)
	if err != nil {
		if sdkErr, ok := err.(*errors.TencentCloudSDKError); ok {
			return fmt.Errorf("腾讯云API错误: %s - %s", sdkErr.Code, sdkErr.Message)
		}
		return fmt.Errorf("下载证书失败: %w", err)
	}

	if response.Response.Content == nil {
		return fmt.Errorf("证书内容为空")
	}

	// 生成证书文件路径
	certPath := filepath.Join(s.certDir, certificateID+".crt")
	keyPath := filepath.Join(s.certDir, certificateID+".key")

	// 从ZIP格式的证书包中提取证书和私钥
	if err := s.extractCertificateFromZip(*response.Response.Content, certPath, keyPath); err != nil {
		return fmt.Errorf("提取证书文件失败: %w", err)
	}

	log.Printf("Certificate downloaded successfully: %s", certificateID)
	return nil
}

// extractCertificateFromZip 从ZIP格式的证书包中提取证书和私钥文件
func (s *TencentSSLService) extractCertificateFromZip(zipContent, certPath, keyPath string) error {
	// 解码base64内容
	zipData, err := base64.StdEncoding.DecodeString(zipContent)
	if err != nil {
		return fmt.Errorf("failed to decode base64 zip content: %w", err)
	}

	// 创建ZIP读取器
	zipReader, err := zip.NewReader(bytes.NewReader(zipData), int64(len(zipData)))
	if err != nil {
		return fmt.Errorf("failed to create zip reader: %w", err)
	}

	var certContent, keyContent string

	// 遍历ZIP文件中的所有文件
	for _, file := range zipReader.File {
		rc, err := file.Open()
		if err != nil {
			log.Printf("Warning: Failed to open file %s in zip: %v", file.Name, err)
			continue
		}

		content, err := io.ReadAll(rc)
		if closeErr := rc.Close(); closeErr != nil {
			log.Printf("Warning: Failed to close file %s: %v", file.Name, closeErr)
		}
		if err != nil {
			log.Printf("Warning: Failed to read file %s: %v", file.Name, err)
			continue
		}

		fileName := strings.ToLower(file.Name)

		// 识别证书文件（.crt, .pem, .cer）
		if strings.HasSuffix(fileName, ".crt") ||
			strings.HasSuffix(fileName, ".pem") ||
			strings.HasSuffix(fileName, ".cer") {
			certContent = string(content)
		}

		// 识别私钥文件（.key）
		if strings.HasSuffix(fileName, ".key") {
			keyContent = string(content)
		}
	}

	// 检查是否找到了证书和私钥
	if certContent == "" {
		return fmt.Errorf("certificate file not found in zip")
	}
	if keyContent == "" {
		return fmt.Errorf("private key file not found in zip")
	}

	// 保存证书文件
	if err := os.WriteFile(certPath, []byte(certContent), 0644); err != nil {
		return fmt.Errorf("failed to save certificate file: %w", err)
	}

	// 保存私钥文件
	if err := os.WriteFile(keyPath, []byte(keyContent), 0600); err != nil {
		return fmt.Errorf("failed to save key file: %w", err)
	}

	return nil
}

// parseCertificateExpiry 解析证书过期时间
func (s *TencentSSLService) parseCertificateExpiry(certContent string) (time.Time, error) {
	block, _ := pem.Decode([]byte(certContent))
	if block == nil {
		return time.Time{}, fmt.Errorf("failed to parse certificate PEM")
	}

	cert, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		return time.Time{}, fmt.Errorf("failed to parse certificate: %w", err)
	}

	return cert.NotAfter, nil
}

// parseCertificateExpiryFromFile 从证书文件解析过期时间
func (s *TencentSSLService) parseCertificateExpiryFromFile(certPath string) (time.Time, error) {
	certData, err := os.ReadFile(certPath)
	if err != nil {
		return time.Time{}, fmt.Errorf("failed to read certificate file: %w", err)
	}

	return s.parseCertificateExpiry(string(certData))
}

// TencentCertificateInfo 腾讯云证书信息
type TencentCertificateInfo struct {
	CertificateID string `json:"certificate_id"`
	Domain        string `json:"domain"`
	Alias         string `json:"alias"`
	Status        string `json:"status"`
	ExpiresAt     string `json:"expires_at"`
}

// GetAllTencentCertificates 获取腾讯云所有证书列表
func (s *TencentSSLService) GetAllTencentCertificates() ([]TencentCertificateInfo, error) {
	log.Printf("Fetching all certificates from Tencent Cloud...")

	// 创建查询证书列表请求
	request := ssl.NewDescribeCertificatesRequest()

	// 设置分页参数，一次获取所有证书
	request.Limit = common.Uint64Ptr(100) // 腾讯云单次最多返回100个
	request.Offset = common.Uint64Ptr(0)

	var allCertificates []TencentCertificateInfo

	for {
		// 调用腾讯云API
		response, err := s.sslClient.DescribeCertificates(request)
		if err != nil {
			if sdkErr, ok := err.(*errors.TencentCloudSDKError); ok {
				return nil, fmt.Errorf("腾讯云API错误: %s - %s", sdkErr.Code, sdkErr.Message)
			}
			return nil, fmt.Errorf("获取证书列表失败: %w", err)
		}

		if response.Response.Certificates == nil || len(response.Response.Certificates) == 0 {
			break
		}

		// 处理当前页的证书
		for _, cert := range response.Response.Certificates {
			if cert.CertificateId == nil {
				continue
			}

			certInfo := TencentCertificateInfo{
				CertificateID: *cert.CertificateId,
			}

			if cert.Domain != nil {
				certInfo.Domain = *cert.Domain
			}

			if cert.Alias != nil {
				certInfo.Alias = *cert.Alias
			}

			if cert.Status != nil {
				statusMap := map[uint64]string{
					0:  "审核中",
					1:  "已通过",
					2:  "审核失败",
					3:  "已过期",
					4:  "DNS记录添加中",
					5:  "企业证书，待提交",
					6:  "订单取消中",
					7:  "已取消",
					8:  "已提交资料，待上传确认函",
					9:  "证书吊销中",
					10: "已吊销",
					11: "重颁发中",
					12: "待上传吊销确认函",
				}
				if statusText, exists := statusMap[*cert.Status]; exists {
					certInfo.Status = statusText
				} else {
					certInfo.Status = "未知状态"
				}
			}

			if cert.CertEndTime != nil {
				certInfo.ExpiresAt = *cert.CertEndTime
			}

			allCertificates = append(allCertificates, certInfo)
		}

		// 检查是否还有更多证书
		if response.Response.TotalCount == nil ||
			uint64(len(allCertificates)) >= *response.Response.TotalCount {
			break
		}

		// 更新偏移量获取下一页
		*request.Offset += *request.Limit
	}

	log.Printf("Found %d certificates in Tencent Cloud", len(allCertificates))
	return allCertificates, nil
}

// RevokeTencentCertificateByID 根据证书ID吊销腾讯云证书
func (s *TencentSSLService) RevokeTencentCertificateByID(certificateID string) error {
	return s.revokeTencentCloudCertificate(certificateID)
}
