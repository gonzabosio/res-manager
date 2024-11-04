package repository

import (
	"database/sql"
	"fmt"
	"log"
	"path"
	"time"

	"github.com/gonzabosio/res-manager/model"
	"github.com/lib/pq"
)

type ResourceRepository interface {
	CreateResource(*model.Resource) error
	ReadResourcesBySectionID(int64) (*[]model.Resource, error)
	UpdateResource(*model.PatchResource) error
	DeleteResourceByID(int64) error
	SaveImageURL(string, int64) error
	GetImagesByResourceID(int64) ([]string, error)
	DeleteImageByResourceID(string, int64) error
}

var _ ResourceRepository = (*DBService)(nil)

func (s *DBService) CreateResource(res *model.Resource) error {
	query := "INSERT INTO resource (title, content, url, images, last_edition_at, last_edition_by, section_id) VALUES($1, $2, $3, $4, $5, $6, $7) RETURNING id"
	now := time.Now()
	res.LastEditionAt = now
	if err := s.DB.QueryRow(query, res.Title, res.Content, res.URL, pq.Array(res.Images), time.Now(), res.LastEditionBy, res.SectionId).Scan(&res.Id); err != nil {
		return err
	}
	return nil
}

func (s *DBService) ReadResourcesBySectionID(sectionId int64) (*[]model.Resource, error) {
	var resources []model.Resource
	query := "SELECT * FROM resource WHERE section_id=$1"
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
	if err := s.DB.QueryRow("UPDATE resource SET title=$1, content=$2, url=$3, last_edition_at=$4, last_edition_by=$5 WHERE id=$6 RETURNING title, content, url, last_edition_at, section_id",
		res.Title, res.Content, res.URL, time.Now().Format("2006-01-02T15:04:05Z07:00"), res.LastEditionBy, res.Id).Scan(&res.Title, &res.Content, &res.URL, &res.LastEditionAt, &res.SectionId); err != nil {
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

func (s *DBService) SaveImageURL(imgUrl string, resourceId int64) error {
	imgs := new([]string)
	err := s.DB.QueryRow("SELECT images FROM resource WHERE id=$1", resourceId).Scan(pq.Array(imgs))
	if err != nil {
		return err
	}
	*imgs = append(*imgs, imgUrl)
	log.Printf("Images Slice: %v\nResource ID: %v", *imgs, resourceId)
	_, err = s.DB.Exec("UPDATE resource SET images=$1 WHERE id=$2", pq.Array(*imgs), resourceId)
	if err != nil {
		return err
	}
	return nil
}

func (s *DBService) GetImagesByResourceID(resourceId int64) ([]string, error) {
	var images []string
	err := s.DB.QueryRow("SELECT images FROM resource WHERE id=$1", resourceId).Scan(pq.Array(&images))
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("no images found")
		}
		return nil, err
	}
	return images, nil
}

func (s *DBService) DeleteImageByResourceID(imgName string, resourceId int64) error {
	var images []string
	err := s.DB.QueryRow("SELECT images FROM resource WHERE id=$1", resourceId).Scan(pq.Array(&images))
	if err != nil {
		return err
	}
	var newImagesList []string
	for i, img := range images {
		if path.Base(img) == imgName {
			newImagesList = append(images[:i], images[i+1:]...)
			break
		}
	}
	_, err = s.DB.Exec("UPDATE resource SET images=$1 WHERE id=$2", pq.Array(newImagesList), resourceId)
	if err != nil {
		return err
	}
	return nil
}
