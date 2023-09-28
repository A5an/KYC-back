package product

import (
	"context"
	"fmt"
	"strings"

	"github.com/google/uuid"

	"github.com/Sinbad-HQ/kyc/core/components/product/models"
	"github.com/Sinbad-HQ/kyc/core/components/usersession"
)

var (
	SupportedCountries = []string{"nigeria", "indonesia"}
)

type Component interface {
	Create(ctx context.Context, product *models.Product) (*models.Product, error)
	GetByProviderID(ctx context.Context) ([]models.Product, error)
	GetByID(ctx context.Context, id string, providerID string) (*models.Product, error)

	CreateRiskParameter(ctx context.Context, riskParameter *models.RiskParameter) (*models.RiskParameter, error)
	GetRiskParameters(ctx context.Context) ([]models.RiskParameter, error)
	GetRiskParameterByCountry(ctx context.Context, country string) (*models.RiskParameter, error)
}

type component struct {
	repo                 Repo
	userSessionComponent usersession.Component
}

func NewComponent(repo Repo, userSessionComponent usersession.Component) *component {
	return &component{
		repo:                 repo,
		userSessionComponent: userSessionComponent,
	}
}

func (c *component) Create(ctx context.Context, product *models.Product) (*models.Product, error) {
	authCtx := c.userSessionComponent.GetAuthContextFromCtx(ctx)
	product.ID = uuid.NewString()
	product.ProviderID = authCtx.ProviderID

	for _, country := range SupportedCountries {
		_, err := c.repo.GetRiskParameterByCountry(ctx, country)
		if err != nil {
			return nil, fmt.Errorf("please create: %s risk-parameter before creating product", country)
		}
	}

	createdProduct, err := c.repo.Create(ctx, product)
	if err != nil {
		return nil, err
	}

	return createdProduct, nil
}

func (c *component) GetByProviderID(ctx context.Context) ([]models.Product, error) {
	authCtx := c.userSessionComponent.GetAuthContextFromCtx(ctx)
	return c.repo.GetByProviderID(ctx, authCtx.ProviderID)
}

func (c *component) GetByID(ctx context.Context, id string, providerID string) (*models.Product, error) {
	//authCtx := c.userSessionComponent.GetAuthContextFromCtx(ctx)
	return c.repo.GetByID(ctx, id, providerID)
}

func (c *component) CreateRiskParameter(ctx context.Context, riskParameter *models.RiskParameter) (*models.RiskParameter, error) {
	authCtx := c.userSessionComponent.GetAuthContextFromCtx(ctx)
	riskParameter.ProviderID = authCtx.ProviderID
	riskParameter.ID = uuid.NewString()
	riskParameter.Country = strings.ToLower(riskParameter.Country)
	return c.repo.CreateRiskParameter(ctx, riskParameter)
}

func (c *component) GetRiskParameters(ctx context.Context) ([]models.RiskParameter, error) {
	authCtx := c.userSessionComponent.GetAuthContextFromCtx(ctx)
	return c.repo.GetRiskParameters(ctx, authCtx.ProviderID)
}

func (c *component) GetRiskParameterByCountry(ctx context.Context, country string) (*models.RiskParameter, error) {
	return c.repo.GetRiskParameterByCountry(ctx, country)
}
