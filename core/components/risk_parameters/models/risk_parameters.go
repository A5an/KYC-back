package models

import (
	"github.com/Sinbad-HQ/kyc/core/components/packages/models"
	"github.com/Sinbad-HQ/kyc/db/model"
)

type RiskParameter struct {
	model.Model
	Name             string           `json:"name" gorm:"unique"`
	AccountBalance   float64          `json:"account_balance"`
	AverageSalary    float64          `json:"average_salary"`
	EmploymentStatus *bool            `json:"employment_status"`
	Packages         []models.Package `json:"-"`
	OrgID            string           `json:"org_id" gorm:"index"`
}
