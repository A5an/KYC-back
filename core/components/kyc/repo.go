package kyc

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/Sinbad-HQ/kyc/core/components/kyc/models"
	pmodels "github.com/Sinbad-HQ/kyc/core/components/packages/models"
	rmodels "github.com/Sinbad-HQ/kyc/core/components/risk_parameters/models"
	"github.com/Sinbad-HQ/kyc/db/model"
)

type Repo interface {
	Create(ctx context.Context, kyc *models.KycSubmission) (*models.KycSubmission, error)
	GetByProductID(ctx context.Context, productID string, orgID string) ([]models.KycSubmission, error)
	GetByOrgID(ctx context.Context, orgID string) ([]models.KycSubmission, error)
	GetByID(ctx context.Context, id string, productID string, orgID string) (*models.KycSubmission, error)
	UpdateByID(ctx context.Context, updatedKycSubmission *models.KycSubmission) (*models.KycSubmission, error)
	UpdateStatusByID(ctx context.Context, id string, productID string, orgID string, status string) error
	HasUserSubmissionInQueue(ctx context.Context, idNumber string, email string) bool
	UpdateByProviderInfo(ctx context.Context, kyc *models.ProviderCallback) error
}

type repo struct {
	db *gorm.DB
}

func NewRepo(db *gorm.DB) (r *repo, err error) {
	r = &repo{db: db}
	return
}

// Create creates new kyc submission and it's associations.
func (r *repo) Create(_ context.Context, kyc *models.KycSubmission) (*models.KycSubmission, error) {
	var product pmodels.Package
	tx := r.db.Where(&pmodels.Package{Model: model.Model{ID: kyc.PackageID}}).First(&product)
	if tx.Error != nil {
		return nil, fmt.Errorf("packages with id %s not found", kyc.PackageID)
	}

	kyc.OrgID = product.OrgID
	result := r.db.Session(&gorm.Session{FullSaveAssociations: true}).Create(kyc)
	return kyc, result.Error
}

func (r *repo) GetByProductID(_ context.Context, productID string, orgID string) ([]models.KycSubmission, error) {
	var product pmodels.Package
	tx := r.db.Where(&pmodels.Package{
		Model: model.Model{ID: productID},
		OrgID: orgID,
	}).Preload("KycSubmissions").
		Preload("KycSubmissions.UserInfo").
		Preload("KycSubmissions.PassportInfo").
		Preload("KycSubmissions.EmploymentInfo").
		Preload("KycSubmissions.BankInfo").
		Preload("KycSubmissions.AddressInfo").
		First(&product)
	if err := tx.Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return []models.KycSubmission{}, nil
		}

		return nil, err
	}

	return product.KycSubmissions, nil
}

func (r *repo) GetByOrgID(_ context.Context, orgID string) ([]models.KycSubmission, error) {
	kycSubmissions := make([]models.KycSubmission, 0)
	tx := r.db.Where(&models.KycSubmission{
		OrgID: orgID,
	}).
		Preload("UserInfo").
		Preload("PassportInfo").
		Preload("EmploymentInfo").
		Preload("BankInfo").
		Preload("AddressInfo").Find(&kycSubmissions)
	if err := tx.Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return []models.KycSubmission{}, nil
		}
		return nil, err
	}

	return kycSubmissions, nil
}

func (r *repo) GetByID(_ context.Context, id string, productID string, orgID string) (*models.KycSubmission, error) {
	var kyc models.KycSubmission
	tx := r.db.Where(&models.KycSubmission{
		Model:     model.Model{ID: id},
		PackageID: productID,
		OrgID:     orgID,
	},
	).Preload("UserInfo").
		Preload("PassportInfo").
		Preload("EmploymentInfo").
		Preload("BankInfo").
		Preload("AddressInfo").First(&kyc)
	if tx.Error != nil {
		return nil, fmt.Errorf("kyc submission with id %s not found", id)
	}

	return &kyc, nil
}

// UpdateByID updated the kyc submission only without updating the underlying associations.
func (r *repo) UpdateByID(ctx context.Context, updatedKycSubmission *models.KycSubmission) (*models.KycSubmission, error) {
	id := updatedKycSubmission.ID
	orgID := updatedKycSubmission.OrgID
	productID := updatedKycSubmission.PackageID

	tx := r.db.Omit(clause.Associations).Where(&models.KycSubmission{
		Model:     model.Model{ID: id},
		PackageID: productID,
		OrgID:     orgID,
	}).Updates(updatedKycSubmission)
	if tx.Error != nil {
		return nil, fmt.Errorf("kyc submission with id %s not found", id)
	}

	return r.GetByID(ctx, id, productID, orgID)
}

func (r *repo) UpdateStatusByID(_ context.Context, id string, productID string, orgID string, status string) error {
	tx := r.db.Model(&models.KycSubmission{}).
		Omit(clause.Associations).
		Where(&models.KycSubmission{
			Model:     model.Model{ID: id},
			PackageID: productID,
			OrgID:     orgID,
		}).Update("status", status)
	if tx.Error != nil {
		return fmt.Errorf("kyc submission with id %s not found", id)
	}

	return nil
}

