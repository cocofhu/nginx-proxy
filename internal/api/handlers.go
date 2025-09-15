package api

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"gorm.io/gorm"

	"nginx-proxy/internal/core"
	"nginx-proxy/internal/db"
)

// Handler API 处理器
type Handler struct {
	db           *gorm.DB
	generator    *core.Generator
	nginxManager *core.NginxManager
	certDir      string
}

// NewHandler 创建新的 API 处理器
func NewHandler(database *gorm.DB, generator *core.Generator, nginxManager *core.NginxManager, certDir string) *Handler {
	h := &Handler{
		db:           database,
		generator:    generator,
		nginxManager: nginxManager,
		certDir:      certDir,
	}

	return h
}

// CreateRuleRequest 创建规则请求
type CreateRuleRequest struct {
	ServerName  string        `json:"server_name" binding:"required"`
	ListenPorts []int         `json:"listen_ports" binding:"required"`
	SSLCert     string        `json:"ssl_cert"`
	SSLKey      string        `json:"ssl_key"`
	Locations   []db.Location `json:"locations" binding:"required"`
}

// validateSSLConfig 验证 SSL 配置
func (h *Handler) validateSSLConfig(req *CreateRuleRequest) error {
	// 如果提供了证书或密钥中的任何一个，则两者都必须提供
	if (req.SSLCert != "" && req.SSLKey == "") || (req.SSLCert == "" && req.SSLKey != "") {
		return fmt.Errorf("ssl_cert and ssl_key must be provided together or both omitted")
	}

	// 如果提供了证书配置，检查文件是否存在
	if req.SSLCert != "" && req.SSLKey != "" {
		if _, err := os.Stat(req.SSLCert); os.IsNotExist(err) {
			return fmt.Errorf("ssl certificate file does not exist: %s", req.SSLCert)
		}
		if _, err := os.Stat(req.SSLKey); os.IsNotExist(err) {
			return fmt.Errorf("ssl key file does not exist: %s", req.SSLKey)
		}
	}

	return nil
}

// validateUniqueServerNamePort 验证域名和端口组合的唯一性
func (h *Handler) validateUniqueServerNamePort(serverName string, listenPorts []int, excludeID string) error {
	var existingRules []db.Rule
	query := h.db.Where("server_name = ?", serverName)
	if excludeID != "" {
		query = query.Where("id != ?", excludeID)
	}

	if err := query.Find(&existingRules).Error; err != nil {
		return fmt.Errorf("failed to check existing rules: %w", err)
	}

	for _, existingRule := range existingRules {
		existingPorts, err := existingRule.GetListenPorts()
		if err != nil {
			continue // 跳过无法解析端口的规则
		}

		// 检查是否有端口冲突
		for _, newPort := range listenPorts {
			for _, existingPort := range existingPorts {
				if newPort == existingPort {
					return fmt.Errorf("server_name '%s' with port %d already exists", serverName, newPort)
				}
			}
		}
	}

	return nil
}

// GetRules 获取所有规则
func (h *Handler) GetRules(c *gin.Context) {
	var rules []db.Rule
	if err := h.db.Find(&rules).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	var responses []*db.RuleResponse
	for _, rule := range rules {
		resp, err := rule.ToResponse()
		if err != nil {
			log.Printf("Error converting rule to response: %v", err)
			continue
		}
		responses = append(responses, resp)
	}

	c.JSON(http.StatusOK, gin.H{"rules": responses})
}

