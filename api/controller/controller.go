package controller

import (
	repository "core/internal/repository"
	service "core/internal/service"

	"gorm.io/gorm"
)

type Controller struct {
	db            *gorm.DB
	campaigns     *repository.CampaignRepository
	companies     *repository.CompanyRepository
	research      *repository.ResearchRepository
	opportunities *repository.OpportunityRepository
	drafts        *repository.DraftRepository

	researcher *service.Research
	gmail      *service.Gmail
}

func Create(
	db *gorm.DB,
	researcher *service.Research,
	gmail *service.Gmail,
) *Controller {
	return &Controller{
		db:            db,
		campaigns:     repository.NewCampaigns(db),
		companies:     repository.NewCompanies(db),
		research:      repository.NewResearch(db),
		opportunities: repository.NewOpportunities(db),
		drafts:        repository.NewDrafts(db),
		researcher:    researcher,
		gmail:         gmail,
	}
}
