package repository

import (
	"fmt"
	"time"

	"github.com/gonzabosio/res-manager/model"
	"github.com/lib/pq"
)

type ResourceRepository interface {
	CreateResource(*model.Resource) error
	ReadResourcesBySectionID(int64) (*[]model.Resource, error)
	UpdateResource(*model.PatchResource) error
	DeleteResourceByID(int64) error
}

var _ ResourceRepository = (*DBService)(nil)

func (s *DBService) CreateResource(res *model.Resource) error {
	query := "INSERT INTO public.resource (title, content, url, images, last_edition_at, last_edition_by, section_id) VALUES($1, $2, $3, $4, $5, $6, $7) RETURNING id"
	now := time.Now()
	res.LastEditionAt = now
	if err := s.DB.QueryRow(query, res.Title, res.Content, res.URL, pq.Array(res.Images), time.Now(), res.LastEditionBy, res.SectionId).Scan(&res.Id); err != nil {
		return err
	}
	return nil
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
		if err := rows.Scan(&r.Id, &r.Title, &r.Content, &r.URL, pq.Array(&r.Images), &r.LastEditionAt, &r.LastEditionBy, &r.SectionId); err != nil {
			return nil, fmt.Errorf("failed reading rows: %v", err)
		}
		resources = append(resources, r)
	}
	return &resources, nil
}

func (s *DBService) UpdateResource(res *model.PatchResource) error {
	if err := s.DB.QueryRow("UPDATE public.resource SET title=$1, content=$2, url=$3, last_edition_at=$4, last_edition_by=$5 WHERE id=$6 RETURNING title, content, url, last_edition_at, section_id",
		res.Title, res.Content, res.URL, time.Now(), res.LastEditionBy, res.Id).Scan(&res.Title, &res.Content, &res.URL, &res.LastEditionAt, &res.SectionId); err != nil {
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
