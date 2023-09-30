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
)

const (
	PendingStatus  = "pending"
	AprovedStatus  = "approved"
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
	CreateLink(kycID string) (string, error)
	GetUserInfoFromCallback(req *http.Request) (models.UserInfo, []byte, error)
}

type component struct {
	// mapping provider to country code for prototype (db table)
	providers            map[string]Provider
	repo                 Repo
	productComponent     product.Component
	userSessionComponent usersession.Component
}

func NewComponent(repo Repo, productComponent product.Component, userSessionComponent usersession.Component, providers map[string]Provider) *component {
	return &component{
		providers:            providers,
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
	kyc.Status = PendingStatus

	kyc.Nationality = strings.ToLower(kyc.Nationality)
	// hack: temporary for making things works before more design
	if kyc.Nationality != "nigeria" {
		kyc.ID = uuid.NewString()
	}

	provider, ok := c.providers[kyc.Nationality]
	if !ok {
		return nil, fmt.Errorf("user verification from %s is not current supported", kyc.Nationality)
	}

	link, err := provider.CreateLink(kyc.ID)
	if err != nil {
		return nil, err
	}
	kyc.Link = &link

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
		kyc.AccountBalance = &userInfo.AccountBalance
	}
	if kyc.AverageSalary == nil {
		kyc.AverageSalary = &userInfo.AverageSalary
	}
	if kyc.EmploymentStatus == nil {
		kyc.EmploymentStatus = &userInfo.EmploymentStatus
	}

	kyc.AccountBalanceRiskLevel = &HighRiskLevel
	if userInfo.AccountBalance >= riskParameter.AccountBalance {
		kyc.AccountBalanceRiskLevel = &LowRiskLevel
	}

	kyc.AverageSalaryRiskLevel = &HighRiskLevel
	if userInfo.AverageSalary >= riskParameter.AverageSalary {
		kyc.AverageSalaryRiskLevel = &LowRiskLevel
	}

	kyc.EmploymentRiskLevel = &LowRiskLevel
	if riskParameter.EmploymentStatus {
		if !userInfo.EmploymentStatus {
			kyc.EmploymentRiskLevel = &HighRiskLevel
		}
	}

	kyc.IdentityResponse = userInfo.ProviderResponse
	return c.repo.UpdateByID(ctx, kyc)
}
