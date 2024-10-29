package model

type Participant struct {
	Id     int64 `json:"id,omitempty"`
	Admin  bool  `json:"admin" validate:"required"`
	UserId int64 `json:"user_id" validate:"required"`
	TeamId int64 `json:"team_id" validate:"required"`
}

type ParticipantsResp struct {
	Id       int64  `json:"id"`
	Username string `json:"username"`
	Admin    bool   `json:"admin"`
}
