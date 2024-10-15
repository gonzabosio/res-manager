package model

type Team struct {
	Id       int64  `json:"id,omitempty"`
	Name     string `json:"name" validate:"required"`
	Password string `json:"password" validate:"required"`
}
