package entity

import "time"

type Ad struct {
	ID        string    `json:"id"`
	Link      string    `json:"link"`
	ImageURL  string    `json:"image_url"`
	ViewCount int       `json:"view_count"`
	CreatedAt time.Time `json:"created_at"`
}

type CreateAdRequest struct {
	Link     string `json:"link"`
	ImageURL string `json:"image_url"`
	ID       string `json:"id"`
}

type GetAdRequest struct {
	IsAdmin bool
	ID      string `json:"id"`
}
