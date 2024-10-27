package repository

import (
	"encoding/json"
	"fmt"

	"github.com/gonzabosio/gobo-patcher"
	"github.com/gonzabosio/res-manager/model"
)

type ProjectRepository interface {
	CreateProject(*model.Project) (int64, error)
	ReadProject(int64) (*model.Project, error)
	ReadProjectsByTeamID(int64) (*[]model.Project, error)
	UpdateProject(*model.PatchProject) error
	DeleteProjectByID(int64) error
}

var _ ProjectRepository = (*DBService)(nil)

func (s *DBService) CreateProject(proj *model.Project) (int64, error) {
	var insertedID int64
	query := "INSERT INTO public.project(name, details, team_id) VALUES($1, $2, $3) RETURNING id"
	err := s.DB.QueryRow(query, proj.Name, proj.Details, proj.TeamId).Scan(&insertedID)
	if err != nil {
		return 0, err
	}
	return insertedID, nil
}

func (s *DBService) ReadProject(projectId int64) (*model.Project, error) {
	var project model.Project
	query := "SELECT * FROM public.project WHERE id=$1"
	row := s.DB.QueryRow(query, projectId)
	row.Scan(&project.Name, &project.Details, &project.TeamId)
	return &project, nil
}

func (s *DBService) ReadProjectsByTeamID(teamId int64) (*[]model.Project, error) {
	var projs []model.Project
	query := "SELECT * FROM public.project WHERE team_id=$1"
	rows, err := s.DB.Query(query, teamId)
	if err != nil {
		return nil, err
	}
	for rows.Next() {
		var r model.Project
		if err := rows.Scan(&r.Id, &r.Name, &r.Details, &r.TeamId); err != nil {
			return nil, fmt.Errorf("failed reading rows: %v", err)
		}
		projs = append(projs, r)
	}
	return &projs, nil
}

func (s *DBService) UpdateProject(project *model.PatchProject) error {
	origProject := model.Project{}
	s.DB.QueryRow("SELECT * FROM public.project WHERE id=$1", project.Id).
		Scan(&origProject.Id, &origProject.Name, &origProject.Details, &origProject.TeamId)
	orB, err := json.Marshal(origProject)
	if err != nil {
		return fmt.Errorf("could not parse project struct to json: %v", err)
	}
	newB, err := json.Marshal(project)
	if err != nil {
		return fmt.Errorf("could not parse project struct to json: %v", err)
	}
	query, err := gobo.PatchWithQuery(orB, newB, "public.project", "id", true, nil)
	if err != nil {
		return err
	}
	q := fmt.Sprintf("%v RETURNING name,details", query)
	if err = s.DB.QueryRow(q).Scan(&project.Name, &project.Details); err != nil {
		return err
	}
	return nil
}

func (s *DBService) DeleteProjectByID(projectId int64) error {
	_, err := s.DB.Exec("DELETE FROM public.project WHERE id=$1", projectId)
	if err != nil {
		return err
	}
	return nil
}
