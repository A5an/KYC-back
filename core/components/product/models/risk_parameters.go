package models

type RiskParameter struct {
	ID               string  `json:"id"`
	Country          string  `json:"country"`
	AccountBalance   float64 `json:"account_balance"`
	AverageSalary    float64 `json:"average_salary"`
	EmploymentStatus bool    `json:"employment_status"`
	ProviderID       string  `json:"provider_id"`
}