// HasUserSubmissionInQueue checks if a user with the given ID number has a submission in the queue.
// Used for Nigeria since the current provider does not support adding kyc submission id for identification.
func (r *repo) HasUserSubmissionInQueue(_ context.Context, idNumber string, email string) bool {
	kycSubmission := make([]models.KycSubmission, 0)
	if idNumber != "" {
		tx := r.db.Joins("JOIN user_infos ON user_infos.kyc_submission_id = kyc_submissions.id AND user_infos.id_number = ?", idNumber).
			Where(&models.KycSubmission{Status: QueueStatus}).First(&kycSubmission)
		if tx.Error != nil {
			if errors.Is(tx.Error, gorm.ErrRecordNotFound) {
				return false
			}

			slog.Error("error fetching user submission by id number", "idNumber", idNumber, "error", tx.Error)
			return true
		}
	}

	if email != "" {
		tx := r.db.Joins("JOIN user_infos ON user_infos.kyc_submission_id = kyc_submissions.id AND user_infos.email = ?", email).
			Where(&models.KycSubmission{Status: QueueStatus}).First(&kycSubmission)
		if tx.Error != nil {
			if errors.Is(tx.Error, gorm.ErrRecordNotFound) {
				return false
			}

			slog.Error("error fetching user submission by email", "email", email, "error", tx.Error)
			return true
		}
	}

	return len(kycSubmission) > 0
}

// UpdateByProviderInfo is responsible for updating the Know Your Customer (KYC) submission information.
// Handles updates that occur after a callback is received from the KYC provider.
// It is particularly meant for handling automatic verifications for specific countries.
func (r *repo) UpdateByProviderInfo(_ context.Context, providerCallback *models.ProviderCallback) error {
	var kycSubmission models.KycSubmission

	// fetch the kyc submission using kyc submission id
	if providerCallback.KycSubmissionID != "" {
		tx := r.db.Where(&models.KycSubmission{
			Model: model.Model{ID: providerCallback.KycSubmissionID},
		}).First(&kycSubmission)
		if tx.Error != nil {
			return tx.Error
		}
	} else {
		// for provider with no kyc submission id use user id number to fetch the kyc submission
		tx := r.db.Joins("JOIN user_infos ON user_infos.kyc_submission_id = kyc_submissions.id AND user_infos.id_number = ?", providerCallback.UserIDNumber).
			Where(&models.KycSubmission{Status: QueueStatus}).First(&kycSubmission)
		if tx.Error != nil {
			return tx.Error
		}
	}

	var kycPackage pmodels.Package
	tx := r.db.Where(&pmodels.Package{Model: model.Model{ID: kycSubmission.PackageID}}).First(&kycPackage)
	if tx.Error != nil {
		return fmt.Errorf("packages with id %s not found", kycSubmission.PackageID)
	}

	var riskParameter rmodels.RiskParameter
	tx = r.db.Where(&rmodels.RiskParameter{Model: model.Model{ID: kycPackage.RiskParameterID}}).First(&riskParameter)
	if tx.Error != nil {
		return fmt.Errorf("risk parameter with id %s not found", kycPackage.RiskParameterID)
	}

	if providerCallback.PassportInfo != nil {
		if err := r.db.Create(providerCallback.PassportInfo).Error; err != nil {
			return err
		}
	}

	if providerCallback.AddressInfo != nil {
		if err := r.db.Create(providerCallback.AddressInfo).Error; err != nil {
			return err
		}
	}

	if providerCallback.BankInfo != nil {
		providerCallback.BankInfo.KycSubmissionID = kycSubmission.ID
		providerCallback.BankInfo.AccountBalanceRiskLevel = HighRiskLevel
		if providerCallback.BankInfo.AccountBalance >= riskParameter.AccountBalance {
			providerCallback.BankInfo.AccountBalanceRiskLevel = LowRiskLevel
		}

		if err := r.db.Create(providerCallback.BankInfo).Error; err != nil {
			return err
		}
	}

	if providerCallback.EmploymentInfo != nil {
		providerCallback.EmploymentInfo.KycSubmissionID = kycSubmission.ID
		providerCallback.EmploymentInfo.AverageSalaryRiskLevel = HighRiskLevel
		if providerCallback.EmploymentInfo.AverageSalary >= riskParameter.AverageSalary {
			providerCallback.EmploymentInfo.AverageSalaryRiskLevel = LowRiskLevel
		}

		providerCallback.EmploymentInfo.EmploymentRiskLevel = LowRiskLevel
		if *riskParameter.EmploymentStatus {
			if !providerCallback.EmploymentInfo.EmploymentStatus {
				providerCallback.EmploymentInfo.EmploymentRiskLevel = HighRiskLevel
			}
		}

		if err := r.db.Create(providerCallback.EmploymentInfo).Error; err != nil {
			return err
		}
	}

	return nil
}
