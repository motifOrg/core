package model

import (
	"time"

	"github.com/google/uuid"
)

const (
	OpportunityNew       = "new"
	OpportunityReviewing = "reviewing"
	OpportunityQualified = "qualified"
	OpportunityDeclined  = "declined"
)

type Opportunity struct {
	ID         uuid.UUID  `gorm:"<-:create;type:uuid;default:gen_random_uuid();primaryKey" json:"id"`
	CampaignID uuid.UUID  `gorm:"type:uuid" json:"campaign_id"`
	CompanyID  *uuid.UUID `gorm:"type:uuid" json:"company_id"`

	Source      string `gorm:"default:manual" json:"source"`
	ExternalRef string `json:"external_ref"`

	SenderName    string    `json:"sender_name"`
	SenderEmail   string    `json:"sender_email"`
	Subject       string    `json:"subject"`
	Body          string    `json:"body"`
	CompanyName   string    `json:"company_name"`
	CompanyDomain string    `json:"company_domain"`
	ReceivedAt    time.Time `json:"received_at"`
	IsSample      bool      `json:"is_sample"`

	FitScore   *int   `json:"fit_score"`
	FitSummary string `json:"fit_summary"`
	Priority   string `gorm:"default:medium" json:"priority"`
	Status     string `gorm:"default:new" json:"status"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func (Opportunity) TableName() string {
	return "inbound_opportunities"
}
