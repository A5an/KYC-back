package models

import (
	"gorm.io/datatypes"

	"github.com/Sinbad-HQ/kyc/db/model"
)

type KycSubmission struct {
	model.Model
	PackageID      string            `json:"package_id"`
	Checklist      datatypes.JSONMap `json:"checklist"`
	Status         string            `json:"status"`
	UserInfo       UserInfo          `json:"user_info"`
	PassportInfo   PassportInfo      `json:"passport_info"`
	EmploymentInfo EmploymentInfo    `json:"employment_info"`
	BankInfo       BankInfo          `json:"bank_info"`
	AddressInfo    AddressInfo       `json:"address_info"`
	OrgID          string            `json:"org_id" gorm:"index"`
}

type UserInfo struct {
	model.Model
	KycSubmissionID string `json:"kyc_submission_id" gorm:"unique"`
	FirstName       string `json:"first_name"`
	LastName        string `json:"last_name"`
	Nationality     string `json:"nationality"`
	Address         string `json:"address"`
	Email           string `json:"email"`
	PhoneNumber     string `json:"phone_number"`
	SignatureLink   string `json:"signature_link"`
	ImageLink       string `json:"image_logo"`
	IDNumber        string `json:"id_number" gorm:"index"`
}

type PassportInfo struct {
	KycSubmissionID   string `json:"kyc_submission_id" gorm:"unique"`
	FullName          string `json:"full_name"`
	PassportNumber    string `json:"passport_number"`
	Status            string `json:"status"`
	Sex               string `json:"sex"`
	Nationality       string `json:"nationality"`
	DateOfBirth       string `json:"date_of_birth"`
	PlaceOfBirth      string `json:"place_of_birth"`
	Authority         string `json:"authority"`
	IssuedDate        string `json:"issued_date"`
	ExpiryDate        string `json:"expiry_date"`
	AgeEstimate       string `json:"age_estimate"`
	FaceMatch         bool   `json:"face_match"`
	PassportFrontLink string `json:"passport_front_link"`
	PassportFaceLink  string `json:"passport_face_link"`
}

type EmploymentInfo struct {
	KycSubmissionID        string         `json:"kyc_submission_id" gorm:"unique"`
	EmployerName           string         `json:"employer_name"`
	AverageSalary          float64        `json:"average_salary"`
	AverageSalaryRiskLevel string         `json:"average_salary_level"`
	EmploymentRiskLevel    string         `json:"employment_risk_level"`
	EmploymentLetterLink   string         `json:"employment_letter_link"`
	EmploymentStatus       bool           `json:"employment_status"`
	ProviderResponse       datatypes.JSON `json:"-"`
}

type BankInfo struct {
	KycSubmissionID         string         `json:"kyc_submission_id" gorm:"unique"`
	AccountHolderName       string         `json:"account_holder_name"`
	BankName                string         `json:"bank_name"`
	AccountNumber           string         `json:"account_number"`
	AccountBalance          float64        `json:"account_balance"`
	AccountBalanceRiskLevel string         `json:"account_balance_risk_level"`
	BankStatementLink       string         `json:"bank_statement_link"`
	ProviderResponse        datatypes.JSON `json:"-"`
}

type AddressInfo struct {
	KycSubmissionID string `json:"kyc_submission_id" gorm:"unique"`
	Address         string `json:"address"`
	UtilityBillLink string `json:"utility_bill_link"`
}

type ProviderCallback struct {
	KycSubmissionID string
	// for providers with no support for kyc submission id. eg credit check
	UserIDNumber   string
	PassportInfo   *PassportInfo
	EmploymentInfo *EmploymentInfo
	BankInfo       *BankInfo
	AddressInfo    *AddressInfo
}
