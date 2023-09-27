package providers

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/Sinbad-HQ/kyc/core/components/kyc/models"
)

const (
	incomeTransactionEvent = "income_transaction"
	pdfUploadedEvent       = "pdf_upload"
)

type CreditChekClient struct{}

func NewCreditChekClient() *CreditChekClient { return &CreditChekClient{} }

func (c *CreditChekClient) CreateLink(_ string) (string, error) {
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
			userInfo.AccountBalance = incomeTransaction.Balance
			userInfo.KycID = incomeTransaction.Bvn
		}
	case pdfUploadedEvent:
		var pdfUploadedEvent PdfUploadedEvent
		err = json.Unmarshal(buf, &pdfUploadedEvent)
		if err != nil {
			return userInfo, nil, err
		}

		if pdfUploadedEvent.Success {
			userInfo.KycID = pdfUploadedEvent.Bvn
			userInfo.ProviderResponse = buf
		}
	}

	return userInfo, providerResponse, nil
}
