package model

type Team struct {
	Id       int8   `json:"id,omitempty"`
	Name     string `json:"name"`
	Password string `json:"password"`
}
