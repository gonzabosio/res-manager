package model

import "time"

type Resource struct {
	Id            int64     `json:"id,omitempty"`
	Title         string    `json:"title" validate:"required"`
	Content       string    `json:"content"`
	URL           string    `json:"url"`
	Images        []string  `json:"images"`
	LastEditionAt time.Time `json:"last_edition_at"`
	LastEditionBy string    `json:"last_edition_by"`
	SectionId     int64     `json:"section_id" validate:"required"`
}

type PatchResource struct {
	Id        int64    `json:"id" validate:"required"`
	Title     string   `json:"title"`
	Content   string   `json:"content"`
	URL       string   `json:"url"`
	Images    []string `json:"images"`
	SectionId int64    `json:"section_id"`
}
