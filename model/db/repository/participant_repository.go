package repository

import (
	"fmt"

	"github.com/gonzabosio/res-manager/model"
)

type ParticipantRepository interface {
	InsertParticipant(*model.Participant) (int64, error)
	ReadParticipants(int64) (*[]model.ParticipantResp, error)
	DeleteParticipantByID(int64) error
}

var _ ParticipantRepository = (*DBService)(nil)

func (s *DBService) InsertParticipant(participant *model.Participant) (int64, error) {
	var insertedID int64
	query := "INSERT INTO public.participant(user_id, team_id) VALUES($1, $2) RETURNING id"
	err := s.DB.QueryRow(query, participant.UserId, participant.TeamId).Scan(&insertedID)
	if err != nil {
		return 0, fmt.Errorf("failed participant creation: %v", err)
	}
	return insertedID, nil
}

func (s *DBService) ReadParticipants(teamId int64) (*[]model.ParticipantResp, error) {
	var participants []model.ParticipantResp
	query := `SELECT p.id, u.username FROM participant p JOIN "user" u ON p.user_id = u.id WHERE p.team_id = $1`
	rows, err := s.DB.Query(query, teamId)
	if err != nil {
		return nil, fmt.Errorf("failed getting participants: %v", err)
	}
	for rows.Next() {
		var r model.ParticipantResp
		if err := rows.Scan(&r.Id, &r.Username); err != nil {
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
