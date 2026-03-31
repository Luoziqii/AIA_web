package auth

const (
	RoleSuperAdmin = "super_admin"
)

type LoginRequest struct {
	Password string `json:"password" binding:"required" example:"x6+xJ/c5fsWaVyxu"`
}

type LoginResponse struct {
	Token string `json:"token"`
}

type ErrorResponse struct {
	Error string `json:"error"`
}

type Claims struct {
	Role string `json:"role"`
	Iat  int64  `json:"iat"` // 用于校验是否被顶号
	Exp  int64  `json:"exp"`
}
