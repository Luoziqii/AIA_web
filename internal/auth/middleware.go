// internal/auth/middleware.go
package auth

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

const (
	// 只保留 RoleKey，去除 UserID 和 Username
	ContextRoleKey = "auth_role"
)

type Middleware struct {
	svc *Service
}

func NewMiddleware(svc *Service) *Middleware {
	return &Middleware{svc: svc}
}

// RequireAdmin 统一的权限拦截器 (验证 Token + 验证终极管理员身份)
// 之前在 main.go 中你使用的是 authorized.Use(authMiddleware.RequireAdmin())，
// 所以保留这个方法名，你的 main.go 就不需要做任何修改！
func (m *Middleware) RequireAdmin() gin.HandlerFunc {
	return func(c *gin.Context) {
		token := extractBearerToken(c.GetHeader("Authorization"))
		if token == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, ErrorResponse{Error: "未提供 Bearer Token"})
			return
		}

		// 解析并验证 Token（包含我们刚加的“防顶号 / Token过期”校验）
		claims, err := m.svc.ParseToken(token)
		if err != nil {
			// 直接返回 Service 层抛出的具体错误（如："账号已在其他终端登录，当前 Token 已失效"）
			c.AbortWithStatusJSON(http.StatusUnauthorized, ErrorResponse{Error: err.Error()})
			return
		}

		// 终极管理员角色校验 (RoleSuperAdmin 是在 model.go 中定义的常量)
		if claims.Role != RoleSuperAdmin {
			c.AbortWithStatusJSON(http.StatusForbidden, ErrorResponse{Error: "非法的身份标识，权限不足"})
			return
		}

		// 验证全部通过，记录当前角色并放行请求
		c.Set(ContextRoleKey, claims.Role)
		c.Next()
	}
}

// extractBearerToken 提取 HTTP 头的 Token (不需要修改)
func extractBearerToken(header string) string {
	const prefix = "Bearer "
	if len(header) < len(prefix) || !strings.EqualFold(header[:len(prefix)], prefix) {
		return ""
	}
	return strings.TrimSpace(header[len(prefix):])
}
