package auth

// RegisterRequest represents request to register a new user
type RegisterRequest struct {
	Name     string `json:"name" validate:"required,min=2,max=100"`
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=8"`
}

// LoginRequest represents request to login
type LoginRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required"`
}

// LoginResponse represents response after successful login
type LoginResponse struct {
	Token string      `json:"token"`
	User  interface{} `json:"user"`
}
