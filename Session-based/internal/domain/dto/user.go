package dto

type LoginRequest struct {
	Email    string `json:"email" binding:"email,required"`
	Password string `json:"password" binding:"min=8,required"`
}
type RegisterRequest struct {
	UserName      string `json:"username" binding:"min=3,required"`
	Email         string `json:"email" binding:"email,required"`
	Password      string `json:"password" binding:"min=8,required"`
	RetryPassword string `json:"retry_password" binding:"min=8,required"`
}
