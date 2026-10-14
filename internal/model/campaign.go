package model

import (
	"time"

	"github.com/google/uuid"
)

const (
	CampaignStatusDraft    = "draft"
	CampaignStatusActive   = "active"
	CampaignStatusPaused   = "paused"
	CampaignStatusArchived = "archived"
)

const (
	WorkflowOutbound = "outbound"
	WorkflowInbound  = "inbound"
	WorkflowBoth     = "both"
)

type Campaign struct {
	ID uuid.UUID `gorm:"<-:create;type:uuid;default:gen_random_uuid();primaryKey" json:"id"`

	Name        string `json:"name"`
	Description string `json:"description"`
	Status      string `gorm:"default:draft" json:"status"`
	Workflow    string `gorm:"default:outbound" json:"workflow"`

	ICP *ICPCriteria `gorm:"foreignKey:CampaignID" json:"icp,omitempty"`

	CompanyCount     int64 `gorm:"->;-:migration" json:"company_count"`
	OpportunityCount int64 `gorm:"->;-:migration" json:"opportunity_count"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
