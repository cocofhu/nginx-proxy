package core

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/cloudflare/cloudflare-go"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common/profile"
	dnspod "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/dnspod/v20210323"
	ssl "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/ssl/v20191205"
	"gorm.io/gorm"

	"nginx-proxy/internal/db"
)

// CleanupService 证书验证记录清理服务
type CleanupService struct {
	db              *gorm.DB
	cfAPI           *cloudflare.API
	sslClient       *ssl.Client
	dpClient        *dnspod.Client
	cleanupInterval time.Duration
	ctx             context.Context
	cancel          context.CancelFunc
}

// CleanupConfig 清理服务配置
type CleanupConfig struct {
	// Cloudflare配置
	CloudflareAPIToken string

	// 腾讯云配置
	TencentSecretId  string
	TencentSecretKey string
	TencentRegion    string

	// 清理间隔，默认1小时
	CleanupInterval time.Duration
}

// NewCleanupService 创建清理服务实例
func NewCleanupService(database *gorm.DB, config CleanupConfig) (*CleanupService, error) {
	ctx, cancel := context.WithCancel(context.Background())
	service := &CleanupService{
		db:              database,
		cleanupInterval: config.CleanupInterval,
		ctx:             ctx,
		cancel:          cancel,
	}
	if service.cleanupInterval == 0 {
		service.cleanupInterval = time.Minute // 默认每分钟清理一次
	}
	// 初始化Cloudflare API
	if config.CloudflareAPIToken != "" {
		cfAPI, err := cloudflare.NewWithAPIToken(config.CloudflareAPIToken)
		if err != nil {
			return nil, fmt.Errorf("failed to create Cloudflare API client: %w", err)
		}
		service.cfAPI = cfAPI
	}
	// 初始化腾讯云SSL客户端
	if config.TencentSecretId != "" && config.TencentSecretKey != "" {
		credential := common.NewCredential(config.TencentSecretId, config.TencentSecretKey)
		cpf := profile.NewClientProfile()
		cpf.HttpProfile.Endpoint = "ssl.tencentcloudapi.com"
		client, err := ssl.NewClient(credential, config.TencentRegion, cpf)
		if err != nil {
			return nil, fmt.Errorf("failed to create Tencent Cloud SSL client: %w", err)
		}
		service.sslClient = client
	}
	// 初始化腾讯云DNSPod客户端
	if config.TencentSecretId != "" && config.TencentSecretKey != "" {
		credential := common.NewCredential(config.TencentSecretId, config.TencentSecretKey)
		cpf := profile.NewClientProfile()
		cpf.HttpProfile.Endpoint = "dnspod.tencentcloudapi.com"
		client, err := dnspod.NewClient(credential, config.TencentRegion, cpf)
		if err != nil {
			return nil, fmt.Errorf("failed to create Tencent Cloud DNSPod client: %w", err)
		}
		service.dpClient = client
	}
	return service, nil
}

// Start 启动清理服务
func (s *CleanupService) Start() {
	log.Printf("Start cleanup service interval: %v", s.cleanupInterval)
	// 立即执行一次清理
	go s.cleanup()
	// 启动定时器
	ticker := time.NewTicker(s.cleanupInterval)
	go func() {
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				s.cleanup()
			case <-s.ctx.Done():
				log.Println("cleanup service stopped")
				return
			}
		}
	}()
}

// Stop 停止清理服务
func (s *CleanupService) Stop() {
	log.Println("stopping cleanup services...")
	s.cancel()
}

// cleanup 执行清理操作
func (s *CleanupService) cleanup() {
	// 获取所有待清理的记录
	var records []db.AuthRecord
	if err := s.db.Find(&records).Error; err != nil {
		log.Printf("find dns records error: %v", err)
		return
	}
	if len(records) == 0 {
		log.Println("no records need to cleanup skipped")
		return
	}
	log.Printf("there are %d record(s) need to clean", len(records))
	for _, record := range records {
		if err := s.cleanupRecord(record); err != nil {
			continue
		}

	}
}

