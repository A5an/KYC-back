package models

import (
	"github.com/Sinbad-HQ/kyc/core/components/kyc/models"
	"github.com/Sinbad-HQ/kyc/db/model"
)

type Package struct {
	model.Model
	Name            string `json:"name"`
	Description     string `json:"description"`
	LogoURL         string `json:"logo_url"`
	RiskParameterID string `json:"risk_parameter_id"`
	KycSubmissions  []models.KycSubmission
	OrgID           string `json:"org_id" gorm:"index"`
}
