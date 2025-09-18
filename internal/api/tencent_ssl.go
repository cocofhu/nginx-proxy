package api

import (
	"log"
	"net/http"
	"nginx-proxy/internal/core"

	"github.com/gin-gonic/gin"
)

// 腾讯云SSL证书管理API
// 提供腾讯云SSL证书的申请、查询、下载、删除等功能

// ApplyTencentCertificate 申请腾讯云证书
func (h *Handler) ApplyTencentCertificate(c *gin.Context) {
	var req core.ApplyCertificateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请求参数格式错误: " + err.Error()})
		return
	}

	// 检查腾讯云配置
	if h.tencentSSL == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "腾讯云服务未配置"})
		return
	}

	// 验证域名格式
	if req.Domain == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "域名不能为空"})
		return
	}

	response, err := h.tencentSSL.ApplyCertificate(&req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "申请证书失败: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, response)
}

// CheckTencentCertificateStatus 检查腾讯云证书状态
func (h *Handler) CheckTencentCertificateStatus(c *gin.Context) {
	certificateID := c.Param("id")

	if certificateID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "证书ID不能为空"})
		return
	}

	if h.tencentSSL == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "腾讯云服务未配置"})
		return
	}

	response, err := h.tencentSSL.CheckCertificateStatus(certificateID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "查询证书状态失败: " + err.Error()})
		return
	}

	if response.Reloaded {
		log.Printf("reload nginx conf due to certificate updated")
		if err := h.RegenerateAllConfigs(); err != nil {
			log.Printf("Warning: failed to reload nginx conf: %v", err)
		}
	}

	c.JSON(http.StatusOK, response)
}

// DownloadTencentCertificate 下载腾讯云证书
func (h *Handler) DownloadTencentCertificate(c *gin.Context) {
	certificateID := c.Param("id")

	if certificateID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "证书ID不能为空"})
		return
	}

	if h.tencentSSL == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "腾讯云服务未配置"})
		return
	}

	if err := h.tencentSSL.DownloadCertificate(certificateID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "下载证书失败: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "证书下载成功"})
}

// UpdateCertificateName 更新证书名称
func (h *Handler) UpdateCertificateName(c *gin.Context) {
	certificateID := c.Param("id")

	if certificateID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "证书ID不能为空"})
		return
	}

	var req struct {
		Name string `json:"name" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请求参数错误: " + err.Error()})
		return
	}

	if h.tencentSSL == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "腾讯云服务未配置"})
		return
	}

	if err := h.tencentSSL.UpdateCertificateName(certificateID, req.Name); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "更新证书名称失败: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "证书名称已更新"})
}

// DeleteTencentCertificate 删除腾讯云证书
func (h *Handler) DeleteTencentCertificate(c *gin.Context) {
	certificateID := c.Param("id")

	if certificateID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "证书ID不能为空"})
		return
	}

	if h.tencentSSL == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "腾讯云服务未配置"})
		return
	}

	if err := h.tencentSSL.DeleteTencentCertificate(certificateID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "删除证书失败: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "证书删除成功"})
}

// RenewTencentCertificate 续期腾讯云证书（重新申请新证书）
func (h *Handler) RenewTencentCertificate(c *gin.Context) {
	certificateID := c.Param("id")

	if certificateID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "证书ID不能为空"})
		return
	}

	if h.tencentSSL == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "腾讯云服务未配置"})
		return
	}

	response, err := h.tencentSSL.RenewTencentCertificate(certificateID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "续期证书失败: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":       "证书续期申请成功",
		"old_cert_id":   certificateID,
		"new_cert_id":   response.NewCertificateID,
		"status":        response.Status,
		"validate_info": response.ValidateInfo,
	})
}
