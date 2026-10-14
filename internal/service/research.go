package service

import (
	"errors"

	model "core/internal/model"
	repository "core/internal/repository"
	logger "core/pkg/log"
	openai "core/pkg/openai"

	"gorm.io/gorm"
)

var ErrResearchNotConfigured = errors.New("OPENAI_API_KEY is not set")

type Research struct {
	client    *openai.ClientStruct
	companies *repository.CompanyRepository
	research  *repository.ResearchRepository
}

func NewResearch(
	db *gorm.DB,
	client *openai.ClientStruct,
) *Research {
	return &Research{
		client:    client,
		companies: repository.NewCompanies(db),
		research:  repository.NewResearch(db),
	}
}

func (r *Research) Enabled() bool {
	return r.client != nil
}

func (r *Research) Enrich(
	companies []model.Company,
	icp *model.ICPCriteria,
) {
	for _, company := range companies {
		if _, err := r.Run(company, icp);
			err != nil {
			logger.Error(
				"Failed to enrich %s: %v",
				company.Name,
				err,
			)
		}
	}
}

func (r *Research) Run(
	company model.Company,
	icp *model.ICPCriteria,
) ([]model.ResearchFinding, error) {
	if !r.Enabled() {
		return nil, ErrResearchNotConfigured
	}

	_ = r.companies.Update(company.ID, map[string]any{
		"research_status": model.ResearchInProgress,
	})

	failed := true

	defer func() {
		if !failed {
			return
		}

		_ = r.companies.Update(company.ID, map[string]any{
			"research_status": model.ResearchNotStarted,
		})
	}()

	result, err := r.client.Enrich(company, icp)

	if err != nil {
		return nil, err
	}

	fields := map[string]any{
		"fit_score":       result.FitScore,
		"fit_summary":     result.FitSummary,
		"priority":        result.Priority,
		"research_status": model.ResearchComplete,
	}

	findings := []model.ResearchFinding{}

	if result.Revenue != nil {
		if result.Revenue.RevenueEstimate != "" {
			fields["revenue_estimate"] = result.Revenue.RevenueEstimate
		}

		if result.Revenue.Employees != nil {
			fields["employees"] = *result.Revenue.Employees
		}

		source := ""
		if len(result.Revenue.Sources) > 0 {
			source = result.Revenue.Sources[0]
		}

		findings = append(findings, model.ResearchFinding{
			Kind:       model.FindingGeneral,
			Title:      "Revenue and headcount",
			Content:    result.Revenue.Summary,
			SourceURL:  source,
			Confidence: result.Revenue.Confidence,
		})
	}

	if result.News != nil {
		for _, item := range result.News.Items {
			findings = append(findings, model.ResearchFinding{
				Kind:       item.Kind,
				Title:      item.Title,
				Content:    item.Content,
				SourceURL:  item.SourceURL,
				Confidence: item.Confidence,
			})
		}
	}

	for i := range findings {
		findings[i].CompanyID = company.ID
		findings[i].EvidenceStatus = model.EvidenceUnverified

		if findings[i].SourceURL != "" {
			findings[i].EvidenceStatus = model.EvidenceVerified
		}

		if err := r.research.CreateFinding(&findings[i]);
			err != nil {
			return nil, err
		}
	}

	if err := r.companies.Update(company.ID, fields);
		err != nil {
		return nil, err
	}

	failed = false

	return findings, nil
}

func (r *Research) Email(
	company model.Company,
	campaign model.Campaign,
) (*openai.EmailResponse, error) {
	if !r.Enabled() {
		return nil, ErrResearchNotConfigured
	}

	return r.client.Email(company, campaign)
}

func (r *Research) Ask(
	company model.Company,
	icp *model.ICPCriteria,
	history []model.ResearchMessage,
	question string,
) (string, error) {
	if !r.Enabled() {
		return "", ErrResearchNotConfigured
	}

	result, err := r.client.Ask(
		company, 
		icp, 
		history, 
		question,
	)

	if err != nil {
		return "", err
	}

	return result.Output, nil
}
