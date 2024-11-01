package repository

import (
	"database/sql"
	"fmt"
	"log"

	"github.com/gonzabosio/res-manager/model"
)

type ParticipantRepository interface {
	RegisterParticipant(*model.Participant) (bool, error)
	ReadParticipants(int64) (*[]model.ParticipantsResp, error)
	DeleteParticipantByID(int64, int64) error
}

var _ ParticipantRepository = (*DBService)(nil)

func (s *DBService) RegisterParticipant(participant *model.Participant) (wasInserted bool, err error) {
	if err := s.DB.QueryRow("SELECT id, admin FROM participant WHERE user_id=$1 AND team_id=$2",
		participant.UserId, participant.TeamId,
	).Scan(&participant.Id, &participant.Admin); err != nil {
		if err == sql.ErrNoRows {
			query := "INSERT INTO participant(admin, user_id, team_id) VALUES($1, $2, $3) RETURNING id"
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
	query := `SELECT p.id, p.admin, p.user_id, u.username FROM participant p JOIN "user" u ON p.user_id = u.id WHERE p.team_id = $1`
	rows, err := s.DB.Query(query, teamId)
	if err != nil {
		return nil, fmt.Errorf("failed getting participants: %v", err)
	}
	for rows.Next() {
		var r model.ParticipantsResp
		if err := rows.Scan(&r.Id, &r.Admin, &r.UserId, &r.Username); err != nil {
			return nil, fmt.Errorf("failed reading rows: %v", err)
		}
		participants = append(participants, r)
	}
	return &participants, nil
}

func (s *DBService) AssignAdminRole(pId int64) error {
	if _, err := s.DB.Exec("UPDATE participant SET admin=$1 WHERE id=$2", true, pId); err != nil {
		return err
	}
	return nil
}

func (s *DBService) DeleteParticipantByID(teamId int64, pId int64) error {
	log.Println(teamId, pId)
	rows, err := s.DB.Query("SELECT id, admin FROM participant WHERE team_id=$1", teamId)
	if err != nil {
		return err
	}
	type participantObj struct {
		id    int64
		admin bool
	}
	var isAdmin bool
	var pSli []participantObj
	for rows.Next() {
		var p participantObj
		rows.Scan(&p.id, &p.admin)
		if p.id == pId && p.admin {
			isAdmin = true
		}
		pSli = append(pSli, p)
	}
	if len(pSli) == 1 {
		log.Println("DELETE TEAM")
		if _, err := s.DB.Exec("DELETE FROM team WHERE id=$1", teamId); err != nil {
			return err
		}
	} else {
		addAdmin := true
		idx := 0
		if isAdmin {
			for i, p := range pSli {
				if p.id != pId && p.admin {
					addAdmin = false
					break
				} else {
					fmt.Println("Participant Slice len", len(pSli))
					fmt.Println("Participant Slice i", i)
					if len(pSli) == i {
						idx = i - 1
					}
				}
				fmt.Println("Member: ", p)
			}
		}
		if addAdmin {
			fmt.Println("Member to assign admin is: ", idx)
			if _, err := s.DB.Exec("UPDATE participant SET admin=$1 WHERE id=$2", true, pSli[idx].id); err != nil {
				return err
			}
			if _, err := s.DB.Exec("DELETE FROM participant WHERE id=$1", pId); err != nil {
				return err
			}
		} else {
			if _, err := s.DB.Exec("DELETE FROM participant WHERE id=$1", pId); err != nil {
				return err
			}
		}
	}
	return nil
}
