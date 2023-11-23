package risk_parameters

import (
	"context"

	"github.com/Sinbad-HQ/kyc/core/components/risk_parameters/models"
	"github.com/Sinbad-HQ/kyc/core/components/usersession"
)

type Component interface {
	Create(ctx context.Context, riskParameter *models.RiskParameter) (*models.RiskParameter, error)
	GetByProviderID(ctx context.Context) ([]models.RiskParameter, error)
	GetByID(ctx context.Context, id string) (*models.RiskParameter, error)
	UpdateByID(ctx context.Context, id string, updatedRiskParameter *models.RiskParameter) (*models.RiskParameter, error)
	DeleteByID(ctx context.Context, id string) error
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

func (c *component) Create(ctx context.Context, riskParameter *models.RiskParameter) (*models.RiskParameter, error) {
	authCtx := c.userSessionComponent.GetAuthContextFromCtx(ctx)
	riskParameter.OrgID = authCtx.OrgID

	createdProduct, err := c.repo.Create(ctx, riskParameter)
	if err != nil {
		return nil, err
	}

	return createdProduct, nil
}

func (c *component) GetByProviderID(ctx context.Context) ([]models.RiskParameter, error) {
	authCtx := c.userSessionComponent.GetAuthContextFromCtx(ctx)
	return c.repo.GetByOrgID(ctx, authCtx.OrgID)
}

func (c *component) GetByID(ctx context.Context, id string) (*models.RiskParameter, error) {
	authCtx := c.userSessionComponent.GetAuthContextFromCtx(ctx)
	return c.repo.GetByID(ctx, id, authCtx.OrgID)
}

func (c *component) UpdateByID(ctx context.Context, id string, updatedRiskParameter *models.RiskParameter) (*models.RiskParameter, error) {
	authCtx := c.userSessionComponent.GetAuthContextFromCtx(ctx)
	updatedRiskParameter.OrgID = authCtx.OrgID
	return c.repo.UpdateByID(ctx, id, updatedRiskParameter)
}

func (c *component) DeleteByID(ctx context.Context, id string) error {
	authCtx := c.userSessionComponent.GetAuthContextFromCtx(ctx)
	return c.repo.DeleteByID(ctx, id, authCtx.OrgID)
}
