package repository

import (
	"strings"

	model "core/internal/model"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type CompanyRepository struct {
	db *gorm.DB
}

func NewCompanies(db *gorm.DB) *CompanyRepository {
	return &CompanyRepository{db: db}
}

type CompanyFilter struct {
	CampaignID       uuid.UUID
	Query            string
	Priority         string
	ResearchStatus   string
	EvaluationStatus string
	Sort             string
	Page             int
	Limit            int
	WithFindings     bool
}

var companySorts = map[string]string{
	"fit":      "fit_score DESC NULLS LAST, name ASC",
	"fit-asc":  "fit_score ASC NULLS LAST, name ASC",
	"a-z":      "name ASC",
	"z-a":      "name DESC",
	"priority": "CASE priority WHEN 'high' THEN 0 WHEN 'medium' THEN 1 ELSE 2 END, fit_score DESC NULLS LAST",
	"recents":  "updated_at DESC",
}

func (r *CompanyRepository) List(filter CompanyFilter) ([]model.Company, int64, error) {
	query := r.db.
		Model(&model.Company{}).
		Where("campaign_id = ?", filter.CampaignID)

	if filter.Query != "" {
		like := "%" + filter.Query + "%"
		query = query.Where(
			"name ILIKE ? OR domain ILIKE ? OR industry ILIKE ? OR location ILIKE ?",
			like, like, like, like,
		)
	}
	if filter.Priority != "" {
		query = query.Where("priority IN ?", strings.Split(filter.Priority, ","))
	}
	if filter.ResearchStatus != "" {
		query = query.Where("research_status IN ?", strings.Split(filter.ResearchStatus, ","))
	}
	if filter.EvaluationStatus != "" {
		query = query.Where("evaluation_status IN ?", strings.Split(filter.EvaluationStatus, ","))
	}

	var count int64
	if err := query.Count(&count).Error;
		err != nil {
		return nil, 0, err
	}

	order, ok := companySorts[filter.Sort]
	if !ok {
		order = companySorts["fit"]
	}

	if filter.WithFindings {
		query = query.Preload("Findings", func(db *gorm.DB) *gorm.DB {
			return db.Order("created_at ASC")
		})
	}

	var companies []model.Company
	err := query.
		Order(order).
		Limit(filter.Limit).
		Offset((filter.Page - 1) * filter.Limit).
		Find(&companies).Error
	return companies, count, err
}

func (r *CompanyRepository) Get(id uuid.UUID) (*model.Company, error) {
	var company model.Company
	err := r.db.
		Preload("Findings", func(db *gorm.DB) *gorm.DB {
			return db.Order("created_at ASC")
		}).
		First(&company, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &company, nil
}

func (r *CompanyRepository) Create(company *model.Company) error {
	return r.db.Omit("Findings").Create(company).Error
}

func (r *CompanyRepository) GetByDomain(
	campaignID uuid.UUID,
	domain string,
) (*model.Company, error) {
	if domain == "" {
		return nil, nil
	}

	var company model.Company

	err := r.db.
		Where("campaign_id = ? AND lower(domain) = lower(?)", campaignID, domain).
		First(&company).Error

	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}

	if err != nil {
		return nil, err
	}

	return &company, nil
}

func (r *CompanyRepository) DomainExists(campaignID uuid.UUID, domain string) (bool, error) {
	if domain == "" {
		return false, nil
	}
	var count int64
	err := r.db.
		Model(&model.Company{}).
		Where("campaign_id = ? AND lower(domain) = lower(?)", campaignID, domain).
		Count(&count).Error
	return count > 0, err
}

func (r *CompanyRepository) InsertIgnoreDuplicates(companies []model.Company) (inserted []model.Company, skipped int, err error) {
	for _, company := range companies {
		exists, err := r.DomainExists(company.CampaignID, company.Domain)
		if err != nil {
			return inserted, skipped, err
		}
		if exists {
			skipped++
			continue
		}
		if err := r.Create(&company);
			err != nil {
			return inserted, skipped, err
		}
		inserted = append(inserted, company)
	}
	return inserted, skipped, nil
}

func (r *CompanyRepository) Update(id uuid.UUID, fields map[string]any) error {
	result := r.db.
		Model(&model.Company{}).
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

func (r *CompanyRepository) Delete(id uuid.UUID) error {
	result := r.db.Delete(&model.Company{}, "id = ?", id)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}
