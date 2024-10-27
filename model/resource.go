package model

type Resource struct {
	Id        int64  `json:"id,omitempty"`
	Title     string `json:"title" validate:"required"`
	Content   string `json:"content"`
	URL       string `json:"url"`
	SectionId int64  `json:"section_id" validate:"required"`
}

type PatchResource struct {
	Id        int64  `json:"id" validate:"required"`
	Title     string `json:"title"`
	Content   string `json:"content"`
	URL       string `json:"url"`
	SectionId int64  `json:"section_id"`
}
