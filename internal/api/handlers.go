package api

import (
	"errors"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/patrickmn/go-cache"

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
	cache        *cache.Cache
	tencentSSL   *core.TencentSSLService
}

// NewHandler 创建新的 API 处理器
func NewHandler(database *gorm.DB, generator *core.Generator, nginxManager *core.NginxManager, certDir string, tencentSSL *core.TencentSSLService) *Handler {
	h := &Handler{
		db:           database,
		generator:    generator,
		nginxManager: nginxManager,
		certDir:      certDir,
		cache:        cache.New(5*time.Minute, 10*time.Minute), // 5分钟过期，10分钟清理
		tencentSSL:   tencentSSL,
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

// clearRuleCache 清除指定 server_name 的缓存
func (h *Handler) clearRuleCache(serverName string) {
	cacheKey := fmt.Sprintf("rule:%s", serverName)
	h.cache.Delete(cacheKey)
	log.Printf("Cleared cache for server_name: %s", serverName)
}

// clearAllRuleCache 清除所有规则缓存
func (h *Handler) clearAllRuleCache() {
	h.cache.Flush()
	log.Printf("Cleared all rule cache")
}

// Upstream 上游服务器结构
type Upstream struct {
	Target      string            `json:"target"`
	ConditionIP string            `json:"condition_ip"`
	Headers     map[string]string `json:"headers"`
}

// RouteRequest 路由请求结构（简化版，配置从数据库查询）
type RouteRequest struct {
	Path       string            `json:"path"`
	RemoteAddr string            `json:"remote_addr"`
	Headers    map[string]string `json:"headers"`
	ServerName string            `json:"server_name"`
}

// RouteResponse 路由响应结构
type RouteResponse struct {
	Target string `json:"target"`
	Match  bool   `json:"match"`
}

// Route 统一路由接口（供 OpenResty 调用）
func (h *Handler) Route(c *gin.Context) {
	var req RouteRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		log.Printf("Route request binding error: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format"})
		return
	}

	// 验证请求数据
	if req.Path == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Path is required"})
		return
	}

	log.Printf("Route request: path=%s, remote_addr=%s, server_name=%s, headers=%v",
		req.Path, req.RemoteAddr, req.ServerName, req.Headers)

	// 从缓存或数据库查询匹配的规则
	var rule db.Rule
	cacheKey := fmt.Sprintf("rule:%s", req.ServerName)

	// 先尝试从缓存获取
	if cached, found := h.cache.Get(cacheKey); found {
		rule = cached.(db.Rule)
		log.Printf("Cache hit for server_name: %s", req.ServerName)
	} else {
		// 缓存未命中，从数据库查询
		result := h.db.Where("server_name = ?", req.ServerName).First(&rule)
		if result.Error != nil {
			log.Printf("No rule found for server_name: %s", req.ServerName)
			c.JSON(http.StatusOK, RouteResponse{Target: "", Match: false})
			return
		}

		// 存入缓存
		h.cache.Set(cacheKey, rule, cache.DefaultExpiration)
		log.Printf("Cache miss, loaded from DB for server_name: %s", req.ServerName)
	}

	// 解析 locations 配置
	locations, err := rule.GetLocations()
	if err != nil {
		log.Printf("Error parsing locations: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Configuration error"})
		return
	}

	// 查找匹配的 location 和 upstream
	for _, location := range locations {
		if h.matchPath(req.Path, location.Path) {
			for i, upstream := range location.Upstreams {
				if h.matchUpstream(req, upstream) {
					log.Printf("Route matched location=%s, upstream %d: %s",
						location.Path, i, upstream.Target)
					c.JSON(http.StatusOK, RouteResponse{
						Target: upstream.Target,
						Match:  true,
					})
					return
				}
			}
		}
	}

	// 如果没有匹配，返回空（使用默认）
	log.Printf("No route matched for path=%s, server_name=%s", req.Path, req.ServerName)
	c.JSON(http.StatusOK, RouteResponse{Target: "", Match: false})
}

// matchPath 检查路径是否匹配
func (h *Handler) matchPath(requestPath, locationPath string) bool {
	// 简单的路径匹配，可以扩展为更复杂的匹配规则
	if locationPath == "/" {
		return true // 根路径匹配所有
	}
	return strings.HasPrefix(requestPath, locationPath)
}

