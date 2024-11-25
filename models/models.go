package models

import (
	"time"

	"gorm.io/gorm"
)

// Структуры с метаданными GORM(параметры + индексы)

type Model struct {
	ID        int            `gorm:"primarykey" json:"id" example:"1"`
	CreatedAt time.Time      `json:"created_at" example:"2024-11-23 18:55:28.896205+03"`
	UpdatedAt time.Time      `gorm:"autoUpdateTime" json:"updated_at" example:"2024-11-23 18:55:28.896205+03"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

type Group struct {
	Model
	Name string `gorm:"size:255;not null;uniqueIndex" json:"name"`
}

type Song struct {
	Model
	GroupID     int      `gorm:"not null;index" json:"group_id" example:"1"`
	Title       string   `gorm:"size:255;not null;index" json:"title" example:"Centuries"`
	ReleaseDate string   `gorm:"size:10;index" json:"release_date" example:"01.01.2019"`
	Link        string   `gorm:"size:255;index" json:"link" example:"https://www.youtube.com/watch?v=LBr7kECsjcQ"`
	Lyrics      []Lyrics `gorm:"foreignKey:SongID" json:"lyrics"`
}

type Lyrics struct {
	Model
	SongID int    `gorm:"not null;index" json:"song_id" example:"1"`
	Verse  string `gorm:"not null;index" json:"verse" example:"Some legends are told"`
	Order  int    `gorm:"not null;index" json:"order" example:"1"`
}

// Запросы

type Input struct {
	Group string `json:"group" example:"Fall Out Boys"`
	Song  string `json:"song" example:"Centuries"`
}

type Edit struct {
	Title       string   `json:"title" example:"Centuries"`
	Lyrics      []Lyrics `json:"lyrics"`
	ReleaseDate string   `json:"release_date" example:"01.01.2019"`
	Link        string   `json:"link" example:"https://www.youtube.com/watch?v=LBr7kECsjcQ"`
	GroupName   string   `json:"group_name" example:"Fall Out Boys"`
}

// Ответы

type SongsList struct {
	Data       []Song `json:"data"`
	TotalCount int64  `json:"total_count" example:"100"`
	Page       int    `json:"page" example:"1"`
	Limit      int    `json:"limit" example:"10"`
}
