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
		Locations:   locations,
		CreatedAt:   r.CreatedAt,
		UpdatedAt:   r.UpdatedAt,
	}, nil
}

// Certificate 代表一个SSL证书
type Certificate struct {
	ID               string         `json:"id" gorm:"primaryKey"`
	Name             string         `json:"name" gorm:"not null"`
	Domain           string         `json:"domain"`
	CertPath         string         `json:"cert_path" gorm:"not null"`
	KeyPath          string         `json:"key_path" gorm:"not null"`
	ExpiresAt        time.Time      `json:"expires_at"`
	Source           string         `json:"source"`             // 证书来源: "upload", "tencent_cloud"
	SourceID         string         `json:"source_id"`          // 来源ID，如腾讯云证书ID
	Status           string         `json:"status"`             // 证书状态: "active", "renewing", "expired"
	RenewalSourceID  string         `json:"renewal_source_id"`  // 续期新证书的ID
	OriginalSourceID string         `json:"original_source_id"` // 原始证书ID（用于续期证书）
	CreatedAt        time.Time      `json:"created_at"`
	UpdatedAt        time.Time      `json:"updated_at"`
	DeletedAt        gorm.DeletedAt `json:"-" gorm:"index"`
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
