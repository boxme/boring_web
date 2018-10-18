package models

import "github.com/jinzhu/gorm"

const (
	ErrUserIDRequired modelError = "models: user ID is required"
	ErrTitleRequired  modelError = "models: title is required"
)

type Gallery struct {
	gorm.Model
	UserID uint   `gorm:"not_null;index"`
	Title  string `gorm:"not_null"`
}

type GalleryDB interface {
	Create(gallery *Gallery) error
	ByID(id uint) (*Gallery, error)
	ByUserID(id uint) ([]Gallery, error)
	Update(gallery *Gallery) error
	Delete(id uint) error
}

type GalleryService interface {
	GalleryDB
}

type galleryGorm struct {
	db *gorm.DB
}

type galleryService struct {
	GalleryDB
}

type galleryValidator struct {
	GalleryDB
}

type galleryValFn func(*Gallery) error

var _ GalleryDB = &galleryGorm{}

func NewGalleryService(db *gorm.DB) GalleryService {
	return &galleryService{
		GalleryDB: &galleryValidator{
			GalleryDB: &galleryGorm{
				db: db,
			},
		},
	}
}
func (gv *galleryValidator) Create(gallery *Gallery) error {
	err := runGalleryValFns(gallery, gv.userIDRequired, gv.titleRequired)
	if err != nil {
		return err
	}

	return gv.GalleryDB.Create(gallery)
}

func (gg *galleryGorm) ByID(id uint) (*Gallery, error) {
	var gallery Gallery
	db := gg.db.Where("id = ?", id)
	err := first(db, &gallery)
	if err != nil {
		return nil, err
	}

	return &gallery, nil
}

func (gg *galleryGorm) ByUserID(userID uint) ([]Gallery, error) {
	var galleries []Gallery

	db := gg.db.Where("user_id = ?", userID)

	if err := db.Find(&galleries).Error; err != nil {
		return nil, err
	}

	return galleries, nil
}

func (gg *galleryGorm) Create(gallery *Gallery) error {
	return gg.db.Create(gallery).Error
}

func (gv *galleryValidator) Update(gallery *Gallery) error {
	err := runGalleryValFns(gallery, gv.userIDRequired, gv.titleRequired)
	if err != nil {
		return err
	}

	return gv.GalleryDB.Update(gallery)
}

func (gg *galleryGorm) Update(gallery *Gallery) error {
	return gg.db.Save(gallery).Error
}

func (gv *galleryValidator) Delete(id uint) error {
	var gallery Gallery
	gallery.ID = id
	if err := runGalleryValFns(&gallery, gv.nonZeroID); err != nil {
		return err
	}

	return gv.GalleryDB.Delete(gallery.ID)
}

func (gg *galleryGorm) Delete(id uint) error {
	gallery := Gallery{
		Model: gorm.Model{
			ID: id,
		},
	}
	return gg.db.Delete(&gallery).Error
}

func (gv *galleryValidator) userIDRequired(g *Gallery) error {
	if g.UserID <= 0 {
		return ErrUserIDRequired
	}

	return nil
}

func (gv *galleryValidator) titleRequired(g *Gallery) error {
	if g.Title == "" {
		return ErrTitleRequired
	}

	return nil
}

func (gv *galleryValidator) nonZeroID(gallery *Gallery) error {
	if gallery.ID <= 0 {
		return ErrIDInvalid
	}
	return nil
}

func runGalleryValFns(gallery *Gallery, fns ...galleryValFn) error {
	for _, fn := range fns {
		if err := fn(gallery); err != nil {
			return err
		}
	}

	return nil
}
