package packages

import (
	"context"

	"github.com/Sinbad-HQ/kyc/core/components/packages/models"
	"github.com/Sinbad-HQ/kyc/core/components/risk_parameters"
	"github.com/Sinbad-HQ/kyc/core/components/usersession"
)

type Component interface {
	Create(ctx context.Context, product *models.Package) (*models.Package, error)
	GetByProviderID(ctx context.Context) ([]models.Package, error)
	GetByID(ctx context.Context, id string) (*models.Package, error)
	UpdateByID(ctx context.Context, id string, updatedPackage *models.Package) (*models.Package, error)
	DeleteByID(ctx context.Context, id string) error
}

type component struct {
	repo                   Repo
	userSessionComponent   usersession.Component
	riskParameterComponent risk_parameters.Component
}

func NewComponent(repo Repo, userSessionComponent usersession.Component, riskParameterComponent risk_parameters.Component) *component {
	return &component{
		repo:                   repo,
		userSessionComponent:   userSessionComponent,
		riskParameterComponent: riskParameterComponent,
	}
}

func (c *component) Create(ctx context.Context, kycPackage *models.Package) (*models.Package, error) {
	authCtx := c.userSessionComponent.GetAuthContextFromCtx(ctx)
	kycPackage.OrgID = authCtx.OrgID

	_, err := c.riskParameterComponent.GetByID(ctx, kycPackage.RiskParameterID)
	if err != nil {
		return nil, err
	}

	createdPackage, err := c.repo.Create(ctx, kycPackage)
	if err != nil {
		return nil, err
	}

	return createdPackage, nil
}

func (c *component) GetByProviderID(ctx context.Context) ([]models.Package, error) {
	authCtx := c.userSessionComponent.GetAuthContextFromCtx(ctx)
	return c.repo.GetByOrgID(ctx, authCtx.OrgID)
}

func (c *component) GetByID(ctx context.Context, id string) (*models.Package, error) {
	authCtx := c.userSessionComponent.GetAuthContextFromCtx(ctx)
	return c.repo.GetByID(ctx, id, authCtx.OrgID)
}

func (c *component) UpdateByID(ctx context.Context, id string, updatedPackage *models.Package) (*models.Package, error) {
	authCtx := c.userSessionComponent.GetAuthContextFromCtx(ctx)
	updatedPackage.OrgID = authCtx.OrgID

	_, err := c.riskParameterComponent.GetByID(ctx, updatedPackage.RiskParameterID)
	if err != nil {
		return nil, err
	}

	return c.repo.UpdateByID(ctx, id, updatedPackage)
}

func (c *component) DeleteByID(ctx context.Context, id string) error {
	authCtx := c.userSessionComponent.GetAuthContextFromCtx(ctx)
	return c.repo.DeleteByID(ctx, id, authCtx.OrgID)
}
