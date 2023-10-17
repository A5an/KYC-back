package providers

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/Sinbad-HQ/kyc/core/components/kyc/models"
)

const (
	incomeTransactionEvent = "income_transaction"
	pdfUploadedEvent       = "pdf_upload"
	incomeInsight          = "income_insight"
)

type CreditChekClient struct {
	baseURL   string
	publicKey string
}

func NewCreditChekClient(baseURL string, publicKey string) *CreditChekClient {
	return &CreditChekClient{
		baseURL:   baseURL,
		publicKey: publicKey,
	}
}

func (c *CreditChekClient) CreateLink(_, _, _ string) (string, error) {
	return "https://app.creditchek.africa/customer/onboarding?type=short&appId=5293878414&appLink=eFd1ZNdJda&app_id=64aac9d453a97b63508946e7&status=true", nil
}

type CallbackEvent struct {
	Event string                 `json:"event"`
	Data  map[string]interface{} `json:"data"`
}

type PdfUploadedEvent struct {
	Success       bool   `json:"success"`
	PageCount     int    `json:"pageCount"`
	AccountName   string `json:"accountName"`
	AccountNumber string `json:"accountNumber"`
	AccountType   string `json:"accountType"`
	BankName      string `json:"bankName"`
	BankCode      string `json:"bankCode"`
	PdfURL        string `json:"pdfUrl"`
	Bvn           string `json:"bvn"`
	BorrowerID    string `json:"borrowerId"`
}

type IncomeTransaction struct {
	AccountNumber       string    `json:"accountNumber"`
	AccountName         string    `json:"accountName"`
	Bvn                 string    `json:"bvn"`
	AppID               string    `json:"appId"`
	BorrowerID          string    `json:"borrowerId"`
	BusinessID          string    `json:"businessId"`
	Balance             float64   `json:"balance"`
	BankCode            string    `json:"bankCode"`
	BankName            string    `json:"bankName"`
	CreatedAt           int64     `json:"createdAt"`
	PdfURL              string    `json:"pdfUrl"`
	LastTransactionDate time.Time `json:"lastTransactionDate"`
	Success             bool      `json:"success"`
}

type IncomeInsight struct {
	Status  string `json:"status"`
	Message string `json:"message"`
	Data    struct {
		Edti struct {
			BusinessID                          string `json:"businessId"`
			BorrowerID                          string `json:"borrowerId"`
			AverageMonthlyIncome                string `json:"average_monthly_income"`
			AnnualEdti                          string `json:"annual_edti"`
			AverageMonthlyBalance               string `json:"average_monthly_balance"`
			AverageMonthlyRecurringDebtExpenses string `json:"average_monthly_recurring_debt_expenses"`
			NumberOfActiveMonths                string `json:"number_of_active_months"`
			AverageMonthlyEdti                  string `json:"average_monthly_edti"`
			DTIReason                           string `json:"DTI_reason"`
			Balance                             struct {
				August2022 int `json:"August 2022"`
			} `json:"balance"`
			Edti struct {
				August2022 int `json:"August 2022"`
			} `json:"EDTI"`
			RecurrentExpensesDebt struct {
				August2022 int `json:"August 2022"`
			} `json:"recurrent_expenses_debt"`
			RecurrentDebtSum      int    `json:"recurrent_debt_sum"`
			AverageMonthlyIncome0 string `json:"averageMonthlyIncome"`
			TotalMoneyReceive     string `json:"totalMoneyReceive"`
			TotalMoneySpent       string `json:"totalMoneySpent"`
			EligibleAmount        string `json:"eligibleAmount"`
			TotalBorrowed         string `json:"totalBorrowed"`
			TotalOutstanding      string `json:"totalOutstanding"`
			TotalOverdue          string `json:"totalOverdue"`
			CreditInsight         string `json:"creditInsight"`
			Salary                struct {
				Narration []interface{} `json:"narration"`
				Amount    []interface{} `json:"amount"`
			} `json:"salary"`
		} `json:"EDTI"`
	} `json:"data"`
}

func (c *CreditChekClient) getIncomeInsight(borrowerID string) (IncomeInsight, error) {
	var incomeInsight IncomeInsight

	url := fmt.Sprintf("%s/income/insight-data/%s", c.baseURL, borrowerID)
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return incomeInsight, err
	}
	req.Header.Set("token", c.publicKey)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return incomeInsight, err
	}
	defer resp.Body.Close()

	if err := json.NewDecoder(resp.Body).Decode(&incomeInsight); err != nil {
		return incomeInsight, err
	}

	return incomeInsight, nil
}

func (c *CreditChekClient) GetUserInfoFromCallback(req *http.Request) (models.UserInfo, []byte, error) {
	var (
		providerResponse []byte
		userInfo         models.UserInfo
		callbackEvent    CallbackEvent
	)

	err := json.NewDecoder(req.Body).Decode(&callbackEvent)
	if err != nil {
		return userInfo, nil, err
	}
	defer req.Body.Close()

	buf, err := json.Marshal(callbackEvent.Data)
	if err != nil {
		return userInfo, nil, err
	}

	switch callbackEvent.Event {
	case incomeTransactionEvent:
		var incomeTransaction IncomeTransaction
		err = json.Unmarshal(buf, &incomeTransaction)
		if err != nil {
			return userInfo, nil, err
		}

		if incomeTransaction.Success {
			userInfo.AccountBalance = &incomeTransaction.Balance
			userInfo.IDType = &incomeTransaction.Bvn
			userInfo.BankAccountNumber = &incomeTransaction.AccountNumber
			userInfo.KycID = incomeTransaction.Bvn
		}
	case pdfUploadedEvent:
		var pdfUploadedEvent PdfUploadedEvent
		err = json.Unmarshal(buf, &pdfUploadedEvent)
		if err != nil {
			return userInfo, nil, err
		}

		if pdfUploadedEvent.Success {
			insight, err := c.getIncomeInsight(pdfUploadedEvent.BorrowerID)
			if err != nil {
				log.Printf("failed to get income insight data for user: %s error: %v\n", pdfUploadedEvent.Bvn, err)
			}

			avgMonthlyIncome, err := strconv.ParseFloat(insight.Data.Edti.AverageMonthlyIncome, 64)
			if err != nil {
				log.Printf("failed to parse income insight data for user: %s error: %v\n", pdfUploadedEvent.Bvn, err)
			}

			userInfo.IDType = &pdfUploadedEvent.Bvn
			userInfo.KycID = pdfUploadedEvent.Bvn
			userInfo.BankAccountNumber = &pdfUploadedEvent.AccountNumber
			userInfo.ProviderResponse = &buf
			userInfo.AverageSalary = &avgMonthlyIncome
		}

	}

	return userInfo, providerResponse, nil
}
