package repository

import (
	model "core/internal/model"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type OpportunityRepository struct {
	db *gorm.DB
}

func NewOpportunities(db *gorm.DB) *OpportunityRepository {
	return &OpportunityRepository{db: db}
}

type OpportunityFilter struct {
	CampaignID uuid.UUID
	Query      string
	Status     string
	Priority   string
}

type OpportunityCounts struct {
	All       int64 `json:"all"`
	New       int64 `json:"new"`
	Reviewing int64 `json:"reviewing"`
	Qualified int64 `json:"qualified"`
	Declined  int64 `json:"declined"`
}

func (r *OpportunityRepository) List(filter OpportunityFilter) ([]model.InboundOpportunity, error) {
	query := r.db.
		Model(&model.InboundOpportunity{}).
		Where("campaign_id = ?", filter.CampaignID)

	if filter.Query != "" {
		like := "%" + filter.Query + "%"
		query = query.Where(
			"subject ILIKE ? OR sender_name ILIKE ? OR sender_email ILIKE ? OR company_name ILIKE ?",
			like, like, like, like,
		)
	}
	if filter.Status != "" {
		query = query.Where("status = ?", filter.Status)
	}
	if filter.Priority != "" {
		query = query.Where("priority = ?", filter.Priority)
	}

	var opportunities []model.InboundOpportunity
	err := query.
		Order("CASE status WHEN 'new' THEN 0 WHEN 'reviewing' THEN 1 ELSE 2 END").
		Order("CASE priority WHEN 'high' THEN 0 WHEN 'medium' THEN 1 ELSE 2 END").
		Order("received_at DESC").
		Find(&opportunities).Error
	return opportunities, err
}

func (r *OpportunityRepository) Counts(campaignID uuid.UUID) (OpportunityCounts, error) {
	var counts OpportunityCounts
	err := r.db.Raw(`
		SELECT
			COUNT(*) AS "all",
			COUNT(*) FILTER (WHERE status = 'new') AS new,
			COUNT(*) FILTER (WHERE status = 'reviewing') AS reviewing,
			COUNT(*) FILTER (WHERE status = 'qualified') AS qualified,
			COUNT(*) FILTER (WHERE status = 'declined') AS declined
		FROM inbound_opportunities
		WHERE campaign_id = ?
	`, campaignID).
		Scan(&counts).Error
	return counts, err
}

func (r *OpportunityRepository) Get(id uuid.UUID) (*model.InboundOpportunity, error) {
	var opportunity model.InboundOpportunity
	if err := r.db.First(&opportunity, "id = ?", id).Error;
		err != nil {
		return nil, err
	}
	return &opportunity, nil
}

func (r *OpportunityRepository) Create(opportunity *model.InboundOpportunity) error {
	return r.db.Create(opportunity).Error
}

func (r *OpportunityRepository) Update(id uuid.UUID, fields map[string]interface{}) error {
	result := r.db.
		Model(&model.InboundOpportunity{}).
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
