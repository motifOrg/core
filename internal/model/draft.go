package model

import (
	"time"

	"github.com/google/uuid"
)

const (
	DraftStatusDraft    = "draft"
	DraftStatusReviewed = "reviewed"
)

type MessageDraft struct {
	ID            uuid.UUID  `gorm:"<-:create;type:uuid;default:gen_random_uuid();primaryKey" json:"id"`
	CompanyID     *uuid.UUID `gorm:"type:uuid" json:"company_id"`
	OpportunityID *uuid.UUID `gorm:"type:uuid" json:"opportunity_id"`

	Recipient string `json:"recipient"`
	Subject   string `json:"subject"`
	Body      string `json:"body"`
	Status    string `gorm:"default:draft" json:"status"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
