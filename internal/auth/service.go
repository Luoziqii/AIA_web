package auth

import (
	"errors"
	"sync"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type Service struct {
	jwtSecret []byte
	secretKey []byte
	tokenTTL  time.Duration

	mu        sync.RWMutex
	latestIAT int64 // 记录最后一次成功签发 Token 的时间，用于顶号
}

func NewService(jwtSecret, dynamicSecretB64 string, tokenTTL time.Duration) *Service {
	if jwtSecret == "" || dynamicSecretB64 == "" {
		panic("致命错误: JWT_SECRET 或 DYNAMIC_SECRET 不能为空")
	}

	secretKey, err := ParseSecretKey(dynamicSecretB64)
	if err != nil {
		panic("致命错误: DYNAMIC_SECRET 解析失败, 必须是有效的 Base64 字符串")
	}

	return &Service{
		jwtSecret: []byte(jwtSecret),
		secretKey: secretKey,
		tokenTTL:  tokenTTL,
	}
}

func (s *Service) Login(password string) (string, error) {
	// 严格时间同步校验强动态密码
	if !VerifyStrongDynamicPassword(s.secretKey, password, time.Now()) {
		return "", errors.New("认证失败或密码已过期")
	}

	now := time.Now()

	// 更新最新签发时间，实现顶号逻辑
	s.mu.Lock()
	s.latestIAT = now.Unix()
	s.mu.Unlock()

	claims := jwt.MapClaims{
		"role": RoleSuperAdmin,
		"iat":  now.Unix(),
		"exp":  now.Add(s.tokenTTL).Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(s.jwtSecret)
}

func (s *Service) ParseToken(tokenString string) (*Claims, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if token.Method != jwt.SigningMethodHS256 {
			return nil, errors.New("签名算法异常")
		}
		return s.jwtSecret, nil
	})
	if err != nil || !token.Valid {
		return nil, errors.New("无效的 Token")
	}

	mapClaims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return nil, errors.New("解析 Claims 失败")
	}

	role, _ := mapClaims["role"].(string)
	iatFloat, _ := mapClaims["iat"].(float64)
	expFloat, _ := mapClaims["exp"].(float64)
	iat := int64(iatFloat)

	// 单设备登录限制校验
	s.mu.RLock()
	latest := s.latestIAT
	s.mu.RUnlock()

	if iat < latest {
		return nil, errors.New("账号已在其他终端登录，当前 Token 已失效")
	}

	return &Claims{Role: role, Iat: iat, Exp: int64(expFloat)}, nil
}
