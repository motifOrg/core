package model

import (
	"time"

	"github.com/google/uuid"
)

const (
	SourceSample  = "sample"
	SourceCSV     = "csv"
	SourceManual  = "manual"
	SourceInbound = "inbound"
	SourceGmail   = "gmail"
)

const (
	PriorityHigh   = "high"
	PriorityMedium = "medium"
	PriorityLow    = "low"
)

const (
	ResearchNotStarted = "not_started"
	ResearchInProgress = "in_progress"
	ResearchComplete   = "complete"
)

const (
	EvaluationNew          = "new"
	EvaluationShortlisted  = "shortlisted"
	EvaluationContacted    = "contacted"
	EvaluationInDiscussion = "in_discussion"
	EvaluationRejected     = "rejected"
)

type Company struct {
	ID         uuid.UUID `gorm:"<-:create;type:uuid;default:gen_random_uuid();primaryKey" json:"id"`
	CampaignID uuid.UUID `gorm:"type:uuid" json:"campaign_id"`

	Name            string `json:"name"`
	Domain          string `json:"domain"`
	Industry        string `json:"industry"`
	Location        string `json:"location"`
	Description     string `json:"description"`
	Employees       *int   `json:"employees"`
	RevenueEstimate string `json:"revenue_estimate"`
	FoundedYear     *int   `json:"founded_year"`
	OwnerName       string `json:"owner_name"`

	Source   string `gorm:"default:manual" json:"source"`
	IsSample bool   `json:"is_sample"`

	FitScore         *int   `json:"fit_score"`
	FitSummary       string `json:"fit_summary"`
	Priority         string `gorm:"default:medium" json:"priority"`
	ResearchStatus   string `gorm:"default:not_started" json:"research_status"`
	EvaluationStatus string `gorm:"default:new" json:"evaluation_status"`

	Findings []ResearchFinding `gorm:"foreignKey:CompanyID" json:"findings,omitempty"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
