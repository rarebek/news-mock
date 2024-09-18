package models

import (
	"time"

	"tarkib.uz/internal/entity"
)

type NewsWithCategoryNames struct {
	ID              string    `json:"id"`
	Name            string    `json:"name"`
	Description     string    `json:"description"`
	ImageURL        string    `json:"image_url"`
	CreatedAt       time.Time `json:"created_at"`
	CategoryName    string    `json:"category_name"`
	SubCategoryName string    `json:"subcategory_name"`
}

type News struct {
	UzName         string        `json:"uz_name"`
	RuName         string        `json:"ru_name"`
	UzDescription  string        `json:"uz_description"`
	RuDescription  string        `json:"ru_description"`
	ImageURL       string        `json:"image_url"`
	VoiceURL       string        `json:"voice_url"`
	SubCategoryIDs []string      `json:"sub_category_ids"`
	Links          []entity.Link `json:"links"`
	SiteImageLink  string        `json:"site_image_link"`
	VideoURL       string        `json:"video_url"`
	UzText         string        `json:"uz_text"`
	RuText         string        `json:"ru_text"`
}

type NewsOneLang struct {
	Name           string        `json:"name"`
	Description    string        `json:"description"`
	ImageURL       string        `json:"image_url"`
	VoiceURL       string        `json:"voice_url"`
	SubCategoryIDs []string      `json:"sub_category_ids"`
	Links          []entity.Link `json:"links"`
	SiteImgeLink   string        `json:"site_image_link"`
	VideoURL       string        `json:"video_url"`
	Text           string        `json:"text"`
}
type Link struct {
	LinkName string `json:"link_name"`
	LinkURL  string `json:"link_url"`
}
