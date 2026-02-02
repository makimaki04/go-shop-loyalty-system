package dto

type RegisterRequest struct {
	Login string `json:"login"`
	Password string `json:"password"`
}

type LoginRequest struct {
	Login string `json:"login"`
	Password string `json:"password"`
}

type SignUpResponse struct {
	ID       int64    `json:"id"`
	JWTToken string `json:"jwt-token"`
}

type LoginResponse struct {
	User     UserResponse `json:"user"`
	JWTToken string       `json:"jwt-token"`
}

type UserResponse struct {
	ID    int64    `json:"id"`
	Login string `json:"login"`
}