package entity

import "time"

type News struct {
	ID             string    `json:"id"`
	Name           string    `json:"name"`
	Description    string    `json:"description"`
	ImageURL       string    `json:"image_url"`
	VoiceURL       string    `json:"voice_url"`
	CreatedAt      time.Time `json:"created_at"`
	Links          []Link    `json:"links"`
	SubCategoryIDs []string  `json:"sub_category_ids"`
	Language       string    `json:"language"`
	SiteImageLink  string    `json:"site_image_link"`
	VideoURL       string    `json:"video_url"`
	Text           string    `json:"text" db:"text"`
	SpecialID      string    `json:"special_id"`
}

type Source struct {
	ID           string `json:"id"`
	SiteName     string `json:"site_name"`
	SiteURL      string `json:"site_url"`
	SiteImageURL string `json:"site_image_url"`
}

type Link struct {
	LinkName string `json:"link_name"`
	LinkURL  string `json:"link_url"`
}

type NewsWithCategoryNames struct {
	ID              string    `json:"id"`
	Name            string    `json:"name"`
	Description     string    `json:"description"`
	ImageURL        string    `json:"image_url"`
	CreatedAt       time.Time `json:"created_at"`
	CategoryName    string    `json:"category_name"`
	SubCategoryName string    `json:"subcategory_name"`
}

type GetAllNewsRequest struct {
	Page  int `json:"page"`
	Limit int `json:"limit"`
}

type GetNewsBySubCategory struct {
	Page          int    `json:"page"`
	Limit         int    `json:"limit"`
	SubCategoryId string `json:"subcategory_id"`
}

type GetFilteredNewsRequest struct {
	SubCategoryIDs []string `json:"sub_category_ids,omitempty"`
	CategoryID     string   `json:"category_id,omitempty"`
	Page           int      `json:"page"`
	Limit          int      `json:"limit"`
	SearchTerm     string
}
