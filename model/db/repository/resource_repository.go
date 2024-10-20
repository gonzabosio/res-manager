package repository

import (
	"fmt"

	"github.com/gonzabosio/res-manager/model"
)

type ResourceRepository interface {
	CreateResource(*model.Resource) (int64, error)
	ReadResourcesBySectionID(int64) (*[]model.Resource, error)
	UpdateResource(*model.Resource) error
	DeleteResourceByID(int64) error
}

var _ ResourceRepository = (*DBService)(nil)

func (s *DBService) CreateResource(res *model.Resource) (int64, error) {
	var insertedID int64
	query := "INSERT INTO public.resource(title, content, url, section_id) VALUES($1, $2, $3, $4) RETURNING id"
	if err := s.DB.QueryRow(query, res.Title, res.Content, res.URL, res.SectionId).Scan(&insertedID); err != nil {
		return 0, err
	}
	return insertedID, nil
}

func (s *DBService) ReadResourcesBySectionID(sectionId int64) (*[]model.Resource, error) {
	var resources []model.Resource
	query := "SELECT * FROM public.resource WHERE section_id=$1"
	rows, err := s.DB.Query(query, sectionId)
	if err != nil {
		return nil, err
	}
	for rows.Next() {
		var r model.Resource
		if err := rows.Scan(&r.Id, &r.Title, &r.Content, &r.URL, &r.SectionId); err != nil {
			return nil, fmt.Errorf("failed reading rows: %v", err)
		}
		resources = append(resources, r)
	}
	return &resources, nil
}

func (s *DBService) UpdateResource(res *model.Resource) error {
	if err := s.DB.QueryRow("UPDATE public.resource SET title=$1, content=$2, url=$3 WHERE id=$4 RETURNING title, content, url",
		res.Title, res.Content, res.URL, res.Id).Scan(&res.Title, &res.Content, &res.URL); err != nil {
		return err
	}
	return nil
}

func (s *DBService) DeleteResourceByID(resourceId int64) error {
	if _, err := s.DB.Exec("DELETE FROM resource WHERE id=$1", resourceId); err != nil {
		return err
	}
	return nil
}
