package controller

import (
	"fmt"
	"net/http"
	"strings"

	model "core/internal/model"
	logger "core/pkg/log"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type CreateDraftRequest struct {
	CompanyID     *uuid.UUID `json:"company_id"`
	OpportunityID *uuid.UUID `json:"opportunity_id"`
	Recipient     string     `json:"recipient" binding:"max=200"`
	Subject       string     `json:"subject" binding:"max=300"`
	Body          string     `json:"body" binding:"max=20000"`
}

type UpdateDraftRequest struct {
	Recipient *string `json:"recipient" binding:"omitempty,max=200"`
	Subject   *string `json:"subject" binding:"omitempty,max=300"`
	Body      *string `json:"body" binding:"omitempty,max=20000"`
	Status    *string `json:"status" binding:"omitempty,oneof=draft reviewed"`
}

// GetDrafts godoc
//
//	@Summary		List drafts
//	@Description	Drafts for a company or an inbound opportunity
//	@Tags			drafts
//	@Produce		json
//	@Param			company_id		query		string	false	"Company ID"
//	@Param			opportunity_id	query		string	false	"Opportunity ID"
//	@Success		200				{array}		model.MessageDraft
//	@Router			/drafts [get]
func (c *Controller) GetDrafts(context *gin.Context) {
	companyID, _ := uuid.Parse(context.Query("company_id"))
	opportunityID, _ := uuid.Parse(context.Query("opportunity_id"))

	drafts, err := c.drafts.List(companyID, opportunityID)
	if err != nil {
		logger.Error(
			"Failed to list drafts: %v",
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
		gin.H{"data": drafts},
	)
}

// CreateDraft godoc
//
//	@Summary		Create draft
//	@Description	Creates an outreach or reply draft. Outreach drafts with empty subject and body are written by AI from the company, its findings and the campaign ICP; otherwise a plain template fills them. Drafts are never sent.
//	@Tags			drafts
//	@Accept			json
//	@Produce		json
//	@Param			body	body		CreateDraftRequest	true	"Draft"
//	@Success		201		{object}	model.MessageDraft
//	@Failure		400		{object}	ErrorResponse
//	@Router			/drafts [post]
func (c *Controller) CreateDraft(context *gin.Context) {
	var request CreateDraftRequest
	if err := context.ShouldBindJSON(&request);
		err != nil {
		context.JSON(
			http.StatusBadRequest,
			gin.H{"error": err.Error()},
		)

		return
	}
	if (request.CompanyID == nil) == (request.OpportunityID == nil) {
		context.JSON(
			http.StatusBadRequest,
			gin.H{"error": "provide exactly one of company_id or opportunity_id"},
		)
		return
	}

	draft := model.MessageDraft{
		CompanyID:     request.CompanyID,
		OpportunityID: request.OpportunityID,
		Recipient:     strings.TrimSpace(request.Recipient),
		Subject:       strings.TrimSpace(request.Subject),
		Body:          strings.TrimSpace(request.Body),
		Status:        model.DraftStatusDraft,
	}

	if request.CompanyID != nil {
		company, err := c.companies.Get(*request.CompanyID)
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

		if draft.Subject == "" && draft.Body == "" && c.researcher.Enabled() {
			campaign, err := c.campaigns.Get(company.CampaignID)
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

			email, err := c.researcher.Email(*company, *campaign)
			if err != nil {
				logger.Error(
					"Failed to draft email for %s: %v",
					company.Name,
					err,
				)

				context.JSON(
					http.StatusBadGateway,
					gin.H{"error": "Couldn't draft the email with AI. Try again."},
				)

				return
			}

			draft.Subject = email.Subject
			draft.Body = email.Body
		}

		if draft.Subject == "" {
			draft.Subject = fmt.Sprintf("Introduction - %s", company.Name)
		}
		if draft.Body == "" {
			draft.Body = fmt.Sprintf("Hi,\n\nI came across %s and wanted to reach out.\n", company.Name)
		}
	}

	if request.OpportunityID != nil {
		opportunity, err := c.opportunities.Get(*request.OpportunityID)
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
		if draft.Recipient == "" {
			draft.Recipient = opportunity.SenderEmail
		}
		if draft.Subject == "" {
			draft.Subject = "Re: " + opportunity.Subject
		}
		if draft.Body == "" {
			draft.Body = fmt.Sprintf("Hi,\n\nThanks for sending this over.\n\nRe: %s\n", opportunity.Subject)
		}
	}

	if err := c.drafts.Create(&draft);
		err != nil {
		logger.Error(
			"Failed to create draft: %v",
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
		gin.H{"data": draft},
	)
}

// UpdateDraft godoc
//
//	@Summary		Update draft
//	@Description	Edits a draft or marks it reviewed. Editing a reviewed draft returns it to draft.
//	@Tags			drafts
//	@Accept			json
//	@Produce		json
//	@Param			id		path		string				true	"Draft ID"
//	@Param			body	body		UpdateDraftRequest	true	"Fields"
//	@Success		200		{object}	model.MessageDraft
//	@Failure		400		{object}	ErrorResponse
//	@Failure		404		{object}	ErrorResponse
//	@Router			/drafts/{id} [patch]
func (c *Controller) UpdateDraft(context *gin.Context) {
	id, _ := uuid.Parse(context.Param("id"))

	var request UpdateDraftRequest
	if err := context.ShouldBindJSON(&request);
		err != nil {
		context.JSON(
			http.StatusBadRequest,
			gin.H{"error": err.Error()},
		)

		return
	}

	fields := map[string]any{}
	if request.Recipient != nil {
		fields["recipient"] = strings.TrimSpace(*request.Recipient)
	}
	if request.Subject != nil {
		fields["subject"] = strings.TrimSpace(*request.Subject)
	}
	if request.Body != nil {
		fields["body"] = strings.TrimSpace(*request.Body)
	}
	if len(fields) > 0 {
		fields["status"] = model.DraftStatusDraft
	}
	if request.Status != nil {
		fields["status"] = *request.Status
	}
	if len(fields) == 0 {
		context.JSON(
			http.StatusBadRequest,
			gin.H{"error": "nothing to update"},
		)
		return
	}

	if err := c.drafts.Update(id, fields);
		err != nil {
		logger.Error(
			"Failed to update draft: %v",
			err,
		)

		context.JSON(
			http.StatusInternalServerError,
			gin.H{"error": err.Error()},
		)

		return
	}

	draft, err := c.drafts.Get(id)
	if err != nil {
		logger.Error(
			"Failed to load draft: %v",
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
		gin.H{"data": draft},
	)
}

// DeleteDraft godoc
//
//	@Summary		Delete draft
//	@Tags			drafts
//	@Produce		json
//	@Param			id	path		string	true	"Draft ID"
//	@Success		200	{object}	map[string]any
//	@Failure		404	{object}	ErrorResponse
//	@Router			/drafts/{id} [delete]
func (c *Controller) DeleteDraft(context *gin.Context) {
	id, _ := uuid.Parse(context.Param("id"))

	if err := c.drafts.Delete(id);
		err != nil {
		logger.Error(
			"Failed to delete draft: %v",
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
