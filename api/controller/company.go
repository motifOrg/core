package controller

import (
	"math"
	"net/http"
	"strconv"
	"strings"

	model "core/internal/model"
	repository "core/internal/repository"
	csv "core/pkg/csv"
	logger "core/pkg/log"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type CreateCompanyRequest struct {
	CampaignID      uuid.UUID `json:"campaign_id" binding:"required"`
	Name            string    `json:"name" binding:"required,max=200"`
	Domain          string    `json:"domain" binding:"max=200"`
	Industry        string    `json:"industry" binding:"max=120"`
	Location        string    `json:"location" binding:"max=120"`
	Description     string    `json:"description" binding:"max=4000"`
	Employees       *int      `json:"employees" binding:"omitempty,min=0"`
	RevenueEstimate string    `json:"revenue_estimate" binding:"max=60"`
	FoundedYear     *int      `json:"founded_year" binding:"omitempty,min=1800,max=2100"`
	OwnerName       string    `json:"owner_name" binding:"max=120"`
}

type UpdateCompanyRequest struct {
	Name             *string `json:"name" binding:"omitempty,min=1,max=200"`
	Domain           *string `json:"domain" binding:"omitempty,max=200"`
	Industry         *string `json:"industry" binding:"omitempty,max=120"`
	Location         *string `json:"location" binding:"omitempty,max=120"`
	Description      *string `json:"description" binding:"omitempty,max=4000"`
	Employees        *int    `json:"employees" binding:"omitempty,min=0"`
	RevenueEstimate  *string `json:"revenue_estimate" binding:"omitempty,max=60"`
	OwnerName        *string `json:"owner_name" binding:"omitempty,max=120"`
	Priority         *string `json:"priority" binding:"omitempty,oneof=high medium low"`
	ResearchStatus   *string `json:"research_status" binding:"omitempty,oneof=not_started in_progress complete"`
	EvaluationStatus *string `json:"evaluation_status" binding:"omitempty,oneof=new shortlisted contacted in_discussion rejected"`
}

type ImportData struct {
	RowsProcessed int            `json:"rows_processed"`
	Inserted      int            `json:"inserted"`
	Skipped       int            `json:"skipped"`
	Failed        int            `json:"failed"`
	Errors        []csv.RowError `json:"errors"`
}

type CompaniesResponse struct {
	Data  []model.Company `json:"data"`
	Total int             `json:"total"`
	Page  int             `json:"page"`
	Count int64           `json:"count"`
}

// GetCompanies godoc
//
//	@Summary		List companies
//	@Description	Paginated, filterable lead list for a campaign
//	@Tags			companies
//	@Produce		json
//	@Param			campaign_id			query		string	true	"Campaign ID"
//	@Param			q					query		string	false	"Search name, domain, industry, location"
//	@Param			priority			query		string	false	"Comma separated: high,medium,low"
//	@Param			research_status		query		string	false	"Comma separated: not_started,in_progress,complete"
//	@Param			evaluation_status	query		string	false	"Comma separated: new,shortlisted,contacted,in_discussion,rejected"
//	@Param			sort				query		string	false	"fit|fit-asc|a-z|z-a|priority|recents"	default(fit)
//	@Param			page				query		int		false	"Page number"	default(1)
//	@Param			limit				query		int		false	"Items per page"	default(25)
//	@Success		200					{object}	CompaniesResponse
//	@Failure		400					{object}	ErrorResponse
//	@Router			/companies [get]
func (c *Controller) GetCompanies(context *gin.Context) {
	campaignID, _ := uuid.Parse(context.Query("campaign_id"))
	page, _ := strconv.Atoi(context.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(context.DefaultQuery("limit", "25"))

	companies, count, err := c.companies.List(repository.CompanyFilter{
		CampaignID:       campaignID,
		Query:            strings.TrimSpace(context.Query("q")),
		Priority:         context.Query("priority"),
		ResearchStatus:   context.Query("research_status"),
		EvaluationStatus: context.Query("evaluation_status"),
		Sort:             context.DefaultQuery("sort", "fit"),
		Page:             page,
		Limit:            limit,
	})
	if err != nil {
		logger.Error(
			"Failed to list companies: %v",
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
			"data":  companies,
			"total": int(math.Ceil(float64(count) / float64(limit))),
			"page":  page,
			"count": count,
		},
	)
}

// GetCompany godoc
//
//	@Summary		Get company
//	@Description	Company profile with research findings
//	@Tags			companies
//	@Produce		json
//	@Param			id	path		string	true	"Company ID"
//	@Success		200	{object}	model.Company
//	@Failure		404	{object}	ErrorResponse
//	@Router			/companies/{id} [get]
func (c *Controller) GetCompany(context *gin.Context) {
	id, _ := uuid.Parse(context.Param("id"))

	company, err := c.companies.Get(id)
	if err != nil {
		context.JSON(
			http.StatusNotFound,
			gin.H{"error": err.Error()},
		)

		return
	}

	context.JSON(
		http.StatusOK,
		gin.H{"data": company},
	)
}

// CreateCompany godoc
//
//	@Summary		Add company
//	@Description	Manually adds a target company to a campaign
//	@Tags			companies
//	@Accept			json
//	@Produce		json
//	@Param			body	body		CreateCompanyRequest	true	"Company"
//	@Success		201		{object}	model.Company
//	@Failure		400		{object}	ErrorResponse
//	@Failure		409		{object}	ErrorResponse
//	@Router			/companies [post]
func (c *Controller) CreateCompany(context *gin.Context) {
	var request CreateCompanyRequest
	if err := context.ShouldBindJSON(&request);
		err != nil {
		context.JSON(
			http.StatusBadRequest,
			gin.H{"error": err.Error()},
		)

		return
	}

	company := model.Company{
		CampaignID:      request.CampaignID,
		Name:            strings.TrimSpace(request.Name),
		Domain:          csv.NormalizeDomain(request.Domain),
		Industry:        strings.TrimSpace(request.Industry),
		Location:        strings.TrimSpace(request.Location),
		Description:     strings.TrimSpace(request.Description),
		Employees:       request.Employees,
		RevenueEstimate: strings.TrimSpace(request.RevenueEstimate),
		FoundedYear:     request.FoundedYear,
		OwnerName:       strings.TrimSpace(request.OwnerName),
		Source:          model.SourceManual,
	}

	exists, err := c.companies.DomainExists(company.CampaignID, company.Domain)
	if err != nil {
		logger.Error(
			"Failed to check company: %v",
			err,
		)

		context.JSON(
			http.StatusInternalServerError,
			gin.H{"error": err.Error()},
		)

		return
	}
	if exists {
		context.JSON(
			http.StatusConflict,
			gin.H{"error": company.Domain + " is already in this campaign"},
		)
		return
	}

	if err := c.companies.Create(&company);
		err != nil {
		logger.Error(
			"Failed to create company: %v",
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
		gin.H{"data": company},
	)
}

// ImportCompanies godoc
//
//	@Summary		Import companies from CSV
//	@Description	Upload a CSV with at least a name column. Domains already in the campaign are skipped.
//	@Tags			companies
//	@Accept			multipart/form-data
//	@Produce		json
//	@Param			id		path		string	true	"Campaign ID"
//	@Param			file	formData	file	true	"CSV file"
//	@Success		200		{object}	ImportData
//	@Failure		400		{object}	ErrorResponse
//	@Router			/campaigns/{id}/companies/import [post]
func (c *Controller) ImportCompanies(context *gin.Context) {
	id, _ := uuid.Parse(context.Param("id"))

	file, err := context.FormFile("file")
	if err != nil {
		context.JSON(
			http.StatusBadRequest,
			gin.H{"error": "csv file is required (form field: file)"},
		)

		return
	}

	opened, err := file.Open()
	if err != nil {
		logger.Error(
			"Failed to open csv: %v",
			err,
		)

		context.JSON(
			http.StatusInternalServerError,
			gin.H{"error": "failed to read csv"},
		)

		return
	}
	defer opened.Close()

	parsed, err := csv.Parse(opened)
	if err != nil {
		context.JSON(
			http.StatusBadRequest,
			gin.H{"error": err.Error()},
		)

		return
	}

	for i := range parsed.Companies {
		parsed.Companies[i].CampaignID = id
	}

	inserted, skipped, err := c.companies.InsertIgnoreDuplicates(parsed.Companies)
	if err != nil {
		logger.Error(
			"Failed to store companies: %v",
			err,
		)

		context.JSON(
			http.StatusInternalServerError,
			gin.H{"error": err.Error()},
		)

		return
	}

	if c.researcher.Enabled() {
		icp, _ := c.campaigns.GetICP(id)
		go c.researcher.Enrich(inserted, icp)
	}

	context.JSON(
		http.StatusOK,
		gin.H{"data": ImportData{
			RowsProcessed: len(parsed.Companies) + len(parsed.Errors),
			Inserted:      len(inserted),
			Skipped:       skipped,
			Failed:        len(parsed.Errors),
			Errors:        parsed.Errors,
		}},
	)
}

// UpdateCompany godoc
//
//	@Summary		Update company
//	@Description	Partially updates profile fields, priority, research or evaluation status
//	@Tags			companies
//	@Accept			json
//	@Produce		json
//	@Param			id		path		string					true	"Company ID"
//	@Param			body	body		UpdateCompanyRequest	true	"Fields"
//	@Success		200		{object}	model.Company
//	@Failure		400		{object}	ErrorResponse
//	@Failure		404		{object}	ErrorResponse
//	@Router			/companies/{id} [patch]
func (c *Controller) UpdateCompany(context *gin.Context) {
	id, _ := uuid.Parse(context.Param("id"))

	var request UpdateCompanyRequest
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
	if request.Domain != nil {
		fields["domain"] = csv.NormalizeDomain(*request.Domain)
	}
	if request.Industry != nil {
		fields["industry"] = strings.TrimSpace(*request.Industry)
	}
	if request.Location != nil {
		fields["location"] = strings.TrimSpace(*request.Location)
	}
	if request.Description != nil {
		fields["description"] = strings.TrimSpace(*request.Description)
	}
	if request.RevenueEstimate != nil {
		fields["revenue_estimate"] = strings.TrimSpace(*request.RevenueEstimate)
	}
	if request.OwnerName != nil {
		fields["owner_name"] = strings.TrimSpace(*request.OwnerName)
	}
	if request.Priority != nil {
		fields["priority"] = *request.Priority
	}
	if request.ResearchStatus != nil {
		fields["research_status"] = *request.ResearchStatus
	}
	if request.EvaluationStatus != nil {
		fields["evaluation_status"] = *request.EvaluationStatus
	}
	if request.Employees != nil {
		fields["employees"] = *request.Employees
	}

	if len(fields) == 0 {
		context.JSON(
			http.StatusBadRequest,
			gin.H{"error": "nothing to update"},
		)
		return
	}
	if name, ok := fields["name"]; ok && name == "" {
		context.JSON(
			http.StatusBadRequest,
			gin.H{"error": "company name is required"},
		)
		return
	}

	if err := c.companies.Update(id, fields);
		err != nil {
		logger.Error(
			"Failed to update company: %v",
			err,
		)

		context.JSON(
			http.StatusInternalServerError,
			gin.H{"error": err.Error()},
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

	context.JSON(
		http.StatusOK,
		gin.H{"data": company},
	)
}

// DeleteCompany godoc
//
//	@Summary		Remove company
//	@Description	Removes a company and its findings, chat and drafts from the campaign
//	@Tags			companies
//	@Produce		json
//	@Param			id	path		string	true	"Company ID"
//	@Success		200	{object}	map[string]any
//	@Failure		404	{object}	ErrorResponse
//	@Router			/companies/{id} [delete]
func (c *Controller) DeleteCompany(context *gin.Context) {
	id, _ := uuid.Parse(context.Param("id"))

	if err := c.companies.Delete(id);
		err != nil {
		logger.Error(
			"Failed to delete company: %v",
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
