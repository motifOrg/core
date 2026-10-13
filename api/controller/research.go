package controller

import (
	"errors"
	"net/http"
	"strings"

	model "core/internal/model"
	service "core/internal/service"
	logger "core/pkg/log"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type CreateFindingRequest struct {
	Kind       string `json:"kind" binding:"required,oneof=fit finding risk missing next_action"`
	Title      string `json:"title" binding:"required,max=200"`
	Content    string `json:"content" binding:"max=4000"`
	SourceURL  string `json:"source_url" binding:"omitempty,url,max=500"`
	Confidence string `json:"confidence" binding:"omitempty,oneof=high medium low"`
}

type CreateMessageRequest struct {
	Body string `json:"body" binding:"required,max=4000"`
}

// CreateFinding godoc
//
//	@Summary		Add finding
//	@Description	Records a manual research finding, risk, missing item or next action
//	@Tags			research
//	@Accept			json
//	@Produce		json
//	@Param			id		path		string					true	"Company ID"
//	@Param			body	body		CreateFindingRequest	true	"Finding"
//	@Success		201		{object}	model.ResearchFinding
//	@Failure		400		{object}	ErrorResponse
//	@Router			/companies/{id}/findings [post]
func (c *Controller) CreateFinding(context *gin.Context) {
	id, _ := uuid.Parse(context.Param("id"))

	var request CreateFindingRequest
	if err := context.ShouldBindJSON(&request);
		err != nil {
		context.JSON(
			http.StatusBadRequest,
			gin.H{"error": err.Error()},
		)

		return
	}

	finding := model.ResearchFinding{
		CompanyID:      id,
		Kind:           request.Kind,
		Title:          strings.TrimSpace(request.Title),
		Content:        strings.TrimSpace(request.Content),
		SourceURL:      strings.TrimSpace(request.SourceURL),
		Confidence:     request.Confidence,
		EvidenceStatus: model.EvidenceUnverified,
	}
	if finding.Confidence == "" {
		finding.Confidence = model.PriorityMedium
	}
	if finding.SourceURL != "" {
		finding.EvidenceStatus = model.EvidenceVerified
	}

	if err := c.research.CreateFinding(&finding);
		err != nil {
		logger.Error(
			"Failed to save finding: %v",
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
		gin.H{"data": finding},
	)
}

// DeleteFinding godoc
//
//	@Summary		Delete finding
//	@Tags			research
//	@Produce		json
//	@Param			id			path		string	true	"Company ID"
//	@Param			findingId	path		string	true	"Finding ID"
//	@Success		200			{object}	map[string]interface{}
//	@Failure		404			{object}	ErrorResponse
//	@Router			/companies/{id}/findings/{findingId} [delete]
func (c *Controller) DeleteFinding(context *gin.Context) {
	id, _ := uuid.Parse(context.Param("id"))
	findingID, _ := uuid.Parse(context.Param("findingId"))

	if err := c.research.DeleteFinding(id, findingID);
		err != nil {
		logger.Error(
			"Failed to delete finding: %v",
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
		gin.H{"data": gin.H{"id": findingID}},
	)
}

// RunResearch godoc
//
//	@Summary		Run research
//	@Description	Runs the configured research provider. Returns 503 until a provider is implemented.
//	@Tags			research
//	@Produce		json
//	@Param			id	path		string	true	"Company ID"
//	@Success		200	{array}		model.ResearchFinding
//	@Failure		503	{object}	ErrorResponse
//	@Router			/companies/{id}/research [post]
func (c *Controller) RunResearch(context *gin.Context) {
	id, _ := uuid.Parse(context.Param("id"))

	company, err := c.companies.Get(id)
	if err != nil {
		logger.Error(
			"Failed to load company: %v",
			err,
		)

		context.JSON(
			http.StatusInternalServerError,
			gin.H{"error": err.Error()},
		)

		return
	}

	icp, _ := c.campaigns.GetICP(company.CampaignID)
	findings, err := c.researcher.Run(context.Request.Context(), *company, icp)
	if errors.Is(err, service.ErrResearchNotConfigured) {
		context.JSON(
			http.StatusServiceUnavailable,
			gin.H{"error": err.Error()},
		)

		return
	}
	if err != nil {
		logger.Error(
			"Failed to run research: %v",
			err,
		)

		context.JSON(
			http.StatusInternalServerError,
			gin.H{"error": err.Error()},
		)

		return
	}

	for i := range findings {
		findings[i].CompanyID = id
		if err := c.research.CreateFinding(&findings[i]);
			err != nil {
			logger.Error(
				"Failed to save findings: %v",
				err,
			)

			context.JSON(
				http.StatusInternalServerError,
				gin.H{"error": err.Error()},
			)

			return
		}
	}
	_ = c.companies.Update(id, map[string]interface{}{"research_status": model.ResearchComplete})

	context.JSON(
		http.StatusOK,
		gin.H{"data": findings},
	)
}

// GetMessages godoc
//
//	@Summary		Research chat history
//	@Tags			research
//	@Produce		json
//	@Param			id	path		string	true	"Company ID"
//	@Success		200	{array}		model.ResearchMessage
//	@Router			/companies/{id}/messages [get]
func (c *Controller) GetMessages(context *gin.Context) {
	id, _ := uuid.Parse(context.Param("id"))

	messages, err := c.research.Messages(id)
	if err != nil {
		logger.Error(
			"Failed to load messages: %v",
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
		gin.H{"data": messages},
	)
}

// CreateMessage godoc
//
//	@Summary		Ask about a company
//	@Description	Saves the question. An assistant reply is only added when a research provider is configured.
//	@Tags			research
//	@Accept			json
//	@Produce		json
//	@Param			id		path		string					true	"Company ID"
//	@Param			body	body		CreateMessageRequest	true	"Question"
//	@Success		201		{array}		model.ResearchMessage
//	@Failure		400		{object}	ErrorResponse
//	@Router			/companies/{id}/messages [post]
func (c *Controller) CreateMessage(context *gin.Context) {
	id, _ := uuid.Parse(context.Param("id"))

	var request CreateMessageRequest
	if err := context.ShouldBindJSON(&request);
		err != nil {
		context.JSON(
			http.StatusBadRequest,
			gin.H{"error": err.Error()},
		)

		return
	}

	body := strings.TrimSpace(request.Body)
	if body == "" {
		context.JSON(
			http.StatusBadRequest,
			gin.H{"error": "message is required"},
		)
		return
	}

	company, err := c.companies.Get(id)
	if err != nil {
		logger.Error(
			"Failed to load company: %v",
			err,
		)

		context.JSON(
			http.StatusInternalServerError,
			gin.H{"error": err.Error()},
		)

		return
	}

	history, err := c.research.Messages(id)
	if err != nil {
		logger.Error(
			"Failed to load messages: %v",
			err,
		)

		context.JSON(
			http.StatusInternalServerError,
			gin.H{"error": err.Error()},
		)

		return
	}

	question := model.ResearchMessage{
		CompanyID:  id,
		Body:       body,
		SenderRole: model.SenderUser,
	}
	if err := c.research.CreateMessage(&question);
		err != nil {
		logger.Error(
			"Failed to save message: %v",
			err,
		)

		context.JSON(
			http.StatusInternalServerError,
			gin.H{"error": err.Error()},
		)

		return
	}

	created := []model.ResearchMessage{question}
	if c.researcher.Enabled() {
		reply, err := c.researcher.Ask(context.Request.Context(), *company, history, body)
		if err != nil {
			logger.Error(
				"Failed to ask research assistant: %v",
				err,
			)

			context.JSON(
				http.StatusInternalServerError,
				gin.H{"error": err.Error()},
			)

			return
		}
		answer := model.ResearchMessage{
			CompanyID:  id,
			Body:       reply,
			SenderRole: model.SenderAI,
		}
		if err := c.research.CreateMessage(&answer);
			err != nil {
			logger.Error(
				"Failed to save reply: %v",
				err,
			)

			context.JSON(
				http.StatusInternalServerError,
				gin.H{"error": err.Error()},
			)

			return
		}
		created = append(created, answer)
	}

	context.JSON(
		http.StatusCreated,
		gin.H{"data": created},
	)
}
