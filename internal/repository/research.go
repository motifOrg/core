package repository

import (
	model "core/internal/model"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type ResearchRepository struct {
	db *gorm.DB
}

func NewResearch(db *gorm.DB) *ResearchRepository {
	return &ResearchRepository{db: db}
}

func (r *ResearchRepository) CreateFinding(finding *model.ResearchFinding) error {
	return r.db.Create(finding).Error
}

func (r *ResearchRepository) DeleteFinding(companyID uuid.UUID, id uuid.UUID) error {
	result := r.db.Delete(&model.ResearchFinding{}, "id = ? AND company_id = ?", id, companyID)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

func (r *ResearchRepository) Messages(companyID uuid.UUID) ([]model.ResearchMessage, error) {
	var messages []model.ResearchMessage
	err := r.db.
		Where("company_id = ?", companyID).
		Order("created_at ASC").
		Find(&messages).Error
	return messages, err
}

func (r *ResearchRepository) CreateMessage(message *model.ResearchMessage) error {
	return r.db.Create(message).Error
}
