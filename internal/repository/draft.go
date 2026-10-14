package repository

import (
	model "core/internal/model"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type DraftRepository struct {
	db *gorm.DB
}

func NewDrafts(db *gorm.DB) *DraftRepository {
	return &DraftRepository{db: db}
}

func (r *DraftRepository) List(companyID uuid.UUID, opportunityID uuid.UUID) ([]model.MessageDraft, error) {
	query := r.db.Model(&model.MessageDraft{})
	if companyID != uuid.Nil {
		query = query.Where("company_id = ?", companyID)
	}
	if opportunityID != uuid.Nil {
		query = query.Where("opportunity_id = ?", opportunityID)
	}

	var drafts []model.MessageDraft
	err := query.
		Order("created_at DESC").
		Find(&drafts).Error
	return drafts, err
}

func (r *DraftRepository) Get(id uuid.UUID) (*model.MessageDraft, error) {
	var draft model.MessageDraft
	if err := r.db.First(&draft, "id = ?", id).Error;
		err != nil {
		return nil, err
	}
	return &draft, nil
}

func (r *DraftRepository) Create(draft *model.MessageDraft) error {
	return r.db.Create(draft).Error
}

func (r *DraftRepository) Update(id uuid.UUID, fields map[string]any) error {
	result := r.db.
		Model(&model.MessageDraft{}).
		Where("id = ?", id).
		Updates(fields)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

func (r *DraftRepository) Delete(id uuid.UUID) error {
	result := r.db.Delete(&model.MessageDraft{}, "id = ?", id)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}
