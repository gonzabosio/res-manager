package repository

import (
	"fmt"

	"github.com/gonzabosio/res-manager/model"
)

type SectionRepository interface {
	CreateSection(*model.Section) (int64, error)
	ReadSectionsByProjectID(int64) (*[]model.Section, error)
	UpdateSection(*model.Section) error
	DeleteSectionByID(int64) error
}

var _ SectionRepository = (*DBService)(nil)

func (s *DBService) CreateSection(section *model.Section) (int64, error) {
	var insertedID int64
	query := "INSERT INTO public.section(title, project_id) VALUES($1, $2) RETURNING id"
	if err := s.DB.QueryRow(query, section.Title, section.ProjectID).Scan(&insertedID); err != nil {
		return 0, err
	}
	return insertedID, nil
}

func (s *DBService) ReadSectionsByProjectID(projectId int64) (*[]model.Section, error) {
	var sections []model.Section
	query := "SELECT * FROM public.section WHERE project_id=$1"
	rows, err := s.DB.Query(query, projectId)
	if err != nil {
		return nil, err
	}
	for rows.Next() {
		var r model.Section
		if err := rows.Scan(&r.Id, &r.Title, &r.ProjectID); err != nil {
			return nil, fmt.Errorf("failed reading rows: %v", err)
		}
		sections = append(sections, r)
	}
	return &sections, nil
}

func (s *DBService) UpdateSection(section *model.Section) error {
	if err := s.DB.QueryRow("UPDATE public.section SET title=$1 WHERE id=$2 RETURNING title", section.Title, section.Id).Scan(&section.Title); err != nil {
		return err
	}
	return nil
}

func (s *DBService) DeleteSectionByID(sectionId int64) error {
	if _, err := s.DB.Exec("DELETE FROM public.section WHERE id=$1", sectionId); err != nil {
		return err
	}
	return nil
}
