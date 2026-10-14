package controller

import (
	"net/http"
	"strings"

	model "core/internal/model"
	repository "core/internal/repository"
	logger "core/pkg/log"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type ICPRequest struct {
	Thesis          string   `json:"thesis" binding:"max=2000"`
	Industries      []string `json:"industries" binding:"max=30"`
	Geographies     []string `json:"geographies" binding:"max=30"`
	BusinessModels  []string `json:"business_models" binding:"max=30"`
	Characteristics []string `json:"characteristics" binding:"max=30"`
	Exclusions      []string `json:"exclusions" binding:"max=30"`
	RevenueMin      *int64   `json:"revenue_min" binding:"omitempty,min=0"`
	RevenueMax      *int64   `json:"revenue_max" binding:"omitempty,min=0"`
	EbitdaMin       *int64   `json:"ebitda_min" binding:"omitempty,min=0"`
	EmployeesMin    *int     `json:"employees_min" binding:"omitempty,min=0"`
	EmployeesMax    *int     `json:"employees_max" binding:"omitempty,min=0"`
	Ownership       string   `json:"ownership" binding:"max=200"`
}

type CreateCampaignRequest struct {
	Name        string      `json:"name" binding:"required,max=120"`
	Description string      `json:"description" binding:"max=2000"`
	Workflow    string      `json:"workflow" binding:"omitempty,oneof=outbound inbound both"`
	Status      string      `json:"status" binding:"omitempty,oneof=draft active"`
	GmailLabel  string      `json:"gmail_label" binding:"max=120"`
	ICP         *ICPRequest `json:"icp"`
}

type UpdateCampaignRequest struct {
	Name        *string `json:"name" binding:"omitempty,min=1,max=120"`
	Description *string `json:"description" binding:"omitempty,max=2000"`
	Workflow    *string `json:"workflow" binding:"omitempty,oneof=outbound inbound both"`
	Status      *string `json:"status" binding:"omitempty,oneof=draft active paused archived"`
	GmailLabel  *string `json:"gmail_label" binding:"omitempty,max=120"`
}

type CampaignStatsResponse struct {
	Data repository.CampaignStats `json:"data"`
}

type ErrorResponse struct {
	Error string `json:"error"`
}

// GetCampaigns godoc
//
//	@Summary		List campaigns
//	@Description	Lists non-archived campaigns with company and inbound counts
//	@Tags			campaigns
//	@Produce		json
//	@Param			q		query		string	false	"Search by name"
//	@Param			status	query		string	false	"draft|active|paused|archived|all (default excludes archived)"
//	@Success		200		{array}		model.Campaign
//	@Failure		500		{object}	ErrorResponse
//	@Router			/campaigns [get]
func (c *Controller) GetCampaigns(context *gin.Context) {
	campaigns, err := c.campaigns.List(
		strings.TrimSpace(context.Query("q")),
		context.Query("status"),
	)
	if err != nil {
		logger.Error(
			"Failed to list campaigns: %v",
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
		gin.H{"data": campaigns},
	)
}

// GetCampaign godoc
//
//	@Summary		Get campaign
//	@Description	Returns a campaign with its ICP criteria
//	@Tags			campaigns
//	@Produce		json
//	@Param			id	path		string	true	"Campaign ID"
//	@Success		200	{object}	model.Campaign
//	@Failure		404	{object}	ErrorResponse
//	@Router			/campaigns/{id} [get]
func (c *Controller) GetCampaign(context *gin.Context) {
	id, _ := uuid.Parse(context.Param("id"))

	campaign, err := c.campaigns.Get(id)
	if err != nil {
		context.JSON(
			http.StatusNotFound,
			gin.H{"error": err.Error()},
		)

		return
	}

	context.JSON(
		http.StatusOK,
		gin.H{"data": campaign},
	)
}

// CreateCampaign godoc
//
//	@Summary		Create campaign
//	@Description	Creates a campaign and its ICP criteria in one transaction
//	@Tags			campaigns
//	@Accept			json
//	@Produce		json
//	@Param			body	body		CreateCampaignRequest	true	"Campaign"
//	@Success		201		{object}	model.Campaign
//	@Failure		400		{object}	ErrorResponse
//	@Router			/campaigns [post]
func (c *Controller) CreateCampaign(context *gin.Context) {
	var request CreateCampaignRequest
	if err := context.ShouldBindJSON(&request);
		err != nil {
		context.JSON(
			http.StatusBadRequest,
			gin.H{"error": err.Error()},
		)

		return
	}

	name := strings.TrimSpace(request.Name)
	if name == "" {
		context.JSON(
			http.StatusBadRequest,
			gin.H{"error": "campaign name is required"},
		)
		return
	}

	campaign := model.Campaign{
		Name:        name,
		Description: strings.TrimSpace(request.Description),
		Workflow:    request.Workflow,
		Status:      request.Status,
		GmailLabel:  strings.TrimSpace(request.GmailLabel),
	}
	if campaign.Workflow == "" {
		campaign.Workflow = model.WorkflowOutbound
	}
	if campaign.Status == "" {
		campaign.Status = model.CampaignStatusDraft
	}

	var icp *model.ICPCriteria
	if request.ICP != nil {
		icp = &model.ICPCriteria{
			CampaignID:      campaign.ID,
			Thesis:          strings.TrimSpace(request.ICP.Thesis),
			Industries:      request.ICP.Industries,
			Geographies:     request.ICP.Geographies,
			BusinessModels:  request.ICP.BusinessModels,
			Characteristics: request.ICP.Characteristics,
			Exclusions:      request.ICP.Exclusions,
			RevenueMin:      request.ICP.RevenueMin,
			RevenueMax:      request.ICP.RevenueMax,
			EbitdaMin:       request.ICP.EbitdaMin,
			EmployeesMin:    request.ICP.EmployeesMin,
			EmployeesMax:    request.ICP.EmployeesMax,
			Ownership:       strings.TrimSpace(request.ICP.Ownership),
		}
	}

	if err := c.campaigns.Create(&campaign, icp);
		err != nil {
		logger.Error(
			"Failed to create campaign: %v",
			err,
		)

		context.JSON(
			http.StatusInternalServerError,
			gin.H{"error": err.Error()},
		)

		return
	}

	context.JSON(
		http.StatusCreated,
		gin.H{"data": campaign},
	)
}

// UpdateCampaign godoc
//
//	@Summary		Update campaign
//	@Description	Partially updates name, description, workflow or status
//	@Tags			campaigns
//	@Accept			json
//	@Produce		json
//	@Param			id		path		string					true	"Campaign ID"
//	@Param			body	body		UpdateCampaignRequest	true	"Fields"
//	@Success		200		{object}	model.Campaign
//	@Failure		400		{object}	ErrorResponse
//	@Failure		404		{object}	ErrorResponse
//	@Router			/campaigns/{id} [patch]
func (c *Controller) UpdateCampaign(context *gin.Context) {
	id, _ := uuid.Parse(context.Param("id"))

	var request UpdateCampaignRequest
	if err := context.ShouldBindJSON(&request);
		err != nil {
		context.JSON(
			http.StatusBadRequest,
			gin.H{"error": err.Error()},
		)

		return
	}

	fields := map[string]any{}
	if request.Name != nil {
		fields["name"] = strings.TrimSpace(*request.Name)
	}
	if request.Description != nil {
		fields["description"] = strings.TrimSpace(*request.Description)
	}
	if request.Workflow != nil {
		fields["workflow"] = *request.Workflow
	}
	if request.Status != nil {
		fields["status"] = *request.Status
	}
	if request.GmailLabel != nil {
		fields["gmail_label"] = strings.TrimSpace(*request.GmailLabel)
	}
	if len(fields) == 0 {
		context.JSON(
			http.StatusBadRequest,
			gin.H{"error": "nothing to update"},
		)
		return
	}

	if err := c.campaigns.Update(id, fields);
		err != nil {
		logger.Error(
			"Failed to update campaign: %v",
			err,
		)

		context.JSON(
			http.StatusInternalServerError,
			gin.H{"error": err.Error()},
		)

		return
	}

	campaign, err := c.campaigns.Get(id)
	if err != nil {
		logger.Error(
			"Failed to load campaign: %v",
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
		gin.H{"data": campaign},
	)
}

// DeleteCampaign godoc
//
//	@Summary		Delete campaign
//	@Description	Deletes a campaign with its ICP, companies, inbound opportunities and drafts
//	@Tags			campaigns
//	@Produce		json
//	@Param			id	path		string	true	"Campaign ID"
//	@Success		200	{object}	map[string]any
//	@Failure		404	{object}	ErrorResponse
//	@Router			/campaigns/{id} [delete]
func (c *Controller) DeleteCampaign(context *gin.Context) {
	id, _ := uuid.Parse(context.Param("id"))

	if err := c.campaigns.Delete(id);
		err != nil {
		logger.Error(
			"Failed to delete campaign: %v",
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
		gin.H{"data": gin.H{"id": id}},
	)
}

// GetCampaignStats godoc
//
//	@Summary		Campaign stats
//	@Description	Lead and inbound counts for the dashboard
//	@Tags			campaigns
//	@Produce		json
//	@Param			id	path		string	true	"Campaign ID"
//	@Success		200	{object}	CampaignStatsResponse
//	@Router			/campaigns/{id}/stats [get]
func (c *Controller) GetCampaignStats(context *gin.Context) {
	id, _ := uuid.Parse(context.Param("id"))

	stats, err := c.campaigns.Stats(id)
	if err != nil {
		logger.Error(
			"Failed to load stats: %v",
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
		gin.H{"data": stats},
	)
}

// GetICP godoc
//
//	@Summary		Get ICP
//	@Description	Returns the structured acquisition criteria for a campaign
//	@Tags			campaigns
//	@Produce		json
//	@Param			id	path		string	true	"Campaign ID"
//	@Success		200	{object}	model.ICPCriteria
//	@Failure		404	{object}	ErrorResponse
//	@Router			/campaigns/{id}/icp [get]
func (c *Controller) GetICP(context *gin.Context) {
	id, _ := uuid.Parse(context.Param("id"))

	icp, err := c.campaigns.GetICP(id)
	if err != nil {
		logger.Error(
			"Failed to load icp: %v",
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
		gin.H{"data": icp},
	)
}

// UpdateICP godoc
//
//	@Summary		Save ICP
//	@Description	Replaces the structured acquisition criteria for a campaign
//	@Tags			campaigns
//	@Accept			json
//	@Produce		json
//	@Param			id		path		string		true	"Campaign ID"
//	@Param			body	body		ICPRequest	true	"Criteria"
//	@Success		200		{object}	model.ICPCriteria
//	@Failure		400		{object}	ErrorResponse
//	@Router			/campaigns/{id}/icp [put]
func (c *Controller) UpdateICP(context *gin.Context) {
	id, _ := uuid.Parse(context.Param("id"))

	var request ICPRequest
	if err := context.ShouldBindJSON(&request);
		err != nil {
		context.JSON(
			http.StatusBadRequest,
			gin.H{"error": err.Error()},
		)

		return
	}
	icp := &model.ICPCriteria{
		CampaignID:      id,
		Thesis:          strings.TrimSpace(request.Thesis),
		Industries:      request.Industries,
		Geographies:     request.Geographies,
		BusinessModels:  request.BusinessModels,
		Characteristics: request.Characteristics,
		Exclusions:      request.Exclusions,
		RevenueMin:      request.RevenueMin,
		RevenueMax:      request.RevenueMax,
		EbitdaMin:       request.EbitdaMin,
		EmployeesMin:    request.EmployeesMin,
		EmployeesMax:    request.EmployeesMax,
		Ownership:       strings.TrimSpace(request.Ownership),
	}
	if err := c.campaigns.SaveICP(icp);
		err != nil {
		logger.Error(
			"Failed to save icp: %v",
			err,
		)

		context.JSON(
			http.StatusInternalServerError,
			gin.H{"error": err.Error()},
		)

		return
	}
	_ = c.campaigns.Update(id, map[string]any{"updated_at": icp.UpdatedAt})

	context.JSON(
		http.StatusOK,
		gin.H{"data": icp},
	)
}
