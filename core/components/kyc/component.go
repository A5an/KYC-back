package kyc

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"strings"

	"github.com/google/uuid"

	"github.com/Sinbad-HQ/kyc/core/components/kyc/models"
	"github.com/Sinbad-HQ/kyc/core/components/packages"
	"github.com/Sinbad-HQ/kyc/core/components/usersession"
	"github.com/Sinbad-HQ/kyc/notifier"
)

const (
	QueueStatus    = "queue"
	AcceptedStatus = "accepted"
	RejectedStatus = "rejected"
	LowRiskLevel   = "Low Risk"
	HighRiskLevel  = "High Risk"

	Nigeria   = "nigeria"
	Indonesia = "indonesia"

	IdenfyProvider      = "idenfy"
	OneBrickProvider    = "onebrick"
	CreditCheckProvider = "creditcheck"
	OkraProvider        = "okra"
)

type Component interface {
	Create(ctx context.Context, kyc *models.KycSubmission) (*models.KycSubmission, error)
	GetByProductID(ctx context.Context, productID string) ([]models.KycSubmission, error)
	GetByOrgID(ctx context.Context) ([]models.KycSubmission, error)
	GetByID(ctx context.Context, id string, productID string) (*models.KycSubmission, error)
	UpdateStatusByID(ctx context.Context, id string, productID string, status string) error
	UpdateByProviderInfo(ctx context.Context, providerCallback *models.ProviderCallback) error
}

type Provider interface {
	CreateLink(kycID string, firstName string, lastName string) (string, error)
	GetProviderCallback(req *http.Request) (models.ProviderCallback, error)
}

type component struct {
	providers            map[string]Provider
	repo                 Repo
	productComponent     packages.Component
	userSessionComponent usersession.Component
}

func NewComponent(repo Repo, productComponent packages.Component,
	userSessionComponent usersession.Component,
	providers map[string]Provider) *component {
	return &component{
		providers:            providers,
		repo:                 repo,
		productComponent:     productComponent,
		userSessionComponent: userSessionComponent,
	}
}

func (c *component) Create(ctx context.Context, kyc *models.KycSubmission) (*models.KycSubmission, error) {
	kyc.Status = QueueStatus
	nationality := strings.ToLower(kyc.UserInfo.Nationality)
	kyc.UserInfo.Nationality = nationality

	if nationality == Nigeria {
		hasSubmissionInQueue := c.repo.HasUserSubmissionInQueue(ctx, kyc.UserInfo.IDNumber)
		if hasSubmissionInQueue {
			return nil, errors.New("there's a submission in the queue. Please complete it before opening a new one")
		}
	}

	if err := c.createVerificationLinks(kyc); err != nil {
		return nil, err
	}

	newKycSubmission, err := c.repo.Create(ctx, kyc)
	if err != nil {
		return nil, err
	}

	return newKycSubmission, nil
}

func (c *component) GetByProductID(ctx context.Context, productID string) ([]models.KycSubmission, error) {
	authCtx := c.userSessionComponent.GetAuthContextFromCtx(ctx)
	return c.repo.GetByProductID(ctx, productID, authCtx.OrgID)
}

func (c *component) GetByOrgID(ctx context.Context) ([]models.KycSubmission, error) {
	authCtx := c.userSessionComponent.GetAuthContextFromCtx(ctx)
	return c.repo.GetByOrgID(ctx, authCtx.OrgID)
}

func (c *component) GetByID(ctx context.Context, id string, productID string) (*models.KycSubmission, error) {
	authCtx := c.userSessionComponent.GetAuthContextFromCtx(ctx)
	return c.repo.GetByID(ctx, id, productID, authCtx.OrgID)
}