// matchUpstream 检查上游服务器是否匹配
func (h *Handler) matchUpstream(req RouteRequest, upstream db.Upstream) bool {
	// 检查 IP 条件
	if upstream.ConditionIP != "" && !h.matchIP(req.RemoteAddr, upstream.ConditionIP) {
		log.Printf("IP condition not matched: %s not in %s", req.RemoteAddr, upstream.ConditionIP)
		return false
	}

	// 检查头部条件（且关系）
	if len(upstream.Headers) > 0 && !h.matchHeaders(req.Headers, upstream.Headers) {
		log.Printf("Header conditions not matched: expected=%v, actual=%v", upstream.Headers, req.Headers)
		return false
	}

	return true
}

// matchIP 检查 IP 是否匹配
func (h *Handler) matchIP(remoteAddr, conditionIP string) bool {
	// 空条件或默认路由，匹配所有
	if conditionIP == "" || conditionIP == "0.0.0.0/0" {
		return true
	}

	// 解析客户端 IP
	clientIP := net.ParseIP(remoteAddr)
	if clientIP == nil {
		log.Printf("Warning: Invalid client IP: %s", remoteAddr)
		return false
	}

	// 检查是否为 CIDR 格式
	if strings.Contains(conditionIP, "/") {
		_, ipNet, err := net.ParseCIDR(conditionIP)
		if err != nil {
			log.Printf("Warning: Invalid CIDR format: %s", conditionIP)
			return false
		}
		return ipNet.Contains(clientIP)
	}

	// 单个 IP 匹配
	targetIP := net.ParseIP(conditionIP)
	if targetIP == nil {
		log.Printf("Warning: Invalid target IP: %s", conditionIP)
		return false
	}

	return clientIP.Equal(targetIP)
}

// matchHeaders 检查头部是否匹配（且关系）
func (h *Handler) matchHeaders(requestHeaders, expectedHeaders map[string]string) bool {
	// 如果没有期望的头部条件，直接匹配
	if len(expectedHeaders) == 0 {
		return true
	}

	// 创建大小写不敏感的请求头部映射
	normalizedRequestHeaders := make(map[string]string)
	for key, value := range requestHeaders {
		normalizedRequestHeaders[strings.ToLower(key)] = value
	}

	// 检查所有期望的头部条件是否都匹配
	for expectedKey, expectedValue := range expectedHeaders {
		normalizedKey := strings.ToLower(expectedKey)
		actualValue, exists := normalizedRequestHeaders[normalizedKey]

		if !exists {
			log.Printf("Header not found: %s", expectedKey)
			return false
		}

		if actualValue != expectedValue {
			log.Printf("Header value mismatch: %s expected=%s actual=%s",
				expectedKey, expectedValue, actualValue)
			return false
		}

		log.Printf("Header matched: %s=%s", expectedKey, expectedValue)
	}

	return true
}

// validateUniqueServerName 验证域名的唯一性（一个域名只能创建一条记录）
func (h *Handler) validateUniqueServerName(serverName string, excludeID string) error {
	var count int64
	query := h.db.Model(&db.Rule{}).Where("server_name = ?", serverName)
	if excludeID != "" {
		query = query.Where("id != ?", excludeID)
	}

	if err := query.Count(&count).Error; err != nil {
		return fmt.Errorf("failed to check existing rules: %w", err)
	}

	if count > 0 {
		return fmt.Errorf("server_name '%s' already exists, one domain can only have one record", serverName)
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
		if errors.Is(err, gorm.ErrRecordNotFound) {
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

	// 验证域名唯一性
	if err := h.validateUniqueServerName(req.ServerName, ""); err != nil {
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

	// 清除缓存
	h.clearRuleCache(rule.ServerName)

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
		if errors.Is(err, gorm.ErrRecordNotFound) {
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
	if err := h.validateUniqueServerName(req.ServerName, id); err != nil {
		c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
		return
	}

	// 保存旧的 server_name 用于清除缓存
	oldServerName := rule.ServerName

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

	// 清除缓存（清除新旧两个 server_name 的缓存）
	h.clearRuleCache(oldServerName)
	if oldServerName != rule.ServerName {
		h.clearRuleCache(rule.ServerName)
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
		if errors.Is(err, gorm.ErrRecordNotFound) {
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
