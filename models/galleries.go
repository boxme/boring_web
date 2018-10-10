package models

import "github.com/jinzhu/gorm"

type Gallery struct {
	gorm.Model
	UserID uint   `gorm:"not_null;index"`
	Title  string `gorm:"not_null"`
}

type GalleryDB interface {
	Create(gallery *Gallery) error
}

type GalleryService interface {
	GalleryDB
}

type galleryGorm struct {
	db *gorm.DB
}

func (gg *galleryGorm) Create(gallery *Gallery) error {
	return nil
}
