package model

type Resource struct {
	Id        int64  `json:"id,omitempty"`
	Title     string `json:"title" validate:"required"`
	Content   string `json:"content" validate:"required"`
	URL       string `json:"url" validate:"url"`
	SectionId int64  `json:"section_id" validate:"required"`
}
