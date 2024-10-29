package repository

import (
	"database/sql"
	"fmt"

	"github.com/gonzabosio/res-manager/model"
)

type ParticipantRepository interface {
	RegisterParticipant(*model.Participant) (bool, error)
	ReadParticipants(int64) (*[]model.ParticipantsResp, error)
	DeleteParticipantByID(int64) error
}

var _ ParticipantRepository = (*DBService)(nil)

func (s *DBService) RegisterParticipant(participant *model.Participant) (wasInserted bool, err error) {
	if err := s.DB.QueryRow("SELECT id, admin FROM public.participant WHERE user_id=$1 AND team_id=$2",
		participant.UserId, participant.TeamId,
	).Scan(&participant.Id, participant.Admin); err != nil {
		if err == sql.ErrNoRows {
			query := "INSERT INTO public.participant(admin, user_id, team_id) VALUES($1, $2, $3) RETURNING id"
			if err := s.DB.QueryRow(query, participant.Admin, participant.UserId, participant.TeamId).Scan(&participant.Id); err != nil {
				return false, fmt.Errorf("failed participant creation: %v", err)
			}
			return true, nil
		}
	}
	return false, nil
}

func (s *DBService) ReadParticipants(teamId int64) (*[]model.ParticipantsResp, error) {
	var participants []model.ParticipantsResp
	query := `SELECT p.id, p.admin, u.username FROM participant p JOIN "user" u ON p.user_id = u.id WHERE p.team_id = $1`
	rows, err := s.DB.Query(query, teamId)
	if err != nil {
		return nil, fmt.Errorf("failed getting participants: %v", err)
	}
	for rows.Next() {
		var r model.ParticipantsResp
		if err := rows.Scan(&r.Id, &r.Admin, &r.Username); err != nil {
			return nil, fmt.Errorf("failed reading rows: %v", err)
		}
		participants = append(participants, r)
	}
	return &participants, nil
}

func (s *DBService) DeleteParticipantByID(userId int64) error {
	if _, err := s.DB.Exec("DELETE FROM participant WHERE id=$1", userId); err != nil {
		return err
	}
	return nil
}