// cleanupRecord 清理单个记录
func (s *CleanupService) cleanupRecord(record db.AuthRecord) error {
	// 检查证书状态
	if record.CertificateId != "" {
		// 下游已经记录错误 此处不用处理
		completed, _ := s.checkCertificateStatus(record)
		if !completed {
			log.Printf("waiting certificate %s ，skipped", record.CertificateId)
			return nil
		}
	}

	// 根据来源清理DNS记录, 下游已经处理错误
	switch record.Source {
	case "cloudflare":
		_ = s.cleanupCloudflareRecord(record)
	case "tencent_cloud":
		_ = s.cleanupTencentRecord(record)
	default:
		log.Printf("unsupproted source: %s", record.Source)
	}
	// 删除数据库记录
	if err := s.db.Delete(&record).Error; err != nil {
		return fmt.Errorf("database error: %w", err)
	}
	log.Printf("records: %s %s is deleted", record.Key, record.Value)
	return nil
}

// checkCertificateStatus 检查腾讯云证书状态
func (s *CleanupService) checkCertificateStatus(record db.AuthRecord) (bool, error) {
	if s.sslClient == nil {
		log.Println("there is not tencent aksk configuration, skipped.")
		return false, nil
	}
	request := ssl.NewDescribeCertificateDetailRequest()
	request.CertificateId = common.StringPtr(record.CertificateId)
	response, err := s.sslClient.DescribeCertificateDetail(request)
	if err != nil {
		return false, fmt.Errorf("failed to query certificate from tencent cloud: %w", err)
	}
	if response.Response.Status == nil {
		return false, fmt.Errorf("failed to get status of certificate from tencent cloud")
	}
	// 10: 吊销中
	// 1 : 颁发成功
	status := *response.Response.Status
	return status == 10 || status == 1, nil // 状态>=3表示验证已完成
}

// cleanupCloudflareRecord 清理Cloudflare DNS记录
func (s *CleanupService) cleanupCloudflareRecord(record db.AuthRecord) error {
	if s.cfAPI == nil {
		return fmt.Errorf("cloudflare not config")
	}
	ctx := context.Background()
	// 获取Zone ID
	zoneID, err := s.cfAPI.ZoneIDByName(record.Domain)
	if err != nil {
		return fmt.Errorf("query zoneId failed from cloudflare : %w", err)
	}
	// 查找DNS记录
	records, _, err := s.cfAPI.ListDNSRecords(ctx, cloudflare.ZoneIdentifier(zoneID), cloudflare.ListDNSRecordsParams{
		Type: record.Type,
		Name: record.Key + "." + record.Domain,
	})
	if err != nil {
		return fmt.Errorf("query records of dns failed from cloudflare : %w", err)
	}
	// 删除匹配的记录
	for _, dnsRecord := range records {
		if dnsRecord.Content == record.Value {
			err := s.cfAPI.DeleteDNSRecord(ctx, cloudflare.ZoneIdentifier(zoneID), dnsRecord.ID)
			if err != nil {
				return fmt.Errorf("failed to delete recored : %v", err)
			}
			break
		}
	}
	return nil
}

// cleanupTencentRecord 清理腾讯云DNS记录
func (s *CleanupService) cleanupTencentRecord(record db.AuthRecord) error {
	if s.dpClient == nil {
		return fmt.Errorf("tencent aksk not config")
	}

	// 首先查询DNS记录列表，找到要删除的记录ID
	listRequest := dnspod.NewDescribeRecordListRequest()
	listRequest.Domain = common.StringPtr(record.Domain)
	listRequest.Limit = common.Uint64Ptr(3000)
	listRequest.RecordType = common.StringPtr(record.Type)
	listResponse, err := s.dpClient.DescribeRecordList(listRequest)

	if err != nil {
		return fmt.Errorf("query dnspod records error: %v", err)
	}

	// 查找匹配的记录
	var recordID *uint64
	for _, dnsRecord := range listResponse.Response.RecordList {
		if dnsRecord.Value != nil && *dnsRecord.Value == record.Value &&
			dnsRecord.Name != nil && *dnsRecord.Name == record.Key {
			recordID = dnsRecord.RecordId
			break
		}
	}
	if recordID == nil {
		log.Printf("dns record not found, maybe deleted: %s", record.Key)
		return nil
	}
	// 删除DNS记录
	deleteRequest := dnspod.NewDeleteRecordRequest()
	deleteRequest.Domain = common.StringPtr(record.Domain)
	deleteRequest.RecordId = recordID
	_, err = s.dpClient.DeleteRecord(deleteRequest)
	if err != nil {
		return fmt.Errorf("failed to delete recored : %v", err)
	}
	return nil
}
