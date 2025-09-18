package core

import (
	"fmt"
	"os/exec"
)

// NginxManager 负责管理 Nginx 进程
type NginxManager struct {
	nginxPath string
}

// NewNginxManager 创建新的 Nginx 管理器
func NewNginxManager(nginxPath string) *NginxManager {
	return &NginxManager{
		nginxPath: nginxPath,
	}
}

// TestConfig 测试 Nginx 配置
func (n *NginxManager) TestConfig() error {
	cmd := exec.Command(n.nginxPath, "-t")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("nginx config test failed: %s, output: %s", err, string(output))
	}
	return nil
}

// Reload 重新加载 Nginx 配置
func (n *NginxManager) Reload() error {
	// 先测试配置
	if err := n.TestConfig(); err != nil {
		return err
	}

	// 执行重新加载
	cmd := exec.Command(n.nginxPath, "-s", "reload")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("nginx reload failed: %s, output: %s", err, string(output))
	}
	return nil
}
