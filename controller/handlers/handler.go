package handlers

import (
	"github.com/go-playground/validator/v10"
	"github.com/gonzabosio/res-manager/config"
	database "github.com/gonzabosio/res-manager/model/db"
	"github.com/gonzabosio/res-manager/model/db/repository"
)

type Handler struct {
	Service *repository.DBService
	S3      *config.S3aws
}

func NewHandler() (*Handler, error) {
	h := new(Handler)
	// database
	db, err := database.NewDB()
	if err != nil {
		return nil, err
	}
	s := &repository.DBService{DB: db}
	h.Service = s
	// s3
	h.S3, err = config.NewS3Instance()
	if err != nil {
		return nil, err
	}
	return h, nil
}

var validate = validator.New(validator.WithRequiredStructEnabled())
