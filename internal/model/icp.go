package model

import (
	"time"

	"github.com/google/uuid"
	"github.com/lib/pq"
)

type ICP struct {
	ID         uuid.UUID `gorm:"<-:create;type:uuid;default:gen_random_uuid();primaryKey" json:"id"`
	CampaignID uuid.UUID `gorm:"type:uuid" json:"campaign_id"`

	Thesis          string         `json:"thesis"`
	Industries      pq.StringArray `gorm:"type:text[]" json:"industries"`
	Geographies     pq.StringArray `gorm:"type:text[]" json:"geographies"`
	BusinessModels  pq.StringArray `gorm:"type:text[]" json:"business_models"`
	Characteristics pq.StringArray `gorm:"type:text[]" json:"characteristics"`
	Exclusions      pq.StringArray `gorm:"type:text[]" json:"exclusions"`

	RevenueMin   *int64 `json:"revenue_min"`
	RevenueMax   *int64 `json:"revenue_max"`
	EbitdaMin    *int64 `json:"ebitda_min"`
	EmployeesMin *int   `json:"employees_min"`
	EmployeesMax *int   `json:"employees_max"`
	Ownership    string `json:"ownership"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func (ICP) TableName() string {
	return "icp_criteria"
}
