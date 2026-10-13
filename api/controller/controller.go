package controller

import (
	repository "core/internal/repository"
	service "core/internal/service"

	"gorm.io/gorm"
)

type Controller struct {
	db            *gorm.DB
	campaigns     *repository.CampaignRepository
	documents     *repository.DocumentRepository
	companies     *repository.CompanyRepository
	research      *repository.ResearchRepository
	opportunities *repository.OpportunityRepository
	drafts        *repository.DraftRepository

	storage    *service.Storage
	researcher *service.Research
	gmail      *service.Gmail
}

func Create(
	db *gorm.DB,
	storage *service.Storage,
	researcher *service.Research,
	gmail *service.Gmail,
) *Controller {
	return &Controller{
		db:            db,
		campaigns:     repository.NewCampaigns(db),
		documents:     repository.NewDocuments(db),
		companies:     repository.NewCompanies(db),
		research:      repository.NewResearch(db),
		opportunities: repository.NewOpportunities(db),
		drafts:        repository.NewDrafts(db),
		storage:       storage,
		researcher:    researcher,
		gmail:         gmail,
	}
}
