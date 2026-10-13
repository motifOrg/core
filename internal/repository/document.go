package repository

import (
	model "core/internal/model"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type DocumentRepository struct {
	db *gorm.DB
}

func NewDocuments(db *gorm.DB) *DocumentRepository {
	return &DocumentRepository{db: db}
}

func (r *DocumentRepository) List(campaignID uuid.UUID) ([]model.Document, error) {
	var documents []model.Document
	err := r.db.
		Where("campaign_id = ?", campaignID).
		Order("created_at DESC").
		Find(&documents).Error
	return documents, err
}

func (r *DocumentRepository) Get(id uuid.UUID) (*model.Document, error) {
	var document model.Document
	if err := r.db.First(&document, "id = ?", id).Error;
		err != nil {
		return nil, err
	}
	return &document, nil
}

func (r *DocumentRepository) Create(document *model.Document) error {
	return r.db.Create(document).Error
}

func (r *DocumentRepository) Delete(id uuid.UUID) error {
	result := r.db.Delete(&model.Document{}, "id = ?", id)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}
