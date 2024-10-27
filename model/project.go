package model

type Project struct {
	Id      int64  `json:"id,omitempty"`
	Name    string `json:"name" validate:"required"`
	Details string `json:"details" validate:"required"`
	TeamId  int64  `json:"team_id" validate:"required"`
}

type PatchProject struct {
	Id      int64  `json:"id" validate:"required"`
	Name    string `json:"name"`
	Details string `json:"details"`
	TeamId  int64  `json:"team_id"`
}
