package model

type Team struct {
	Id       int64  `json:"id,omitempty"`
	Name     string `json:"name" validate:"required,max=30"`
	Password string `json:"password" validate:"required,min=4"`
}

type PatchTeam struct {
	Id       int64  `json:"id" validate:"required"`
	Name     string `json:"name"`
	Password string `json:"password"`
}

type TeamView struct {
	Id                 int64  `json:"id,omitempty"`
	Name               string `json:"name"`
	ParticipantsNumber int    `json:"participants"`
}
