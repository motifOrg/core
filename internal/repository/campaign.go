package repository

import (
	model "core/internal/model"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type CampaignRepository struct {
	db *gorm.DB
}

func NewCampaigns(db *gorm.DB) *CampaignRepository {
	return &CampaignRepository{db: db}
}

const campaignCounts = `campaigns.*,
	(SELECT COUNT(*) FROM companies WHERE companies.campaign_id = campaigns.id) AS company_count,
	(SELECT COUNT(*) FROM inbound_opportunities WHERE inbound_opportunities.campaign_id = campaigns.id) AS opportunity_count`

func (r *CampaignRepository) List(q string, status string) ([]model.Campaign, error) {
	query := r.db.
		Model(&model.Campaign{}).
		Select(campaignCounts).
		Preload("ICP")

	if q != "" {
		query = query.Where("name ILIKE ?", "%"+q+"%")
	}
	switch status {
	case "all":
	case "":
		query = query.Where("status <> ?", model.CampaignStatusArchived)
	default:
		query = query.Where("status = ?", status)
	}

	var campaigns []model.Campaign
	err := query.
		Order("updated_at DESC").
		Find(&campaigns).Error
	return campaigns, err
}

func (r *CampaignRepository) Get(id uuid.UUID) (*model.Campaign, error) {
	var campaign model.Campaign
	err := r.db.
		Model(&model.Campaign{}).
		Select(campaignCounts).
		Preload("ICP").
		First(&campaign, "campaigns.id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &campaign, nil
}

func (r *CampaignRepository) Create(campaign *model.Campaign, icp *model.ICPCriteria) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Omit("ICP").Create(campaign).Error;
			err != nil {
			return err
		}
		if icp == nil {
			icp = &model.ICPCriteria{}
		}
		icp.CampaignID = campaign.ID
		if err := tx.Create(icp).Error;
			err != nil {
			return err
		}
		campaign.ICP = icp
		return nil
	})
}

func (r *CampaignRepository) Update(id uuid.UUID, fields map[string]any) error {
	result := r.db.
		Model(&model.Campaign{}).
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

func (r *CampaignRepository) Delete(id uuid.UUID) error {
	result := r.db.Delete(&model.Campaign{}, "id = ?", id)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

func (r *CampaignRepository) GetICP(campaignID uuid.UUID) (*model.ICPCriteria, error) {
	var icp model.ICPCriteria
	err := r.db.
		Where("campaign_id = ?", campaignID).
		First(&icp).Error
	if err != nil {
		return nil, err
	}
	return &icp, nil
}

func (r *CampaignRepository) SaveICP(icp *model.ICPCriteria) error {
	existing, err := r.GetICP(icp.CampaignID)
	if err == gorm.ErrRecordNotFound {
		return r.db.Create(icp).Error
	}
	if err != nil {
		return err
	}
	icp.ID = existing.ID
	icp.CreatedAt = existing.CreatedAt
	return r.db.Save(icp).Error
}

type CampaignStats struct {
	Companies    int64 `json:"companies"`
	HighPriority int64 `json:"high_priority"`
	Researched   int64 `json:"researched"`
	Shortlisted  int64 `json:"shortlisted"`
	Contacted    int64 `json:"contacted"`
	InboundNew   int64 `json:"inbound_new"`
}

func (r *CampaignRepository) Stats(campaignID uuid.UUID) (CampaignStats, error) {
	var stats CampaignStats
	err := r.db.Raw(`
		SELECT
			COUNT(*) AS companies,
			COUNT(*) FILTER (WHERE priority = 'high') AS high_priority,
			COUNT(*) FILTER (WHERE research_status = 'complete') AS researched,
			COUNT(*) FILTER (WHERE evaluation_status = 'shortlisted') AS shortlisted,
			COUNT(*) FILTER (WHERE evaluation_status IN ('contacted', 'in_discussion')) AS contacted,
			(SELECT COUNT(*) FROM inbound_opportunities
				WHERE campaign_id = @id AND status = 'new') AS inbound_new
		FROM companies
		WHERE campaign_id = @id
	`, map[string]any{"id": campaignID}).
		Scan(&stats).Error
	return stats, err
}
