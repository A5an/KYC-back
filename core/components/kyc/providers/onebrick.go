package providers

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"math"
	"net/http"
	"strconv"

	"github.com/spf13/viper"

	"github.com/Sinbad-HQ/kyc/core/components/kyc/models"
)

type OneBrickClient struct {
	baseURL      string
	clientID     string
	clientSecret string
}

func NewOneBrickClient(baseURL string, clientID string, clientSecret string) *OneBrickClient {
	return &OneBrickClient{
		baseURL:      baseURL,
		clientID:     clientID,
		clientSecret: clientSecret,
	}
}

type TokenResponse struct {
	Status  int    `json:"status"`
	Message string `json:"message"`
	Data    struct {
		AccessToken string `json:"access_token"`
	} `json:"data"`
}

type CallbackResponse struct {
	BankID      string `json:"bankId"`
	AccessToken string `json:"accessToken"`
	UserID      string `json:"userId"`
}

type AccountInfo struct {
	Data []struct {
		AccountHolder string `json:"accountHolder"`
		AccountID     string `json:"accountId"`
		AccountNumber string `json:"accountNumber"`
		Balances      struct {
			Available float64 `json:"available"`
			Current   float64 `json:"current"`
			Limit     float64 `json:"limit"`
		} `json:"balances"`
		Currency interface{} `json:"currency"`
	} `json:"data"`
	LastUpdateAt string `json:"lastUpdateAt"`
	Message      string `json:"message"`
	Session      string `json:"session"`
	Status       int    `json:"status"`
}

type AverageBalance struct {
	Data []struct {
		AccountID        string  `json:"accountId,omitempty"`
		AccountNumber    string  `json:"accountNumber"`
		AvgBalance       float64 `json:"avgBalance"`
		BeginningBalance float64 `json:"beginningBalance"`
		Currency         string  `json:"currency"`
		EndingBalance    float64 `json:"endingBalance"`
		InstitutionID    float64 `json:"institution_id"`
		Month            string  `json:"month"`
		Year             string  `json:"year"`
	} `json:"data"`
	Message string `json:"message"`
	Status  int    `json:"status"`
}

type IncomeInformation struct {
	Data []struct {
		BpjsCardNumber string `json:"bpjsCardNumber"`
		CompanyName    string `json:"companyName"`
		InstitutionID  int    `json:"institutionId"`
		MonthName      string `json:"monthName"`
		Salary         string `json:"salary"`
		Type           string `json:"type"`
	} `json:"data"`
	Message string `json:"message"`
	Status  int    `json:"status"`
}

const (
	noTokenInResponseMsg = "no token found in the response"
)

func request(baseURL, path, method, auth, accept string, a interface{}) error {
	client := &http.Client{}
	url := fmt.Sprintf("%s/%s", baseURL, path)
	req, err := http.NewRequest(method, url, nil)
	if err != nil {
		return err
	}
	req.Header.Add("Accept", accept)
	req.Header.Add("Authorization", auth)

	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		var errData map[string]interface{}
		err := json.NewDecoder(resp.Body).Decode(&errData)
		if err != nil {
			return err
		}
		return fmt.Errorf("error fetching %s: %v", url, errData["message"])
	}

	return json.NewDecoder(resp.Body).Decode(a)
}

func (c *OneBrickClient) fetchPublicAccessToken() (string, error) {
	auth := base64.StdEncoding.EncodeToString([]byte(fmt.Sprintf("%s:%s", c.clientID, c.clientSecret)))
	var tokenResponse TokenResponse

	err := request(c.baseURL, "auth/token", "GET", "Basic "+auth, "application/json", &tokenResponse)
	if err != nil {
		return "", err
	}

	if tokenResponse.Data.AccessToken == "" {
		return "", errors.New(noTokenInResponseMsg)
	}

	return tokenResponse.Data.AccessToken, nil
}

func (c *OneBrickClient) CreateLink(kycID string, _, _ string) (string, error) {
	token, err := c.fetchPublicAccessToken()
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("https://cdn.onebrick.io/sandbox-widget/v1/?accessToken=%s&redirect_url=%s&user_id=%s",
		token,
		viper.GetString("onebrick.redirect-url"),
		kycID,
	), nil
}

func (c *OneBrickClient) getUserAccountInformation(userAccessToken string) (AccountInfo, error) {
	var accInfo AccountInfo
	err := request(c.baseURL, "account/list", "GET", "Bearer "+userAccessToken, "application/json", &accInfo)
	if err != nil {
		return AccountInfo{}, err
	}

	return accInfo, nil
}

func (c *OneBrickClient) getEmploymentInfo(userAccessToken string) (employerName string, averageIncome float64, err error) {
	var incomeInfo IncomeInformation
	err = request(c.baseURL, "income/salary/", "GET", "Bearer "+userAccessToken, "application/json", &incomeInfo)
	if err != nil {
		return employerName, averageIncome, err
	}

	var totalSalary float64
	for _, income := range incomeInfo.Data {
		salary, err := strconv.ParseFloat(income.Salary, 64)
		if err != nil {
			return employerName, averageIncome, err
		}

		if income.CompanyName != "" {
			employerName = income.CompanyName
		}

		totalSalary += salary
	}

	return employerName, totalSalary / float64(len(incomeInfo.Data)), nil
}

func (c *OneBrickClient) parseCallbackResponse(req *http.Request) ([]CallbackResponse, error) {
	var resp []CallbackResponse
	err := json.NewDecoder(req.Body).Decode(&resp)
	if err != nil {
		return []CallbackResponse{}, err
	}
	return resp, nil
}

func (c *OneBrickClient) GetProviderCallback(req *http.Request) (models.ProviderCallback, error) {
	resp, err := c.parseCallbackResponse(req)
	if err != nil {
		return models.ProviderCallback{}, err
	}

	if len(resp) == 0 {
		return models.ProviderCallback{}, nil
	}

	var (
		bankInfo        *models.BankInfo
		kycSubmissionID = resp[0].UserID
		userAccessToken = resp[0].AccessToken
	)

	accountInfo, err := c.getUserAccountInformation(userAccessToken)
	if err != nil {
		log.Printf("fetching account information for kyc with ID: %s error: %s\n", kycSubmissionID, err)
		//return models.ProviderInfo{}, nil, err
	}

	if len(accountInfo.Data) > 0 {
		data := accountInfo.Data[0]
		bankInfo = &models.BankInfo{
			KycSubmissionID:   kycSubmissionID,
			AccountHolderName: data.AccountHolder,
			AccountNumber:     data.AccountNumber,
			AccountBalance:    data.Balances.Current,
		}
	}

	employerName, averageSalary, err := c.getEmploymentInfo(userAccessToken)
	if err != nil {
		log.Printf("fetching average salary for kyc with ID: %s error: %s\n", kycSubmissionID, err)
	}
	averageSalary = math.Round(averageSalary*100) / 100

	employmentInfo := models.EmploymentInfo{
		KycSubmissionID:  kycSubmissionID,
		EmployerName:     employerName,
		AverageSalary:    averageSalary,
		EmploymentStatus: averageSalary > 0,
	}

	return models.ProviderCallback{
		KycSubmissionID: kycSubmissionID,
		EmploymentInfo:  &employmentInfo,
		BankInfo:        bankInfo,
	}, nil
}
