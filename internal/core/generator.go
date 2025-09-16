package core

import (
	"fmt"
	"net"
	"os"
	"path/filepath"
	"strings"
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
	tmpl := template.New("nginx.conf.tpl").Funcs(template.FuncMap{
		"generateIPCondition":     generateIPCondition,
		"generateHeaderCondition": generateHeaderCondition,
		"isDefaultRoute":          isDefaultRoute,
		"hasHeaderCondition":      hasHeaderCondition,
		"escapeRegex":             escapeRegex,
		"headerToNginxVar":        headerToNginxVar,
	})

	tmpl, err := tmpl.ParseFiles(templatePath)
	if err != nil {
		return fmt.Errorf("failed to parse template: %w", err)
	}
	g.template = tmpl
	return nil
}

// generateIPCondition 生成 IP 条件匹配语句
func generateIPCondition(conditionIP string) string {
	// 处理默认路由
	if conditionIP == "0.0.0.0/0" || conditionIP == "" {
		return ""
	}

	// 处理单个 IP（带 /32 后缀）
	if strings.HasSuffix(conditionIP, "/32") {
		ip := strings.TrimSuffix(conditionIP, "/32")
		if net.ParseIP(ip) != nil {
			return fmt.Sprintf(`if ($remote_addr = "%s")`, ip)
		}
	}

	// 处理单个 IP（不带后缀）
	if net.ParseIP(conditionIP) != nil {
		return fmt.Sprintf(`if ($remote_addr = "%s")`, conditionIP)
	}

	// 处理 IP 段（CIDR 格式）
	if _, ipNet, err := net.ParseCIDR(conditionIP); err == nil {
		// 对于 IP 段，我们需要使用 Nginx 的 geo 模块或者正则表达式
		// 这里使用正则表达式方式处理常见的 IP 段
		return generateCIDRCondition(conditionIP, ipNet)
	}

	// 如果都不匹配，回退到原始的正则匹配（转义特殊字符）
	escapedIP := escapeRegex(conditionIP)
	return fmt.Sprintf(`if ($remote_addr ~ "^%s$")`, escapedIP)
}

// generateCIDRCondition 生成 CIDR 格式的条件
func generateCIDRCondition(cidr string, ipNet *net.IPNet) string {
	// 获取网络地址和掩码
	network := ipNet.IP.String()
	maskSize, _ := ipNet.Mask.Size()

	// 根据不同的掩码长度生成不同的正则表达式
	switch {
	case maskSize == 32:
		// /32 是单个 IP
		return fmt.Sprintf(`if ($remote_addr = "%s")`, network)
	case maskSize == 24:
		// /24 网段，匹配前三段
		parts := strings.Split(network, ".")
		if len(parts) >= 3 {
			pattern := fmt.Sprintf(`^%s\.%s\.%s\.\d+$`,
				escapeRegex(parts[0]),
				escapeRegex(parts[1]),
				escapeRegex(parts[2]))
			return fmt.Sprintf(`if ($remote_addr ~ "%s")`, pattern)
		}
	case maskSize == 16:
		// /16 网段，匹配前两段
		parts := strings.Split(network, ".")
		if len(parts) >= 2 {
			pattern := fmt.Sprintf(`^%s\.%s\.\d+\.\d+$`,
				escapeRegex(parts[0]),
				escapeRegex(parts[1]))
			return fmt.Sprintf(`if ($remote_addr ~ "%s")`, pattern)
		}
	case maskSize == 8:
		// /8 网段，匹配第一段
		parts := strings.Split(network, ".")
		if len(parts) >= 1 {
			pattern := fmt.Sprintf(`^%s\.\d+\.\d+\.\d+$`, escapeRegex(parts[0]))
			return fmt.Sprintf(`if ($remote_addr ~ "%s")`, pattern)
		}
	}

	// 对于其他复杂的 CIDR，使用 geo 模块的方式
	// 这需要在 http 块中定义 geo 变量，这里先用注释提示
	return fmt.Sprintf(`# TODO: Use geo module for complex CIDR %s
        if ($remote_addr ~ "%s")`, cidr, escapeRegex(cidr))
}

// isDefaultRoute 检查是否为默认路由
func isDefaultRoute(conditionIP string) bool {
	return conditionIP == "0.0.0.0/0" || conditionIP == ""
}

// escapeRegex 转义正则表达式中的特殊字符
func escapeRegex(s string) string {
	// 转义正则表达式中的特殊字符
	specialChars := []string{".", "+", "*", "?", "^", "$", "(", ")", "[", "]", "{", "}", "|", "\\", "/"}
	result := s
	for _, char := range specialChars {
		result = strings.ReplaceAll(result, char, "\\"+char)
	}
	return result
}

// generateHeaderCondition 生成HTTP头部条件匹配语句
func generateHeaderCondition(headers map[string]string) string {
	if len(headers) == 0 {
		return ""
	}

	var conditions []string
	for key, value := range headers {
		// 将头部名称转换为nginx变量格式
		headerVar := strings.ToLower(strings.ReplaceAll(key, "-", "_"))
		headerVar = fmt.Sprintf("http_%s", headerVar)

		// 生成条件语句
		condition := fmt.Sprintf(`($%s = "%s")`, headerVar, value)
		conditions = append(conditions, condition)
	}

	// 使用 AND 逻辑连接多个条件
	if len(conditions) == 1 {
		return fmt.Sprintf("if %s", conditions[0])
	}
	return fmt.Sprintf("if (%s)", strings.Join(conditions, " and "))
}

// hasHeaderCondition 检查是否有头部条件
func hasHeaderCondition(headers map[string]string) bool {
	return len(headers) > 0
}

// headerToNginxVar 将HTTP头部名称转换为nginx变量名
func headerToNginxVar(headerName string) string {
	// 转换为小写并将连字符替换为下划线
	return strings.ToLower(strings.ReplaceAll(headerName, "-", "_"))
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