func (c *component) UpdateStatusByID(ctx context.Context, id string, productID string, status string) error {
	product, err := c.productComponent.GetByID(ctx, productID)
	if err != nil {
		return err
	}

	authCtx := c.userSessionComponent.GetAuthContextFromCtx(ctx)
	kycSubmission, err := c.repo.GetByID(ctx, id, productID, authCtx.OrgID)
	if err != nil {
		return err
	}

	if err = c.repo.UpdateStatusByID(ctx, id, productID, authCtx.OrgID, status); err != nil {
		return err
	}

	var body string
	switch strings.ToLower(status) {
	case "accepted":
		body = fmt.Sprintf("Dear %s %s,\n\n"+
			"We are glad to inform you that your visa application %s with %s has been accepted.",
			kycSubmission.UserInfo.FirstName, kycSubmission.UserInfo.LastName, kycSubmission.PackageID, kycSubmission.PackageID)

	case "rejected":
		body = fmt.Sprintf("Dear %s %s,\n\n"+
			"We regret to inform you that your visa application %s with %s has been rejected.",
			kycSubmission.UserInfo.FirstName, kycSubmission.UserInfo.LastName, kycSubmission.ID, kycSubmission.PackageID)
	}

	if err := notifier.SendEmailNotification(
		[]string{kycSubmission.UserInfo.Email},
		fmt.Sprintf("KYC Check - %s", product.Name),
		body,
		true,
	); err != nil {
		slog.Error("error sending notification email", "email", kycSubmission.UserInfo.Email, "status", status)
		return err
	}

	return nil
}

func (c *component) UpdateByProviderInfo(ctx context.Context, providerCallback *models.ProviderCallback) error {
	return c.repo.UpdateByProviderInfo(ctx, providerCallback)
}

func (c *component) createVerificationLinks(kyc *models.KycSubmission) error {
	var (
		employmentVerificationLink string
		incomeVerificationLink     string
	)

	id, err := uuid.NewRandom()
	if err != nil {
		return err
	}
	kyc.ID = id.String()

	switch kyc.UserInfo.Nationality {
	case Nigeria:
		// create an income verification link with CredjtChek as provider
		incomeProvider, ok := c.providers[CreditCheckProvider]
		if !ok {
			return fmt.Errorf("no income verfication provider available for %s", kyc.UserInfo.Nationality)
		}
		id, err := uuid.NewRandom()
		if err != nil {
			return err
		}
		kyc.ID = id.String()

		verificationLink, err := incomeProvider.CreateLink(kyc.ID, kyc.UserInfo.FirstName, kyc.UserInfo.LastName)
		if err != nil {
			return err
		}
		incomeVerificationLink = verificationLink

		// create an employment verification link with Okra as provider
		employmentProvider, ok := c.providers[OkraProvider]
		if !ok {
			return fmt.Errorf("no employment verification provider available for %s", kyc.UserInfo.Nationality)
		}

		verificationLink, err = employmentProvider.CreateLink(kyc.ID, kyc.UserInfo.FirstName, kyc.UserInfo.LastName)
		if err != nil {
			return err
		}
		employmentVerificationLink = verificationLink

	case Indonesia:
		provider, ok := c.providers[OneBrickProvider]
		if !ok {
			return fmt.Errorf("no verification provider available for %s, please contact support", kyc.UserInfo.Nationality)
		}

		verificationLink, err := provider.CreateLink(kyc.ID, kyc.UserInfo.FirstName, kyc.UserInfo.LastName)
		if err != nil {
			return err
		}
		incomeVerificationLink = verificationLink
		employmentVerificationLink = verificationLink
	}

	passportProvider, ok := c.providers[IdenfyProvider]
	if !ok {
		return errors.New("no passport provider available, please contact support")
	}

	passportVerificationLink, err := passportProvider.CreateLink(kyc.ID, kyc.UserInfo.FirstName, kyc.UserInfo.LastName)
	if err != nil {
		return err
	}

	kyc.Checklist = map[string]interface{}{
		"employment_verification_link": employmentVerificationLink,
		"income_verification_link":     incomeVerificationLink,
		"passport_verification_link":   passportVerificationLink,
	}

	return nil
}
