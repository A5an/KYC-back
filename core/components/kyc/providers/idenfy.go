package providers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/Sinbad-HQ/kyc/core"
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
	ClientID    string   `json:"clientId"`
	FirstName   string   `json:"firstName,omitempty"`
	LastName    string   `json:"lastName,omitempty"`
	Documents   []string `json:"documents"`
	UtilityBill bool     `json:"utilityBill"`
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
		DocDob              string `json:"docDob"`
		DocDateOfIssue      string `json:"docDateOfIssue"`
		DocSex              string `json:"docSex"`
		BirthPlace          string `json:"birthPlace"`
		AgeEstimate         string `json:"ageEstimate"`
		Authority           string `json:"authority"`
		ManuallyDataChanged bool   `json:"manuallyDataChanged"`
		FullName            string `json:"fullName"`
		SelectedCountry     string `json:"selectedCountry"`
		Address             string `json:"address"`
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
		ClientID:    kycID,
		FirstName:   firstName,
		LastName:    lastName,
		Documents:   []string{"PASSPORT"},
		UtilityBill: true,
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
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return "", err
		}

		return "", fmt.Errorf("failed to create new Idenfy verification session: status code %d:%s", resp.StatusCode, string(body))
	}

	var verificationSession VerificationSession
	if err := json.NewDecoder(resp.Body).Decode(&verificationSession); err != nil {
		return "", err
	}

	return fmt.Sprintf("%s/api/v2/redirect?authToken=%s", c.baseURL, verificationSession.AuthToken), nil
}

func (c *IdenfyClient) GetProviderCallback(req *http.Request) (models.ProviderCallback, error) {
	var resp VerificationCallback
	err := json.NewDecoder(req.Body).Decode(&resp)
	if err != nil {
		return models.ProviderCallback{}, err
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

	var faceMatch bool
	if resp.Status.AutoFace == "FACE_MATCH" {
		faceMatch = true
	}

	var passportFrontUrl string
	fileLink, err := core.UploadFile(resp.ClientID+"-passport-front", resp.FileUrls.Front)
	if err != nil {
		passportFrontUrl = resp.FileUrls.Front
	}
	passportFrontUrl = fileLink

	var passportFaceUrl string
	fileLink, err = core.UploadFile(resp.ClientID+"-passport-face", resp.FileUrls.Face)
	if err != nil {
		passportFaceUrl = resp.FileUrls.Face
	}
	passportFaceUrl = fileLink

	var utilityBillUrl string
	fileLink, err = core.UploadFile(resp.ClientID+"-utility-bill", resp.FileUrls.UtilityBill)
	if err != nil {
		utilityBillUrl = resp.FileUrls.UtilityBill
	}
	utilityBillUrl = fileLink

	return models.ProviderCallback{
		PassportInfo: &models.PassportInfo{
			KycSubmissionID:   resp.ClientID,
			FullName:          resp.Data.FullName,
			PassportNumber:    resp.Data.DocNumber,
			Status:            status,
			Sex:               resp.Data.DocSex,
			Nationality:       resp.Data.DocNationality,
			DateOfBirth:       resp.Data.DocDob,
			PlaceOfBirth:      resp.Data.BirthPlace,
			Authority:         resp.Data.Authority,
			IssuedDate:        resp.Data.DocDateOfIssue,
			ExpiryDate:        resp.Data.DocExpiry,
			AgeEstimate:       resp.Data.AgeEstimate,
			FaceMatch:         faceMatch,
			PassportFrontLink: passportFrontUrl,
			PassportFaceLink:  passportFaceUrl,
		},
		AddressInfo: &models.AddressInfo{
			KycSubmissionID: resp.ClientID,
			Address:         resp.Data.Address,
			UtilityBillLink: utilityBillUrl,
		},
	}, nil
}
