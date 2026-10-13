package model

import (
	"time"

	"github.com/google/uuid"
)

const (
	DocumentStatusUploaded   = "uploaded"
	DocumentStatusProcessing = "processing"
	DocumentStatusProcessed  = "processed"
	DocumentStatusFailed     = "failed"
)

type Document struct {
	ID         uuid.UUID `gorm:"<-:create;type:uuid;default:gen_random_uuid();primaryKey" json:"id"`
	CampaignID uuid.UUID `gorm:"type:uuid" json:"campaign_id"`

	Filename   string `json:"filename"`
	StorageKey string `json:"-"`
	MimeType   string `json:"mime_type"`
	SizeBytes  int64  `json:"size_bytes"`
	Status     string `gorm:"default:uploaded" json:"status"`

	CreatedAt time.Time `json:"created_at"`
}

func (Document) TableName() string {
	return "documents"
}
