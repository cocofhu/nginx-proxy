package main

import (
	"encoding/json"
	"flag"
	"log"
	"os"

	"nginx-proxy/internal/api"
	"nginx-proxy/internal/core"
	"nginx-proxy/internal/db"

	"github.com/gin-gonic/gin"
)

type Config struct {
	Port         string `json:"port"`
	NginxPath    string `json:"nginx_path"`
	ConfigDir    string `json:"config_dir"`
	CertDir      string `json:"cert_dir"`
	DatabasePath string `json:"database_path"`
	TemplateDir  string `json:"template_dir"`
}

func loadConfig(configPath string) (*Config, error) {
	file, err := os.Open(configPath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var config Config
	decoder := json.NewDecoder(file)
	err = decoder.Decode(&config)
	return &config, err
}

func main() {
	// 定义命令行参数
	configFile := flag.String("config", "config.json", "配置文件路径")
	flag.Parse()

	// 加载配置
	config, err := loadConfig(*configFile)
	if err != nil {
		log.Fatalf("Failed to load config from %s: %v", *configFile, err)
	}

	// 初始化数据库
	database, err := db.InitDB(config.DatabasePath)
	if err != nil {
		log.Fatal("Failed to initialize database:", err)
	}

	// 初始化核心服务
	generator := core.NewGenerator(config.TemplateDir, config.ConfigDir)
	nginxManager := core.NewNginxManager(config.NginxPath)

	// 初始化 API 处理器
	handler := api.NewHandler(database, generator, nginxManager, config.CertDir)

	// 设置路由
	r := gin.Default()

	// API 路由组
	apiGroup := r.Group("/api")
	{
		apiGroup.GET("/rules", handler.GetRules)
		apiGroup.GET("/rules/:id", handler.GetRule)
		apiGroup.POST("/rules", handler.CreateRule)
		apiGroup.PUT("/rules/:id", handler.UpdateRule)
		apiGroup.DELETE("/rules/:id", handler.DeleteRule)
		apiGroup.POST("/reload", handler.ReloadNginx)
		apiGroup.POST("/certificates", handler.UploadCertificate)
	}

	// 启动时重新生成所有配置
	if err := handler.RegenerateAllConfigs(); err != nil {
		log.Printf("Warning: Failed to regenerate configs on startup: %v", err)
	}

	log.Printf("Server starting on port %s", config.Port)
	log.Fatal(r.Run(":" + config.Port))
}
