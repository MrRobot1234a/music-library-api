package models

import (
	"gorm.io/gorm"
)

type Song struct {
    gorm.Model
    Band string `json:"group" gorm:"not null"`
    Title string `json:"song" gorm:"not null"`
    ReleaseDate string `json:"releaseDate"`
    Text string `json:"text"`
    Link string `json:"link"`
}
