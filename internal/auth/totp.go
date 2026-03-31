package auth

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/binary"
	"time"
)

const TotpPeriod = 30 // 30秒绝对同步窗口

// ParseSecretKey 解析 Base64 格式的对称密钥
func ParseSecretKey(b64 string) ([]byte, error) {
	return base64.StdEncoding.DecodeString(b64)
}

// GenerateStrongDynamicPassword 生成 16 字符的强动态密码
func GenerateStrongDynamicPassword(key []byte, t time.Time) string {
	counter := uint64(t.Unix() / TotpPeriod)

	buf := make([]byte, 8)
	binary.BigEndian.PutUint64(buf, counter)

	mac := hmac.New(sha256.New, key)
	mac.Write(buf)
	sum := mac.Sum(nil)

	// 取 HMAC-SHA256 结果的前 12 个字节
	// 12 bytes 经过 Base64 编码后，正好是 16 个字符，且没有冗余的 '=' 填充
	return base64.StdEncoding.EncodeToString(sum[:12])
}

// VerifyStrongDynamicPassword 严格校验动态码（零容差）
func VerifyStrongDynamicPassword(key []byte, code string, t time.Time) bool {
	expectedCode := GenerateStrongDynamicPassword(key, t)
	// 使用 hmac.Equal 防止时序攻击
	return hmac.Equal([]byte(expectedCode), []byte(code))
}
