package main

import (
	"fmt"
	"log"
	"os"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"nginx-proxy/internal/core"
)

func main() {

	// 加载配置
	config, err := core.LoadConfig("config.json")
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// 检查腾讯云配置
	if config.TencentCloud.SecretId == "" || config.TencentCloud.SecretKey == "" {
		log.Fatalf("腾讯云配置不完整，请检查 config.json 中的 tencent_cloud 配置")
	}

	// 初始化数据库（仍需要用于SSL服务初始化）
	database, err := gorm.Open(sqlite.Open(config.Database.DSN), &gorm.Config{})
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	// 初始化腾讯云SSL服务
	tencentSSL := core.NewTencentSSLService(&config.TencentCloud, database, config.SSL.CertDir)
	if tencentSSL == nil {
		log.Fatalf("Failed to initialize Tencent Cloud SSL service")
	}

	// 直接从腾讯云API获取所有证书列表
	fmt.Println("正在从腾讯云获取证书列表...")
	certificates, err := tencentSSL.GetAllTencentCertificates()
	if err != nil {
		log.Fatalf("Failed to get certificates from Tencent Cloud: %v", err)
	}

	if len(certificates) == 0 {
		fmt.Println("腾讯云中没有找到任何证书")
		return
	}

	fmt.Printf("从腾讯云找到 %d 个证书，开始取消吊销...\n", len(certificates))

	successCount := 0
	failCount := 0

	// 逐个吊销证书
	for i, cert := range certificates {
		fmt.Printf("[%d/%d] 正在吊销证书: %s (域名: %s, ID: %s, 状态: %s)\n",
			i+1, len(certificates), cert.Alias, cert.Domain, cert.CertificateID, cert.Status)

		// 跳过已经吊销的证书
		if cert.Status != "证书吊销中" {
			fmt.Printf("⚠️  证书 %s 不是是吊销中状态，跳过\n", cert.CertificateID)
			continue
		}

		if err := tencentSSL.CancelRevokeTencentCertificateByID(cert.CertificateID); err != nil {
			log.Printf("取消吊销证书失败 %s: %v", cert.CertificateID, err)
			failCount++
		} else {
			fmt.Printf("✓ 成功取消吊销证书: %s\n", cert.CertificateID)
			successCount++
		}
	}

	// 输出统计结果
	fmt.Printf("\n=== 取消吊销完成 ===\n")
	fmt.Printf("总计: %d 个证书\n", len(certificates))
	fmt.Printf("成功: %d 个\n", successCount)
	fmt.Printf("失败: %d 个\n", failCount)

	if failCount > 0 {
		fmt.Printf("\n注意：有 %d 个证书取消吊销失败，请检查日志了解详情\n", failCount)
		os.Exit(1)
	} else {
		fmt.Println("\n所有证书已成功取消吊销！")
	}
}
