package core

import (
	"fmt"
	"os"
	"path/filepath"
	"text/template"

	"nginx-proxy/internal/db"
)

// Generator 负责生成 Nginx 配置文件
type Generator struct {
	templateDir string
	configDir   string
	template    *template.Template
}

// NewGenerator 创建新的配置生成器
func NewGenerator(templateDir, configDir string) *Generator {
	return &Generator{
		templateDir: templateDir,
		configDir:   configDir,
	}
}

// loadTemplate 加载模板文件
func (g *Generator) loadTemplate() error {
	templatePath := filepath.Join(g.templateDir, "nginx.conf.tpl")

	// 创建带有自定义函数的模板
	tmpl := template.New("nginx.conf.tpl")

	tmpl, err := tmpl.ParseFiles(templatePath)
	if err != nil {
		return fmt.Errorf("failed to parse template: %w", err)
	}
	g.template = tmpl
	return nil
}

// GenerateConfig 生成单个规则的配置文件
func (g *Generator) GenerateConfig(rule *db.Rule) error {
	if g.template == nil {
		if err := g.loadTemplate(); err != nil {
			return err
		}
	}

	// 确保配置目录存在
	if err := os.MkdirAll(g.configDir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	// 转换为模板数据
	templateData, err := g.prepareTemplateData(rule)
	if err != nil {
		return fmt.Errorf("failed to prepare template data: %w", err)
	}

	// 生成配置文件路径
	configPath := filepath.Join(g.configDir, fmt.Sprintf("%s.conf", rule.ID))

	// 创建配置文件
	file, err := os.Create(configPath)
	if err != nil {
		return fmt.Errorf("failed to create config file: %w", err)
	}
	defer func() {
		if closeErr := file.Close(); closeErr != nil {
			// 记录关闭文件时的错误，但不影响主要流程
			fmt.Printf("Warning: Failed to close config file: %v\n", closeErr)
		}
	}()

	// 执行模板渲染
	if err := g.template.Execute(file, templateData); err != nil {
		return fmt.Errorf("failed to execute template: %w", err)
	}

	return nil
}

// DeleteConfig 删除配置文件
func (g *Generator) DeleteConfig(ruleID string) error {
	configPath := filepath.Join(g.configDir, fmt.Sprintf("%s.conf", ruleID))
	if err := os.Remove(configPath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to delete config file: %w", err)
	}
	return nil
}

// prepareTemplateData 准备模板数据
func (g *Generator) prepareTemplateData(rule *db.Rule) (*TemplateData, error) {
	ports, err := rule.GetListenPorts()
	if err != nil {
		return nil, err
	}

	locations, err := rule.GetLocations()
	if err != nil {
		return nil, err
	}

	return &TemplateData{
		ServerName:  rule.ServerName,
		ListenPorts: ports,
		SSLCert:     rule.SSLCert,
		SSLKey:      rule.SSLKey,
		Locations:   locations,
	}, nil
}

// TemplateData 模板数据结构
type TemplateData struct {
	ServerName  string        `json:"server_name"`
	ListenPorts []int         `json:"listen_ports"`
	SSLCert     string        `json:"ssl_cert"`
	SSLKey      string        `json:"ssl_key"`
	Locations   []db.Location `json:"locations"`
}
