package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"nginx-proxy/internal/api"
	"nginx-proxy/internal/core"
	"nginx-proxy/internal/db"

	"gopkg.in/natefinch/lumberjack.v2"
	// 使用纯 Go 的 SQLite 驱动
	_ "modernc.org/sqlite"
)

func main() {
	// 加载配置
	configFile := flag.String("config", "config.json", "配置文件路径")
	flag.Parse()

	config, err := core.LoadConfig(*configFile)
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// 确保 logs 目录存在
	if err := os.MkdirAll("./logs", 0755); err != nil {
		log.Fatalf("Failed to create logs directory: %v", err)
	}

	// 设置应用日志输出到文件
	appLogger := &lumberjack.Logger{
		Filename:   "./logs/app.log",
		MaxSize:    10, // MB
		MaxBackups: 7,
		MaxAge:     30, // 保留天数
		Compress:   true,
	}
	log.SetOutput(appLogger)
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	// 设置 Gin 日志输出到文件
	ginLogger := &lumberjack.Logger{
		Filename:   "./logs/gin.log",
		MaxSize:    10, // MB
		MaxBackups: 7,
		MaxAge:     30, // 保留天数
		Compress:   true,
	}
	gin.DefaultWriter = io.MultiWriter(ginLogger, os.Stdout) // 同时输出到文件和控制台
	gin.DefaultErrorWriter = io.MultiWriter(ginLogger, os.Stderr) // 错误日志也输出到文件和控制台

	// 初始化数据库 - 使用纯 Go SQLite 驱动
	database, err := gorm.Open(sqlite.Dialector{
		DriverName: "sqlite",
		DSN:        config.Database.DSN,
	}, &gorm.Config{})
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	// 自动迁移数据库
	if err := database.AutoMigrate(&db.Rule{}, &db.Certificate{}, &db.AuthRecord{}); err != nil {
		log.Fatalf("Failed to migrate database: %v", err)
	}

	// 迁移证书表新字段
	if err := db.MigrateCertificateTable(database); err != nil {
		log.Printf("Warning: Failed to migrate certificate table: %v", err)
	}

	// 迁移证书表续期相关字段
	if err := db.MigrateCertificateTableV2(database); err != nil {
		log.Printf("Warning: Failed to migrate certificate table V2: %v", err)
	}

	// 迁移证书表，确保所有字段都存在
	if err := db.MigrateCertificateTableV3(database); err != nil {
		log.Printf("Warning: Failed to migrate certificate table V3: %v", err)
	}

	// 初始化核心组件
	generator := core.NewGenerator(config.Nginx.TemplateDir, config.Nginx.ConfigDir)
	nginxManager := core.NewNginxManager(config.Nginx.Path)
	// 初始化腾讯云SSL服务（如果配置了）
	var tencentSSL *core.TencentSSLService
	if config.TencentCloud.SecretId != "" && config.TencentCloud.SecretKey != "" {
		tencentSSL = core.NewTencentSSLService(&config.TencentCloud, &config.Cloudflare, database, config.SSL.CertDir)
		log.Println("Tencent Cloud SSL service initialized")
	} else {
		log.Println("Tencent Cloud SSL service not configured")
	}

	// 初始化清理服务
	cleanupConfig := core.CleanupConfig{
		CloudflareAPIToken: config.Cloudflare.Token,
		TencentSecretId:    config.TencentCloud.SecretId,
		TencentSecretKey:   config.TencentCloud.SecretKey,
		TencentRegion:      config.TencentCloud.Region,
		CleanupInterval:    time.Minute,
	}

	cleanupService, err := core.NewCleanupService(database, cleanupConfig)
	if err != nil {
		log.Printf("Warning: Failed to initialize cleanup service: %v", err)
	} else {
		// 启动清理服务
		cleanupService.Start()
		log.Println("Certificate cleanup service started")
	}

	// 初始化API处理器
	handler := api.NewHandler(database, generator, nginxManager, config.SSL.CertDir, tencentSSL)

	// 设置路由
	router := gin.Default()

	// 静态文件服务
	router.Static("/static", "./web/static")
	router.StaticFile("/", "./web/static/index.html")

	// API路由
	apiGroup := router.Group("/api")
	{
		// 代理规则管理
		apiGroup.GET("/rules", handler.GetRules)
		apiGroup.GET("/rules/:id", handler.GetRule)
		apiGroup.POST("/rules", handler.CreateRule)
		apiGroup.PUT("/rules/:id", handler.UpdateRule)
		apiGroup.DELETE("/rules/:id", handler.DeleteRule)

		// 路由查询（供OpenResty调用）
		apiGroup.POST("/route", handler.Route)

		// 证书管理
		apiGroup.GET("/certificates", handler.GetCertificates)
		apiGroup.GET("/certificates/:id", handler.GetCertificate)
		apiGroup.POST("/certificates", handler.UploadCertificateNew)
		apiGroup.PUT("/certificates/:id/name", handler.UpdateCertificateName)
		apiGroup.DELETE("/certificates/:id", handler.DeleteCertificate)

		// 腾讯云证书管理（如果启用）
		if tencentSSL != nil {
			tencentGroup := apiGroup.Group("/certificates/tencent")
			tencentGroup.POST("/apply", handler.ApplyTencentCertificate)
			tencentGroup.GET("/:id/status", handler.CheckTencentCertificateStatus)
			tencentGroup.POST("/:id/download", handler.DownloadTencentCertificate)
			tencentGroup.POST("/:id/renew", handler.RenewTencentCertificate)
			tencentGroup.PUT("/:id/name", handler.UpdateCertificateName)
			tencentGroup.DELETE("/:id", handler.DeleteTencentCertificate)
		}

		// 系统管理
		apiGroup.POST("/nginx/reload", handler.ReloadNginx)
	}

	// 启动服务器
	addr := fmt.Sprintf("%s:%d", config.Server.Host, config.Server.Port)
	log.Printf("Server starting on %s", addr)
	log.Printf("Web interface: http://%s:%d", config.Server.Host, config.Server.Port)
	if err := handler.RegenerateAllConfigs(); err != nil {
		log.Printf("Warning: failed to reload nginx conf: %v", err)
	} else {
		log.Printf("Server reload config success")
	}

	// 优雅关闭处理
	go func() {
		if err := http.ListenAndServe(addr, router); err != nil {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	// 等待中断信号
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down server...")

	// 停止清理服务
	if cleanupService != nil {
		cleanupService.Stop()
	}

	log.Println("Server stopped")
}
