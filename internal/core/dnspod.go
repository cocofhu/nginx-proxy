package core

import (
	"fmt"
	"strings"

	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common/errors"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common/profile"
	dnspod "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/dnspod/v20210323"
	ssl "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/ssl/v20191205"
)

type DNSPodService struct {
	client    *dnspod.Client
	sslClient *ssl.Client
}

type CertificateValidationInfo struct {
	Domain          string
	ValidationName  string
	ValidationValue string
}

func NewDNSPodService(config *TencentCloudConfig) *DNSPodService {
	if config.SecretId == "" || config.SecretKey == "" {
		return nil
	}

	credential := common.NewCredential(config.SecretId, config.SecretKey)

	// 创建 DNSPod 客户端
	dnsCpf := profile.NewClientProfile()
	dnsCpf.HttpProfile.Endpoint = "dnspod.tencentcloudapi.com"
	dnsClient, err := dnspod.NewClient(credential, config.Region, dnsCpf)
	if err != nil {
		return nil
	}

	// 创建 SSL 客户端
	sslCpf := profile.NewClientProfile()
	sslCpf.HttpProfile.Endpoint = "ssl.tencentcloudapi.com"
	sslClient, err := ssl.NewClient(credential, config.Region, sslCpf)
	if err != nil {
		return nil
	}

	return &DNSPodService{
		client:    dnsClient,
		sslClient: sslClient,
	}
}

// CheckValidationRecord 检查域名是否有DNS验证记录
func (d *DNSPodService) CheckValidationRecord(domain string) (bool, error) {
	// 提取主域名
	mainDomain := d.extractMainDomain(domain)

	// 查询DNS记录
	request := dnspod.NewDescribeRecordListRequest()
	request.Domain = common.StringPtr(mainDomain)
	request.RecordType = common.StringPtr("TXT")

	response, err := d.client.DescribeRecordList(request)
	if err != nil {
		if sdkError, ok := err.(*errors.TencentCloudSDKError); ok {
			return false, fmt.Errorf("查询DNS记录失败: %s - %s", sdkError.Code, sdkError.Message)
		}
		return false, fmt.Errorf("查询DNS记录失败: %v", err)
	}

	// 检查是否存在验证记录（通常是 _acme-challenge 开头的TXT记录）
	validationPrefix := "_acme-challenge"

	if response.Response.RecordList != nil {
		for _, record := range response.Response.RecordList {
			if record.Name != nil && record.Type != nil &&
				*record.Type == "TXT" && strings.HasPrefix(*record.Name, validationPrefix) {
				return true, nil
			}
		}
	}

	return false, nil
}

// GetCertificateValidationInfo 获取证书的DNS验证信息
func (d *DNSPodService) GetCertificateValidationInfo(certificateID string) (*CertificateValidationInfo, error) {
	request := ssl.NewDescribeCertificateDetailRequest()
	request.CertificateId = common.StringPtr(certificateID)

	response, err := d.sslClient.DescribeCertificateDetail(request)
	if err != nil {
		if sdkError, ok := err.(*errors.TencentCloudSDKError); ok {
			return nil, fmt.Errorf("获取证书详情失败: %s - %s", sdkError.Code, sdkError.Message)
		}
		return nil, fmt.Errorf("获取证书详情失败: %v", err)
	}
	detail := response.Response.DvRevokeAuthDetail
	if detail == nil || len(detail) == 0 {
		return nil, fmt.Errorf("证书没有DNS验证信息")
	}

	// 获取第一个DNS验证信息
	dvAuth := detail[0]
	if dvAuth.DvAuthDomain == nil || dvAuth.DvAuthKey == nil || dvAuth.DvAuthValue == nil {
		return nil, fmt.Errorf("DNS验证信息不完整")
	}

	return &CertificateValidationInfo{
		Domain:          *dvAuth.DvAuthDomain,
		ValidationName:  *dvAuth.DvAuthKey,
		ValidationValue: *dvAuth.DvAuthValue,
	}, nil
}

// AddValidationRecord 添加DNS验证记录
func (d *DNSPodService) AddValidationRecord(domain, recordName, recordValue string) error {
	// 提取主域名
	mainDomain := d.extractMainDomain(domain)

	// 创建DNS记录
	request := dnspod.NewCreateRecordRequest()
	request.Domain = common.StringPtr(mainDomain)
	request.SubDomain = common.StringPtr(strings.ReplaceAll(recordName, "."+mainDomain, ""))
	request.RecordType = common.StringPtr("TXT")
	request.RecordLine = common.StringPtr("默认")
	request.Value = common.StringPtr(recordValue)
	request.TTL = common.Uint64Ptr(600)

	_, err := d.client.CreateRecord(request)
	if err != nil {
		if sdkError, ok := err.(*errors.TencentCloudSDKError); ok {
			return fmt.Errorf("创建DNS验证记录失败: %s - %s", sdkError.Code, sdkError.Message)
		}
		return fmt.Errorf("创建DNS验证记录失败: %v", err)
	}

	return nil
}

// AddValidationRecordWithValue 使用指定的验证值添加DNS验证记录
func (d *DNSPodService) AddValidationRecordWithValue(domain, recordName, recordValue string) error {
	// 提取主域名
	mainDomain := d.extractMainDomain(domain)

	// 创建DNS记录
	request := dnspod.NewCreateRecordRequest()
	request.Domain = common.StringPtr(mainDomain)
	request.SubDomain = common.StringPtr(recordName)
	request.RecordType = common.StringPtr("TXT")
	request.RecordLine = common.StringPtr("默认")
	request.Value = common.StringPtr(recordValue)
	request.TTL = common.Uint64Ptr(600)

	_, err := d.client.CreateRecord(request)
	if err != nil {
		if sdkError, ok := err.(*errors.TencentCloudSDKError); ok {
			return fmt.Errorf("创建DNS验证记录失败: %s - %s", sdkError.Code, sdkError.Message)
		}
		return fmt.Errorf("创建DNS验证记录失败: %v", err)
	}

	return nil
}

// extractMainDomain 从域名中提取主域名
func (d *DNSPodService) extractMainDomain(domain string) string {
	// 移除通配符前缀
	if strings.HasPrefix(domain, "*.") {
		domain = domain[2:]
	}

	// 简单的主域名提取逻辑
	parts := strings.Split(domain, ".")
	if len(parts) >= 2 {
		return strings.Join(parts[len(parts)-2:], ".")
	}
	return domain
}
