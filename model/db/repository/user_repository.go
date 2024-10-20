package repository

import (
	"database/sql"
	"fmt"

	"github.com/gonzabosio/res-manager/model"
)

type UserRepository interface {
	InsertUser(*model.User) (int64, error)
	VerifyUser(*model.User) error
	UpdateUser(*model.User) error
	DeleteUserByID(int64) error
}

var _ UserRepository = (*DBService)(nil)

func (s *DBService) InsertUser(user *model.User) (int64, error) {
	var username string
	queryDupl := "SELECT username FROM public.user WHERE username=$1"
	err := s.DB.QueryRow(queryDupl, user.Username).Scan(&username)
	if err == sql.ErrNoRows {
		var insertedID int64
		query := "INSERT INTO public.user(username, password, email) VALUES($1, $2, $3) RETURNING id"
		err = s.DB.QueryRow(query, user.Username, user.Password, user.Email).Scan(&insertedID)
		if err != nil {
			return 0, fmt.Errorf("failed user creation: %v", err)
		}
		return insertedID, nil
	} else {
		return 0, fmt.Errorf("username already exists")
	}

}
func (s *DBService) VerifyUser(user *model.User) error {
	query := `SELECT * FROM public.user WHERE username=$1`
	err := s.DB.QueryRow(query, user.Username).Scan(&user.Id, &user.Username, &user.Password, &user.Email)
	if err != nil {
		return err
	}
	return nil
}

func (s *DBService) UpdateUser(user *model.User) error {
	row := s.DB.QueryRow("UPDATE public.user SET username=$1, password=$2 WHERE id=$3 RETURNING username,password,email", user.Username, user.Password, user.Id)
	err := row.Scan(&user.Username, &user.Password, &user.Email)
	if err != nil {
		return err
	}
	return nil
}

func (s *DBService) DeleteUserByID(userId int64) error {
	_, err := s.DB.Exec("DELETE FROM public.user WHERE id=$1", userId)
	if err != nil {
		return err
	}
	return nil
}
