package model

type User struct {
	Id       int64
	Username string `json:"username,omitempty" validate:"required"`
	Password string `json:"password" validate:"required"`
	Email    string `json:"email" validate:"required,email"`
}
