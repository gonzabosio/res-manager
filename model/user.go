package model

type User struct {
	Id       int64
	Username string `json:"username,omitempty" validate:"required,max=30"`
	Password string `json:"password" validate:"required,min=4,max=100"`
	Email    string `json:"email" validate:"required,email"`
}
