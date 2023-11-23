package providers

import (
	"encoding/json"
	"net/http"

	"github.com/Sinbad-HQ/kyc/core/components/kyc/models"
)

type OkraClient struct{}

type okraIncomeCallback struct {
	CustomerBvn string `json:"customerBvn"`
	Income      struct {
		OtherStreams struct {
			Details struct {
				Source struct {
					Type string `json:"type"`
				} `json:"source"`
			} `json:"details"`

			History struct {
				PastThreeMonths struct {
					AveragePerMonth float64 `json:"average_per_month"`
					MaxPerMonth     int     `json:"max_per_month"`
					MinPerMonth     int     `json:"min_per_month"`
					Occurrence      int     `json:"occurrence"`
					Total           int     `json:"total"`
				} `json:"past_three_months"`
			} `json:"history"`
		} `json:"other_streams"`
	} `json:"income"`
}

func NewOkraClient() *OkraClient { return &OkraClient{} }

func (c *OkraClient) CreateLink(_ string, _, _ string) (string, error) {
	return "https://app.okra.ng/LANg3W7CO", nil
}

func (c *OkraClient) GetProviderCallback(req *http.Request) (models.ProviderCallback, error) {
	var incomeCallback okraIncomeCallback
	err := json.NewDecoder(req.Body).Decode(&incomeCallback)
	if err != nil {
		return models.ProviderCallback{}, nil
	}
	defer req.Body.Close()

	averagePerMonth := incomeCallback.Income.OtherStreams.History.PastThreeMonths.AveragePerMonth
	employmentInfo := models.EmploymentInfo{
		AverageSalary:    averagePerMonth,
		EmploymentStatus: averagePerMonth > 0,
		EmployerName:     incomeCallback.Income.OtherStreams.Details.Source.Type,
	}

	return models.ProviderCallback{
		UserIDNumber:   incomeCallback.CustomerBvn,
		EmploymentInfo: &employmentInfo,
	}, nil
}
