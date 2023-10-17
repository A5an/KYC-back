package kyc

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/google/uuid"

	"github.com/Sinbad-HQ/kyc/core/components/kyc/models"
	"github.com/Sinbad-HQ/kyc/core/components/product"
	"github.com/Sinbad-HQ/kyc/core/components/usersession"
	"github.com/Sinbad-HQ/kyc/notifier"
)

const (
	QueStatus      = "que"
	ApprovedStatus = "approved"
	RejectedStatus = "rejected"
)

var (
	LowRiskLevel  = "Low Risk"
	HighRiskLevel = "High Risk"
)

type Component interface {
	Create(ctx context.Context, kyc *models.Kyc) (*models.Kyc, error)
	GetByProductID(ctx context.Context, productID string) ([]models.Kyc, error)
	GetByID(ctx context.Context, id string) (*models.Kyc, error)
	UpdateStatusByID(ctx context.Context, kycID string, status string) error
	UpdateByID(ctx context.Context, userInfo *models.UserInfo) error
}

type Provider interface {
	CreateLink(kycID string, firstName string, lastName string) (string, error)
	GetUserInfoFromCallback(req *http.Request) (models.UserInfo, []byte, error)
}

type component struct {
	// mapping provider to country code for prototype (db table)
	providers            map[string]Provider
	passportProvider     Provider
	repo                 Repo
	productComponent     product.Component
	userSessionComponent usersession.Component
}

func NewComponent(repo Repo, productComponent product.Component,
	userSessionComponent usersession.Component,
	providers map[string]Provider,
	passportProvider Provider) *component {
	return &component{
		providers:            providers,
		passportProvider:     passportProvider,
		repo:                 repo,
		productComponent:     productComponent,
		userSessionComponent: userSessionComponent,
	}
}

func (c *component) Create(ctx context.Context, kyc *models.Kyc) (*models.Kyc, error) {
	_, err := c.productComponent.GetByID(ctx, kyc.ProductID, kyc.ProviderID)
	if err != nil {
		return nil, err
	}

	//authCtx := c.userSessionComponent.GetAuthContextFromCtx(ctx)
	//kyc.ProviderID = authCtx.ProviderID
	kyc.Status = QueStatus

	kyc.Nationality = strings.ToLower(kyc.Nationality)
	// hack: temporary for making things works before more design
	if kyc.Nationality != "nigeria" {
		kyc.ID = uuid.NewString()
	}

	provider, ok := c.providers[kyc.Nationality]
	if !ok {
		return nil, fmt.Errorf("user verification from %s is not current supported", kyc.Nationality)
	}

	generalVerificationLink, err := provider.CreateLink(kyc.ID, kyc.FirstName, kyc.LastName)
	if err != nil {
		return nil, err
	}
	kyc.GeneralVerificationLink = &generalVerificationLink

	passportVerificationLink, err := c.passportProvider.CreateLink(kyc.ID, kyc.FirstName, kyc.LastName)
	if err != nil {
		return nil, err
	}
	kyc.PassportVerificationLink = &passportVerificationLink

	return c.repo.Create(ctx, kyc)
}

func (c *component) GetByProductID(ctx context.Context, productID string) ([]models.Kyc, error) {
	authCtx := c.userSessionComponent.GetAuthContextFromCtx(ctx)
	return c.repo.GetByProductID(ctx, authCtx.ProviderID, productID)
}

func (c *component) GetByID(ctx context.Context, id string) (*models.Kyc, error) {
	authCtx := c.userSessionComponent.GetAuthContextFromCtx(ctx)
	return c.repo.GetByIDAndProviderID(ctx, id, authCtx.ProviderID)
}

func (c *component) UpdateStatusByID(ctx context.Context, kycID string, status string) error {
	authCtx := c.userSessionComponent.GetAuthContextFromCtx(ctx)
	kyc, err := c.repo.GetByIDAndProviderID(ctx, kycID, authCtx.ProviderID)
	if err != nil {
		return err
	}
	kyc.Status = strings.ToLower(status)

	product, err := c.productComponent.GetByID(ctx, kyc.ProductID, authCtx.ProviderID)
	if err != nil {
		return err
	}

	var body string
	switch kyc.Status {
	case ApprovedStatus:
		body = "approved"
	case RejectedStatus:
		body = "rejected"
	}

	if err := notifier.SendEmailNotification(
		[]string{kyc.Email},
		fmt.Sprintf("KYC Check - %s", product.Name),
		body,
		true,
	); err != nil {
		return err
	}

	return c.repo.UpdateByID(ctx, kyc)
}

func (c *component) UpdateByID(ctx context.Context, userInfo *models.UserInfo) error {
	kyc, err := c.repo.GetByID(ctx, userInfo.KycID)
	if err != nil {
		return err
	}

	riskParameter, err := c.productComponent.GetRiskParameterByCountry(ctx, kyc.Nationality)
	if err != nil {
		return err
	}

	if kyc.AccountBalance == nil {
		kyc.AccountBalance = userInfo.AccountBalance
	}
	if kyc.AverageSalary == nil {
		kyc.AverageSalary = userInfo.AverageSalary
	}
	if kyc.EmploymentStatus == nil {
		kyc.EmploymentStatus = userInfo.EmploymentStatus
	}
	if kyc.PassportVerificationStatus == nil {
		kyc.PassportVerificationStatus = userInfo.PassportStatus
	}
	if kyc.PassportNumber == nil {
		kyc.PassportNumber = userInfo.PassportNumber
	}
	if kyc.ImageURL == nil {
		kyc.ImageURL = userInfo.ImageURL
	}

	kyc.AccountBalanceRiskLevel = &HighRiskLevel
	if userInfo.AccountBalance != nil && *userInfo.AccountBalance >= riskParameter.AccountBalance {
		kyc.AccountBalanceRiskLevel = &LowRiskLevel
	}

	kyc.AverageSalaryRiskLevel = &HighRiskLevel
	if userInfo.AverageSalary != nil && *userInfo.AverageSalary >= riskParameter.AverageSalary {
		kyc.AverageSalaryRiskLevel = &LowRiskLevel
	}

	kyc.EmploymentRiskLevel = &LowRiskLevel
	if riskParameter.EmploymentStatus {
		if userInfo.EmploymentStatus != nil && !*userInfo.EmploymentStatus {
			kyc.EmploymentRiskLevel = &HighRiskLevel
		}
	}

	if userInfo.ProviderResponse != nil {
		kyc.IdentityResponse = *userInfo.ProviderResponse
	}
	kyc.IDType = userInfo.IDType
	kyc.BankVerificationNumber = userInfo.BankAccountNumber

	return c.repo.UpdateByID(ctx, kyc)
}
