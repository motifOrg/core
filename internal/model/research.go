package model

import (
	"time"

	"github.com/google/uuid"
)

const (
	FindingFit        = "fit"
	FindingGeneral    = "finding"
	FindingRisk       = "risk"
	FindingMissing    = "missing"
	FindingNextAction = "next_action"
)

const (
	EvidenceSample     = "sample"
	EvidenceUnverified = "unverified"
	EvidenceVerified   = "verified"
)

const (
	SenderAI     = "ai"
	SenderUser   = "user"
	SenderSystem = "system"
)

type Finding struct {
	ID        uuid.UUID `gorm:"<-:create;type:uuid;default:gen_random_uuid();primaryKey" json:"id"`
	CompanyID uuid.UUID `gorm:"type:uuid" json:"company_id"`

	Kind           string `json:"kind"`
	Title          string `json:"title"`
	Content        string `json:"content"`
	SourceURL      string `gorm:"column:source_url" json:"source_url"`
	Confidence     string `gorm:"default:medium" json:"confidence"`
	EvidenceStatus string `gorm:"default:unverified" json:"evidence_status"`

	CreatedAt time.Time `json:"created_at"`
}

func (Finding) TableName() string {
	return "research_findings"
}

type Message struct {
	ID        uuid.UUID `gorm:"<-:create;type:uuid;default:gen_random_uuid();primaryKey" json:"id"`
	CompanyID uuid.UUID `gorm:"type:uuid" json:"company_id"`

	Body       string `json:"body"`
	SenderRole string `json:"sender_role"`

	CreatedAt time.Time `json:"created_at"`
}

func (Message) TableName() string {
	return "research_messages"
}
