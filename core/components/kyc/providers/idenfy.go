package providers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/Sinbad-HQ/kyc/core/components/kyc/models"
)

type IdenfyClient struct {
	baseURL   string
	apiKey    string
	apiSecret string
}

func NewIdenfyClient(baseURL string, apiKey string, apiSecret string) *IdenfyClient {
	return &IdenfyClient{
		baseURL:   baseURL,
		apiKey:    apiKey,
		apiSecret: apiSecret,
	}
}

type CreateVerificationSession struct {
	ClientID  string   `json:"clientId"`
	FirstName string   `json:"firstName,omitempty"`
	LastName  string   `json:"lastName,omitempty"`
	Documents []string `json:"documents"`
}

type VerificationSession struct {
	Message       string   `json:"message"`
	AuthToken     string   `json:"authToken"`
	ScanRef       string   `json:"scanRef"`
	ClientID      string   `json:"clientId"`
	PersonScanRef string   `json:"personScanRef"`
	FirstName     string   `json:"firstName"`
	LastName      string   `json:"lastName"`
	Documents     []string `json:"documents"`
}

type VerificationCallback struct {
	Final    bool   `json:"final"`
	Platform string `json:"platform"`
	Status   struct {
		Overall          string      `json:"overall"`
		SuspicionReasons []string    `json:"suspicionReasons"`
		DenyReasons      []string    `json:"denyReasons"`
		FraudTags        []string    `json:"fraudTags"`
		MismatchTags     []string    `json:"mismatchTags"`
		AutoFace         string      `json:"autoFace"`
		ManualFace       interface{} `json:"manualFace"`
		AutoDocument     string      `json:"autoDocument"`
	} `json:"status"`
	Data struct {
		DocFirstName        string `json:"docFirstName"`
		DocLastName         string `json:"docLastName"`
		DocNumber           string `json:"docNumber"`
		DocExpiry           string `json:"docExpiry"`
		DocNationality      string `json:"docNationality"`
		DocIssuingCountry   string `json:"docIssuingCountry"`
		ManuallyDataChanged bool   `json:"manuallyDataChanged"`
		FullName            string `json:"fullName"`
		SelectedCountry     string `json:"selectedCountry"`
	} `json:"data"`
	FileUrls struct {
		Face        string `json:"FACE"`
		Front       string `json:"FRONT"`
		UtilityBill string `json:"UTILITY_BILL"`
	} `json:"fileUrls"`
	AdditionalStepPdfUrls struct {
	} `json:"additionalStepPdfUrls"`
	ScanRef     string      `json:"scanRef"`
	ExternalRef interface{} `json:"externalRef"`
	ClientID    string      `json:"clientId"`
}

func (c *IdenfyClient) CreateLink(kycID string, firstName string, lastName string) (string, error) {
	payload, err := json.Marshal(CreateVerificationSession{
		ClientID:  kycID,
		FirstName: firstName,
		LastName:  lastName,
		Documents: []string{"PASSPORT"},
	})
	if err != nil {
		return "", err
	}

	req, err := http.NewRequest(http.MethodPost, c.baseURL+"/api/v2/token", bytes.NewBuffer(payload))
	if err != nil {
		return "", err
	}
	req.SetBasicAuth(c.apiKey, c.apiSecret)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		return "", fmt.Errorf("failed to create new Idenfy verification session: status code %d", resp.StatusCode)
	}

	var verificationSession VerificationSession
	if err := json.NewDecoder(resp.Body).Decode(&verificationSession); err != nil {
		return "", err
	}

	return fmt.Sprintf("%s/api/v2/redirect?authToken=%s", c.baseURL, verificationSession.AuthToken), nil
}

func (c *IdenfyClient) GetUserInfoFromCallback(req *http.Request) (models.UserInfo, []byte, error) {
	var resp VerificationCallback
	err := json.NewDecoder(req.Body).Decode(&resp)
	if err != nil {
		return models.UserInfo{}, nil, err
	}

	status := resp.Status.Overall
	if strings.ToLower(resp.Status.Overall) == "approved" {
		status += "|" + resp.Status.AutoFace + "|" + resp.Status.AutoDocument
	} else {
		if len(resp.Status.DenyReasons) > 0 {
			status += "|" + strings.Join(resp.Status.DenyReasons, "|")
		}
		if len(resp.Status.SuspicionReasons) > 0 {
			status += "|" + strings.Join(resp.Status.SuspicionReasons, "|")
		}
	}

	userInfo := models.UserInfo{
		PassportStatus: &status,
		PassportNumber: &resp.Data.DocNumber,
		KycID:          resp.ClientID,
		ImageURL:       &resp.FileUrls.Face,
	}
	return userInfo, nil, nil
}
