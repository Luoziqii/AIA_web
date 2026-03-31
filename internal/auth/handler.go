// internal/auth/handler.go
package auth

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	svc *Service
}

func NewHandler(svc *Service) *Handler {
	return &Handler{svc: svc}
}

// LoginAdmin 终极管理员登录
// @Summary 终极管理员登录 (仅依赖高熵动态口令)
// @Tags Auth
// @Accept json
// @Produce json
// @Param body body LoginRequest true "登录参数 (将生成的 16 位强动态口令填入 password 字段)"
// @Success 200 {object} LoginResponse
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Router /auth/login [post]
func (h *Handler) LoginAdmin(c *gin.Context) {
	var req LoginRequest

	// 绑定 JSON 数据到我们精简后的 LoginRequest 模型
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "必须提供有效的动态口令"})
		return
	}

	// 调用重构后的 Service 层，仅传入生成的强动态口令
	token, err := h.svc.Login(req.Password)
	if err != nil {
		// 错误提示保持模糊，不给爆破者任何线索
		c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "认证失败或口令已过期"})
		return
	}

	c.JSON(http.StatusOK, LoginResponse{Token: token})
}
