package dto

type RegisterRequest struct {
	Login string `json:"login"`
	Password string `json:"password"`
}

type LoginRequest struct {
	Login string `json:"login"`
	Password string `json:"password"`
}

type UserResponse struct {
    ID    int    `json:"id"`
    Login string `json:"login"`
}