// GetRule 获取单个规则
func (h *Handler) GetRule(c *gin.Context) {
	id := c.Param("id")

	var rule db.Rule
	if err := h.db.First(&rule, "id = ?", id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "Rule not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	resp, err := rule.ToResponse()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, resp)
}

// CreateRule 创建新规则
func (h *Handler) CreateRule(c *gin.Context) {
	var req CreateRuleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 验证 SSL 配置
	if err := h.validateSSLConfig(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 验证域名和端口组合的唯一性
	if err := h.validateUniqueServerNamePort(req.ServerName, req.ListenPorts, ""); err != nil {
		c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
		return
	}

	// 创建新规则
	rule := db.Rule{
		ID:         uuid.New().String(),
		ServerName: req.ServerName,
		SSLCert:    req.SSLCert,
		SSLKey:     req.SSLKey,
	}

	// 设置端口和位置
	if err := rule.SetListenPorts(req.ListenPorts); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to set listen ports"})
		return
	}

	if err := rule.SetLocations(req.Locations); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to set locations"})
		return
	}

	// 生成配置文件
	if err := h.generator.GenerateConfig(&rule); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate config: " + err.Error()})
		return
	}

	// 测试 Nginx 配置
	if err := h.nginxManager.TestConfig(); err != nil {
		// 配置测试失败，删除生成的配置文件
		if deleteErr := h.generator.DeleteConfig(rule.ID); deleteErr != nil {
			log.Printf("Warning: Failed to cleanup config file after test failure: %v", deleteErr)
		}
		c.JSON(http.StatusBadRequest, gin.H{"error": "Nginx config test failed: " + err.Error()})
		return
	}

	// 保存到数据库
	if err := h.db.Create(&rule).Error; err != nil {
		// 数据库保存失败，删除配置文件
		if deleteErr := h.generator.DeleteConfig(rule.ID); deleteErr != nil {
			log.Printf("Warning: Failed to cleanup config file after database save failure: %v", deleteErr)
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// 重新加载 Nginx
	if err := h.nginxManager.Reload(); err != nil {
		log.Printf("Warning: Failed to reload nginx: %v", err)
	}

	resp, err := rule.ToResponse()
	if err != nil {
		log.Printf("Warning: Failed to convert rule to response: %v", err)
		// 仍然返回成功，因为规则已经创建成功
		c.JSON(http.StatusCreated, gin.H{"message": "Rule created successfully", "id": rule.ID})
		return
	}
	c.JSON(http.StatusCreated, resp)
}

// UpdateRule 更新规则
func (h *Handler) UpdateRule(c *gin.Context) {
	id := c.Param("id")

	var rule db.Rule
	if err := h.db.First(&rule, "id = ?", id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "Rule not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	var req CreateRuleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 验证 SSL 配置
	if err := h.validateSSLConfig(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 验证域名和端口组合的唯一性（排除当前规则）
	if err := h.validateUniqueServerNamePort(req.ServerName, req.ListenPorts, id); err != nil {
		c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
		return
	}

	// 更新规则字段
	rule.ServerName = req.ServerName
	rule.SSLCert = req.SSLCert
	rule.SSLKey = req.SSLKey

	if err := rule.SetListenPorts(req.ListenPorts); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to set listen ports"})
		return
	}

	if err := rule.SetLocations(req.Locations); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to set locations"})
		return
	}

	// 重新生成配置文件
	if err := h.generator.GenerateConfig(&rule); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate config: " + err.Error()})
		return
	}

	// 测试 Nginx 配置
	if err := h.nginxManager.TestConfig(); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Nginx config test failed: " + err.Error()})
		return
	}

	// 更新数据库
	if err := h.db.Save(&rule).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// 重新加载 Nginx
	if err := h.nginxManager.Reload(); err != nil {
		log.Printf("Warning: Failed to reload nginx: %v", err)
	}

	resp, err := rule.ToResponse()
	if err != nil {
		log.Printf("Warning: Failed to convert rule to response: %v", err)
		// 仍然返回成功，因为规则已经更新成功
		c.JSON(http.StatusOK, gin.H{"message": "Rule updated successfully", "id": rule.ID})
		return
	}
	c.JSON(http.StatusOK, resp)
}

// DeleteRule 删除规则
func (h *Handler) DeleteRule(c *gin.Context) {
	id := c.Param("id")

	var rule db.Rule
	if err := h.db.First(&rule, "id = ?", id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "Rule not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// 删除配置文件
	if err := h.generator.DeleteConfig(rule.ID); err != nil {
		log.Printf("Warning: Failed to delete config file: %v", err)
		// 继续执行，不阻止删除操作
	}

	// 从数据库删除
	if err := h.db.Delete(&rule).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// 重新加载 Nginx
	if err := h.nginxManager.Reload(); err != nil {
		log.Printf("Warning: Failed to reload nginx: %v", err)
	}

	c.JSON(http.StatusOK, gin.H{"message": "Rule deleted successfully"})
}

// ReloadNginx 手动重新加载 Nginx
func (h *Handler) ReloadNginx(c *gin.Context) {
	if err := h.nginxManager.Reload(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Nginx reloaded successfully"})
}

// RegenerateAllConfigs 重新生成所有配置文件
func (h *Handler) RegenerateAllConfigs() error {
	var rules []db.Rule
	if err := h.db.Find(&rules).Error; err != nil {
		return err
	}

	for _, rule := range rules {
		if err := h.generator.GenerateConfig(&rule); err != nil {
			log.Printf("Failed to regenerate config for rule %s: %v", rule.ID, err)
		}
	}

	return nil
}
