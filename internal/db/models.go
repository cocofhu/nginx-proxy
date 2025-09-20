package db

import (
	"encoding/json"
	"time"

	"gorm.io/gorm"
)

// Rule 代表一个 Nginx 反向代理规则
type Rule struct {
	ID          string         `json:"id" gorm:"primaryKey"`
	ServerName  string         `json:"server_name" gorm:"not null"`
	ListenPorts string         `json:"listen_ports" gorm:"column:listen_ports"` // JSON 存储
	SSLCert     string         `json:"ssl_cert"`
	SSLKey      string         `json:"ssl_key"`
	Locations   string         `json:"locations" gorm:"column:locations;type:text"` // JSON 存储
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `json:"-" gorm:"index"`
}

// Location 代表一个 location 配置
type Location struct {
	Path      string     `json:"path"`
	Upstreams []Upstream `json:"upstreams"`
}

// Upstream 代表一个上游服务器配置
type Upstream struct {
	ConditionIP string            `json:"condition_ip"`      // CIDR 格式
	Target      string            `json:"target"`            // http://host:port
	Headers     map[string]string `json:"headers,omitempty"` // HTTP头部路由条件
}

// RuleResponse 用于 API 响应
type RuleResponse struct {
	ID          string     `json:"id"`
	ServerName  string     `json:"server_name"`
	ListenPorts []int      `json:"listen_ports"`
	SSLCert     string     `json:"ssl_cert"`
	SSLKey      string     `json:"ssl_key"`
	Enabled     bool       `json:"enabled"`
	Locations   []Location `json:"locations"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
}

// GetListenPorts 解析监听端口
func (r *Rule) GetListenPorts() ([]int, error) {
	var ports []int
	if r.ListenPorts == "" {
		return ports, nil
	}
	err := json.Unmarshal([]byte(r.ListenPorts), &ports)
	return ports, err
}

// SetListenPorts 设置监听端口
func (r *Rule) SetListenPorts(ports []int) error {
	data, err := json.Marshal(ports)
	if err != nil {
		return err
	}
	r.ListenPorts = string(data)
	return nil
}

// GetLocations 解析 location 配置
func (r *Rule) GetLocations() ([]Location, error) {
	var locations []Location
	if r.Locations == "" {
		return locations, nil
	}
	err := json.Unmarshal([]byte(r.Locations), &locations)
	return locations, err
}

// SetLocations 设置 location 配置
func (r *Rule) SetLocations(locations []Location) error {
	data, err := json.Marshal(locations)
	if err != nil {
		return err
	}
	r.Locations = string(data)
	return nil
}

// ToResponse 转换为响应格式
func (r *Rule) ToResponse() (*RuleResponse, error) {
	ports, err := r.GetListenPorts()
	if err != nil {
		return nil, err
	}

	locations, err := r.GetLocations()
	if err != nil {
		return nil, err
	}

	return &RuleResponse{
		ID:          r.ID,
		ServerName:  r.ServerName,
		ListenPorts: ports,
		SSLCert:     r.SSLCert,
		SSLKey:      r.SSLKey,
		Enabled:     true,
		Locations:   locations,
		CreatedAt:   r.CreatedAt,
		UpdatedAt:   r.UpdatedAt,
	}, nil
}

// Certificate 代表一个SSL证书
type Certificate struct {
	ID               string         `json:"id" gorm:"primaryKey"` // 腾讯云的证书ID和SourceID一致
	Name             string         `json:"name" gorm:"not null"`
	Domain           string         `json:"domain"`
	CertPath         string         `json:"cert_path" gorm:"not null"`
	KeyPath          string         `json:"key_path" gorm:"not null"`
	ExpiresAt        time.Time      `json:"expires_at"`
	Source           string         `json:"source"`             // 证书来源: "upload", "tencent_cloud"
	SourceID         string         `json:"source_id"`          // 来源ID，如腾讯云证书ID
	Status           string         `json:"status"`             // 证书状态: "active", "renewing", "expired"
	RenewalSourceID  string         `json:"renewal_source_id"`  // 续期新证书的ID
	OriginalSourceID string         `json:"original_source_id"` // 原始证书ID 指向上一个证书
	CreatedAt        time.Time      `json:"created_at"`
	UpdatedAt        time.Time      `json:"updated_at"`
	DeletedAt        gorm.DeletedAt `json:"-" gorm:"index"`
}

// AuthRecord 待处理的DNS记录, 由于证书验证会导致DNS记录越来越多 这里需要记录下来定期删除
type AuthRecord struct {
	ID            string         `json:"id" gorm:"primaryKey"`
	Domain        string         `json:"domain"`                // 主域名
	Key           string         `json:"key" gorm:"not null"`   // 记录Key
	Value         string         `json:"value" gorm:"not null"` // 记录值
	Type          string         `json:"type"`                  // 记录类型: TXT
	Source        string         `json:"source"`                // 域名托管位置: "cloudflare", "tencent_cloud"
	CertificateId string         `json:"certificate_id"`        // 来源ID，如腾讯云证书ID
	CreatedAt     time.Time      `json:"created_at"`
	UpdatedAt     time.Time      `json:"updated_at"`
	DeletedAt     gorm.DeletedAt `json:"-" gorm:"index"`
}
