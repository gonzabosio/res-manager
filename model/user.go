package model

type User struct {
	Id       int64  `json:"id,omitempty"`
	Username string `json:"username" validate:"required,max=50"`
	Email    string `json:"email" validate:"required,email"`
}

type GoogleUser struct {
	Id            string `json:"id"`
	Email         string `json:"email"`
	VerifiedEmail bool   `json:"verified_email"`
	Name          string `json:"name"`
	GivenName     string `json:"given_name"`
	FamilyName    string `json:"family_name"`
	Picture       string `json:"picture"`
}

type PatchUser struct {
	Id       int64  `json:"id" validate:"required"`
	Username string `json:"username" validate:"required"`
	Email    string `json:"email,omitempty"`
}
