package model

import "time"

type Resource struct {
	Id            int64     `json:"id,omitempty"`
	Title         string    `json:"title" validate:"required"`
	Content       string    `json:"content"`
	URL           string    `json:"url"`
	Images        []string  `json:"images"`
	LastEditionAt time.Time `json:"last_edition_at"`
	LastEditionBy string    `json:"last_edition_by" validate:"required"`
	SectionId     int64     `json:"section_id" validate:"required"`
}

type PatchResource struct {
	Id            int64     `json:"id" validate:"required"`
	Title         string    `json:"title"`
	Content       string    `json:"content"`
	URL           string    `json:"url"`
	Images        []string  `json:"images"`
	LastEditionAt time.Time `json:"last_edition_at"`
	LastEditionBy string    `json:"last_edition_by" validate:"required"`
	SectionId     int64     `json:"section_id"`
}

type DeleteImageReq struct {
	ImageName  string `json:"image" validate:"required"`
	ResourceId int64  `json:"resource_id" validate:"required"`
}
