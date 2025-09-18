package core

import (
	"encoding/json"
	"os"
)

type Config struct {
	Server       ServerConfig       `json:"server"`
	Database     DatabaseConfig     `json:"database"`
	Nginx        NginxConfig        `json:"nginx"`
	SSL          SSLConfig          `json:"ssl"`
	TencentCloud TencentCloudConfig `json:"tencent_cloud"`
}

type ServerConfig struct {
	Port int    `json:"port"`
	Host string `json:"host"`
}

type DatabaseConfig struct {
	Driver string `json:"driver"`
	DSN    string `json:"dsn"`
}

type NginxConfig struct {
	Path        string `json:"path"`
	ConfigDir   string `json:"config_dir"`
	TemplateDir string `json:"template_dir"`
}

type SSLConfig struct {
	CertDir string `json:"cert_dir"`
}

type TencentCloudConfig struct {
	SecretId  string `json:"secret_id"`
	SecretKey string `json:"secret_key"`
	Region    string `json:"region"`
}

// LoadConfig 加载配置文件
func LoadConfig(configPath string) (*Config, error) {
	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, err
	}

	var config Config
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, err
	}

	// 设置默认值
	if config.Server.Port == 0 {
		config.Server.Port = 8080
	}
	if config.Server.Host == "" {
		config.Server.Host = "0.0.0.0"
	}
	if config.SSL.CertDir == "" {
		config.SSL.CertDir = "./certs"
	}
	if config.Nginx.Path == "" {
		config.Nginx.Path = "/usr/sbin/nginx"
	}
	if config.Nginx.ConfigDir == "" {
		config.Nginx.ConfigDir = "/etc/nginx/conf.d"
	}
	if config.Nginx.TemplateDir == "" {
		config.Nginx.TemplateDir = "./template"
	}
	if config.TencentCloud.Region == "" {
		config.TencentCloud.Region = "ap-beijing"
	}

	if envSecretId := os.Getenv("TENCENT_SECRET_ID"); envSecretId != "" {
		config.TencentCloud.SecretId = envSecretId
	}

	if envSecretKey := os.Getenv("TENCENT_SECRET_KEY"); envSecretKey != "" {
		config.TencentCloud.SecretKey = envSecretKey
	}

	if envRegion := os.Getenv("TENCENT_REGION"); envRegion != "" {
		config.TencentCloud.Region = envRegion
	}

	return &config, nil
}
