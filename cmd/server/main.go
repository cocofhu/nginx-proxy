package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"nginx-proxy/internal/api"
	"nginx-proxy/internal/core"
	"nginx-proxy/internal/db"

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

	// 初始化数据库 - 使用纯 Go SQLite 驱动
	database, err := gorm.Open(sqlite.Dialector{
		DriverName: "sqlite",
		DSN:        config.Database.DSN,
	}, &gorm.Config{})
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	// 自动迁移数据库
	if err := database.AutoMigrate(&db.Rule{}, &db.Certificate{}); err != nil {
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
		tencentSSL = core.NewTencentSSLService(&config.TencentCloud, database, config.SSL.CertDir)
		log.Println("Tencent Cloud SSL service initialized")
	} else {
		log.Println("Tencent Cloud SSL service not configured")
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
	if err := http.ListenAndServe(addr, router); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
