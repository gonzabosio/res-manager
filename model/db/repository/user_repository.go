package repository

import (
	"database/sql"
	"fmt"

	"github.com/gonzabosio/res-manager/model"
)

type UserRepository interface {
	InsertOrGetUser(*model.User) (bool, error)
	UpdateUser(*model.PatchUser) error
	DeleteUserByID(int64) error
}

var _ UserRepository = (*DBService)(nil)

func (s *DBService) InsertOrGetUser(user *model.User) (wasInserted bool, err error) {
	queryDuplEmail := "SELECT id, username FROM public.user WHERE email=$1"
	err = s.DB.QueryRow(queryDuplEmail, user.Email).Scan(&user.Id, &user.Username)
	if err == sql.ErrNoRows {
		query := "INSERT INTO public.user(username, email) VALUES($1, $2) RETURNING id"
		err = s.DB.QueryRow(query, user.Username, user.Email).Scan(&user.Id)
		if err != nil {
			return false, fmt.Errorf("failed user creation: %v", err)
		}
		return true, nil
	} else {
		return false, nil
	}
}

func (s *DBService) ReadUsers() (*[]model.User, error) {
	var users []model.User
	query := "SELECT * FROM public.user"
	rows, err := s.DB.Query(query)
	if err != nil {
		return nil, err
	}
	for rows.Next() {
		var r model.User
		if err := rows.Scan(&r.Id, &r.Username, &r.Email); err != nil {
			return nil, fmt.Errorf("failed reading rows: %v", err)
		}
		users = append(users, r)
	}
	return &users, nil
}

func (s *DBService) UpdateUser(user *model.PatchUser) error {
	row := s.DB.QueryRow("UPDATE public.user SET username=$1 WHERE id=$2 RETURNING email", user.Username, user.Id)
	if err := row.Scan(&user.Email); err != nil {
		return err
	}
	return nil
}

func (s *DBService) DeleteUserByID(userId int64) error {
	_, err := s.DB.Exec("DELETE FROM public.user WHERE id=$1", userId)
	if err != nil {
		return err
	}
	s.DB.Query("DELETE FROM public.team WHERE id")
	if err != nil {

	}
	return nil
}
