package models

import (
	"github.com/jmoiron/sqlx/types"
)

type Kyc struct {
	// TODO: custom type later to cover null/nil type conversion
	// embedded tables for proto-type only
	ID         string  `json:"id"`
	Link       *string `json:"link"`
	ProductID  string  `json:"product_id"`
	ProviderID string  `json:"provider_id"`

	// Embed personal details in the same table for prototype
	FirstName  string `json:"first_name"`
	MiddleName string `json:"middle_name"`
	LastName   string `json:"last_name"`
	DoB        string `json:"dob"`
	Country    string `json:"country"`
	Gender     string `json:"gender"`

	// kyc metrics
	AccountBalance   *float64 `json:"account_balance"` //last 3 months
	EmploymentStatus *bool    `json:"employment_status"`
	AverageSalary    *float64 `json:"average_salary"`

	// identity response
	BankVerificationNumber *string        `json:"bank_verification_number"`
	IDType                 *string        `json:"id_type"`
	MobileNumber           *string        `json:"mobile_number"`
	IdentityResponse       types.JSONText `json:"identity_response"`

	Status string `json:"status"`

	// risk levels embedded
	AccountBalanceRiskLevel *string `json:"account_balance_risk_level"`
	AverageSalaryRiskLevel  *string `json:"average_salary_risk_level"`
	EmploymentRiskLevel     *string `json:"employment_risk_level"`
}

type UserInfo struct {
	// bank details
	AccountBalance float64
	AverageSalary  float64

	// employment details
	EmploymentStatus bool

	// kycprovider response
	ProviderResponse []byte

	// personal details
	KycID string
}
