package controller

import (
	"net/http"
	"strings"
	"time"

	model "core/internal/model"
	repository "core/internal/repository"
	service "core/internal/service"
	csv "core/pkg/csv"
	gmail "core/pkg/gmail"
	logger "core/pkg/log"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type CreateOpportunityRequest struct {
	CampaignID    uuid.UUID `json:"campaign_id" binding:"required"`
	SenderName    string    `json:"sender_name" binding:"max=120"`
	SenderEmail   string    `json:"sender_email" binding:"required,email,max=200"`
	Subject       string    `json:"subject" binding:"required,max=300"`
	Body          string    `json:"body" binding:"max=20000"`
	CompanyName   string    `json:"company_name" binding:"max=200"`
	CompanyDomain string    `json:"company_domain" binding:"max=200"`
}

type UpdateOpportunityRequest struct {
	Status     *string `json:"status" binding:"omitempty,oneof=new reviewing qualified declined"`
	Priority   *string `json:"priority" binding:"omitempty,oneof=high medium low"`
	FitSummary *string `json:"fit_summary" binding:"omitempty,max=4000"`
}

type ResearchStatus struct {
	Status  string `json:"status"`
	Message string `json:"message"`
}

type IntegrationsData struct {
	Gmail    any    `json:"gmail"`
	Research ResearchStatus `json:"research"`
}

// GetOpportunities godoc
//
//	@Summary		List inbound opportunities
//	@Description	Inbound deals for a campaign, new and high priority first, with status counts
//	@Tags			inbound
//	@Produce		json
//	@Param			campaign_id	query		string	true	"Campaign ID"
//	@Param			q			query		string	false	"Search subject, sender or company"
//	@Param			status		query		string	false	"new|reviewing|qualified|declined"
//	@Param			priority	query		string	false	"high|medium|low"
//	@Success		200			{array}		model.InboundOpportunity
//	@Failure		400			{object}	ErrorResponse
//	@Router			/inbound [get]
func (c *Controller) GetOpportunities(context *gin.Context) {
	campaignID, _ := uuid.Parse(context.Query("campaign_id"))

	opportunities, err := c.opportunities.List(repository.OpportunityFilter{
		CampaignID: campaignID,
		Query:      strings.TrimSpace(context.Query("q")),
		Status:     context.Query("status"),
		Priority:   context.Query("priority"),
	})
	if err != nil {
		logger.Error(
			"Failed to list inbound: %v",
			err,
		)

		context.JSON(
			http.StatusInternalServerError,
			gin.H{"error": err.Error()},
		)

		return
	}

	counts, err := c.opportunities.Counts(campaignID)
	if err != nil {
		logger.Error(
			"Failed to count inbound: %v",
			err,
		)

		context.JSON(
			http.StatusInternalServerError,
			gin.H{"error": err.Error()},
		)

		return
	}

	context.JSON(
		http.StatusOK,
		gin.H{
			"data":   opportunities,
			"counts": counts,
		},
	)
}

// GetOpportunity godoc
//
//	@Summary		Get inbound opportunity
//	@Tags			inbound
//	@Produce		json
//	@Param			id	path		string	true	"Opportunity ID"
//	@Success		200	{object}	model.InboundOpportunity
//	@Failure		404	{object}	ErrorResponse
//	@Router			/inbound/{id} [get]
func (c *Controller) GetOpportunity(context *gin.Context) {
	id, _ := uuid.Parse(context.Param("id"))

	opportunity, err := c.opportunities.Get(id)
	if err != nil {
		context.JSON(
			http.StatusNotFound,
			gin.H{"error": err.Error()},
		)

		return
	}

	context.JSON(
		http.StatusOK,
		gin.H{"data": opportunity},
	)
}

// CreateOpportunity godoc
//
//	@Summary		Log inbound opportunity
//	@Description	Manually records an inbound deal email until Gmail sync is available
//	@Tags			inbound
//	@Accept			json
//	@Produce		json
//	@Param			body	body		CreateOpportunityRequest	true	"Opportunity"
//	@Success		201		{object}	model.InboundOpportunity
//	@Failure		400		{object}	ErrorResponse
//	@Router			/inbound [post]
func (c *Controller) CreateOpportunity(context *gin.Context) {
	var request CreateOpportunityRequest
	if err := context.ShouldBindJSON(&request);
		err != nil {
		context.JSON(
			http.StatusBadRequest,
			gin.H{"error": err.Error()},
		)

		return
	}

	opportunity := model.InboundOpportunity{
		CampaignID:    request.CampaignID,
		Source:        model.SourceManual,
		SenderName:    strings.TrimSpace(request.SenderName),
		SenderEmail:   strings.TrimSpace(request.SenderEmail),
		Subject:       strings.TrimSpace(request.Subject),
		Body:          strings.TrimSpace(request.Body),
		CompanyName:   strings.TrimSpace(request.CompanyName),
		CompanyDomain: csv.NormalizeDomain(request.CompanyDomain),
		ReceivedAt:    time.Now(),
		Priority:      model.PriorityMedium,
		Status:        model.OpportunityNew,
	}

	if gmail.FreeMail[opportunity.CompanyDomain] {
		opportunity.CompanyDomain = ""
	}

	if err := c.opportunities.Create(&opportunity);
		err != nil {
		logger.Error(
			"Failed to create opportunity: %v",
			err,
		)

		context.JSON(
			http.StatusInternalServerError,
			gin.H{"error": err.Error()},
		)

		return
	}

	if c.researcher.Enabled() {
		icp, _ := c.campaigns.GetICP(opportunity.CampaignID)
		go c.researcher.EvaluateAll([]model.InboundOpportunity{opportunity}, icp)
	}

	context.JSON(
		http.StatusCreated,
		gin.H{"data": opportunity},
	)
}

// UpdateOpportunity godoc
//
//	@Summary		Update inbound opportunity
//	@Description	Changes evaluation status, priority or the ICP fit note
//	@Tags			inbound
//	@Accept			json
//	@Produce		json
//	@Param			id		path		string						true	"Opportunity ID"
//	@Param			body	body		UpdateOpportunityRequest	true	"Fields"
//	@Success		200		{object}	model.InboundOpportunity
//	@Failure		400		{object}	ErrorResponse
//	@Failure		404		{object}	ErrorResponse
//	@Router			/inbound/{id} [patch]
func (c *Controller) UpdateOpportunity(context *gin.Context) {
	id, _ := uuid.Parse(context.Param("id"))

	var request UpdateOpportunityRequest
	if err := context.ShouldBindJSON(&request);
		err != nil {
		context.JSON(
			http.StatusBadRequest,
			gin.H{"error": err.Error()},
		)

		return
	}

	fields := map[string]any{}
	if request.Status != nil {
		fields["status"] = *request.Status
	}
	if request.Priority != nil {
		fields["priority"] = *request.Priority
	}
	if request.FitSummary != nil {
		fields["fit_summary"] = strings.TrimSpace(*request.FitSummary)
	}
	if len(fields) == 0 {
		context.JSON(
			http.StatusBadRequest,
			gin.H{"error": "nothing to update"},
		)
		return
	}

	if err := c.opportunities.Update(id, fields);
		err != nil {
		logger.Error(
			"Failed to update opportunity: %v",
			err,
		)

		context.JSON(
			http.StatusInternalServerError,
			gin.H{"error": err.Error()},
		)

		return
	}

	opportunity, err := c.opportunities.Get(id)
	if err != nil {
		logger.Error(
			"Failed to load opportunity: %v",
			err,
		)

		context.JSON(
			http.StatusInternalServerError,
			gin.H{"error": err.Error()},
		)

		return
	}

	context.JSON(
		http.StatusOK,
		gin.H{"data": opportunity},
	)
}

// GetIntegrations godoc
//
//	@Summary		Integration status
//	@Description	Reports whether Gmail and the research provider are available
//	@Tags			integrations
//	@Produce		json
//	@Success		200	{object}	IntegrationsData
//	@Router			/integrations [get]
func (c *Controller) GetIntegrations(context *gin.Context) {
	research := ResearchStatus{
		Status:  "setup_required",
		Message: "OPENAI_API_KEY is not set. Findings can be added manually.",
	}
	if c.researcher.Enabled() {
		research = ResearchStatus{Status: "ready", Message: "OpenAI research agent connected."}
	}

	context.JSON(
		http.StatusOK,
		gin.H{"data": IntegrationsData{
			Gmail:    c.gmail.Status(),
			Research: research,
		}},
	)
}

// SyncGmail godoc
//
//	@Summary		Sync Gmail
//	@Description	Pulls the last 7 days of inbound emails, adds sender companies as leads and enriches them. Returns 503 until GMAIL_USER and GMAIL_APP_PASSWORD are set.
//	@Tags			integrations
//	@Produce		json
//	@Param			campaign_id	query		string	true	"Campaign ID"
//	@Success		200			{object}	map[string]any
//	@Failure		503			{object}	ErrorResponse
//	@Router			/integrations/gmail/sync [post]
func (c *Controller) SyncGmail(context *gin.Context) {
	campaignID, _ := uuid.Parse(context.Query("campaign_id"))

	campaign, err := c.campaigns.Get(campaignID)
	if err != nil {
		context.JSON(
			http.StatusNotFound,
			gin.H{"error": err.Error()},
		)

		return
	}

	if !c.gmail.Enabled() {
		context.JSON(
			http.StatusServiceUnavailable,
			gin.H{"error": c.gmail.Status().Message},
		)
		return
	}

	messages, err := c.gmail.Sync(
		time.Now().AddDate(0, 0, -7),
		strings.TrimSpace(campaign.GmailLabel),
	)

	if err != nil {
		logger.Error(
			"Failed to sync gmail: %v",
			err,
		)

		context.JSON(
			http.StatusInternalServerError,
			gin.H{"error": err.Error()},
		)

		return
	}

	leads := []model.Company{}
	deals := []model.InboundOpportunity{}

	for _, message := range messages {
		if message.ExternalRef != "" {
			exists, err := c.opportunities.Exists(
				model.SourceGmail,
				message.ExternalRef,
			)
			if err != nil {
				logger.Error(
					"Failed to check gmail message: %v",
					err,
				)

				context.JSON(
					http.StatusInternalServerError,
					gin.H{"error": err.Error()},
				)

				return
			}

			if exists {
				continue
			}
		}

		_, domain, _ := strings.Cut(message.SenderEmail, "@")
		domain = csv.NormalizeDomain(domain)

		if gmail.FreeMail[domain] {
			domain = ""
		}

		inserted := []model.Company{}

		if domain != "" {
			inserted, _, err = c.companies.InsertIgnoreDuplicates([]model.Company{{
				CampaignID: campaignID,
				Name:       domain,
				Domain:     domain,
				Source:     model.SourceGmail,
			}})
			if err != nil {
				logger.Error(
					"Failed to add lead: %v",
					err,
				)

				context.JSON(
					http.StatusInternalServerError,
					gin.H{"error": err.Error()},
				)

				return
			}
		}

		opportunity := model.InboundOpportunity{
			CampaignID:    campaignID,
			Source:        model.SourceGmail,
			ExternalRef:   message.ExternalRef,
			SenderName:    message.SenderName,
			SenderEmail:   message.SenderEmail,
			Subject:       message.Subject,
			Body:          message.Body,
			CompanyName:   domain,
			CompanyDomain: domain,
			ReceivedAt:    message.ReceivedAt,
			Priority:      model.PriorityMedium,
			Status:        model.OpportunityNew,
		}
		if len(inserted) > 0 {
			opportunity.CompanyID = &inserted[0].ID
			leads = append(leads, inserted[0])
		} else if domain != "" {
			company, err := c.companies.GetByDomain(campaignID, domain)
			if err != nil {
				logger.Error(
					"Failed to load lead: %v",
					err,
				)

				context.JSON(
					http.StatusInternalServerError,
					gin.H{"error": err.Error()},
				)

				return
			}

			if company != nil {
				opportunity.CompanyID = &company.ID
			}
		}

		if err := c.opportunities.Create(&opportunity);
			err != nil {
			logger.Error(
				"Failed to create opportunity: %v",
				err,
			)

			context.JSON(
				http.StatusInternalServerError,
				gin.H{"error": err.Error()},
			)

			return
		}

		deals = append(deals, opportunity)
	}

	if c.researcher.Enabled() {
		go func() {
			c.researcher.EvaluateAll(deals, campaign.ICP)
			c.researcher.Enrich(leads, campaign.ICP)
		}()
	}

	context.JSON(
		http.StatusOK,
		gin.H{"data": gin.H{
			"messages":    len(messages),
			"deals":       len(deals),
			"leads":       len(leads),
			"gmail_label": campaign.GmailLabel,
		}},
	)
}

// EvaluateOpportunity godoc
//
//	@Summary		Evaluate inbound opportunity
//	@Description	Runs the research agent on the email against the campaign ICP and saves fit score, summary and priority
//	@Tags			inbound
//	@Produce		json
//	@Param			id	path		string	true	"Opportunity ID"
//	@Success		200	{object}	model.InboundOpportunity
//	@Failure		404	{object}	ErrorResponse
//	@Failure		503	{object}	ErrorResponse
//	@Router			/inbound/{id}/evaluate [post]
func (c *Controller) EvaluateOpportunity(context *gin.Context) {
	id, _ := uuid.Parse(context.Param("id"))

	if !c.researcher.Enabled() {
		context.JSON(
			http.StatusServiceUnavailable,
			gin.H{"error": service.ErrResearchNotConfigured.Error()},
		)
		return
	}

	opportunity, err := c.opportunities.Get(id)
	if err != nil {
		context.JSON(
			http.StatusNotFound,
			gin.H{"error": err.Error()},
		)

		return
	}

	icp, _ := c.campaigns.GetICP(opportunity.CampaignID)

	evaluated, err := c.researcher.Evaluate(*opportunity, icp)
	if err != nil {
		logger.Error(
			"Failed to evaluate opportunity: %v",
			err,
		)

		context.JSON(
			http.StatusBadGateway,
			gin.H{"error": "Couldn't evaluate the deal with AI. Try again."},
		)

		return
	}

	context.JSON(
		http.StatusOK,
		gin.H{"data": evaluated},
	)
}
