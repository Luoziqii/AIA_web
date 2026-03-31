// cmd/totp/main.go
package main

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/binary"
	"fmt"
	"os"
	"time"

	"github.com/joho/godotenv"
)

// GenerateStrongDynamicPassword 与服务端完全一致的算法
func GenerateStrongDynamicPassword(key []byte, t time.Time) string {
	counter := uint64(t.Unix() / 30)
	buf := make([]byte, 8)
	binary.BigEndian.PutUint64(buf, counter)

	mac := hmac.New(sha256.New, key)
	mac.Write(buf)
	sum := mac.Sum(nil)

	return base64.StdEncoding.EncodeToString(sum[:12])
}

func main() {
	// 1. 自动从项目根目录加载 .env 文件（如果存在的话）
	_ = godotenv.Load()

	// 2. 现在的 os.Getenv 就能成功拿到 .env 里的值了
	secretB64 := os.Getenv("DYNAMIC_SECRET")

	if secretB64 == "" {
		fmt.Println("未检测到 DYNAMIC_SECRET 环境变量。")
		fmt.Println("正在为您生成一个新的 32 字节高熵安全密钥...")

		key := make([]byte, 32)
		rand.Read(key)
		newSecret := base64.StdEncoding.EncodeToString(key)

		// 顺便把这里的提示也优化一下，兼容 Windows
		fmt.Printf("\n请将以下密钥配置到服务器的 .env 文件中:\n\nDYNAMIC_SECRET=\"%s\"\n\n", newSecret)
		fmt.Println("配置好 .env 后，请重新运行此程序获取动态密码。")
		return
	}

	key, err := base64.StdEncoding.DecodeString(secretB64)
	if err != nil {
		fmt.Printf("密钥解析失败: %v\n", err)
		return
	}

	fmt.Println("========================================")
	fmt.Println(" 终极动态强密码生成器 (30秒同步)")
	fmt.Println("========================================")

	for {
		now := time.Now()
		password := GenerateStrongDynamicPassword(key, now)
		remain := 30 - (now.Unix() % 30)

		fmt.Printf("\r当前密码: [\033[36;1m%s\033[0m] (有效期剩余: %02d 秒)   ", password, remain)
		time.Sleep(1 * time.Second)
	}
}
