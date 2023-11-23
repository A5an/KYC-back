package packages

import (
	"context"
	"fmt"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/Sinbad-HQ/kyc/core/components/packages/models"
	"github.com/Sinbad-HQ/kyc/db/model"
)

type Repo interface {
	Create(ctx context.Context, kycPackage *models.Package) (*models.Package, error)
	GetByID(ctx context.Context, id string, orgID string) (*models.Package, error)
	GetByOrgID(ctx context.Context, orgID string) ([]models.Package, error)
	UpdateByID(ctx context.Context, id string, updatedPackage *models.Package) (*models.Package, error)
	DeleteByID(ctx context.Context, id string, orgID string) error
}

type repo struct {
	db *gorm.DB
}

func NewRepo(db *gorm.DB) (r *repo, err error) {
	r = &repo{db: db}
	return
}

func (r *repo) Create(ctx context.Context, kycPackage *models.Package) (*models.Package, error) {
	result := r.db.WithContext(ctx).Clauses(clause.OnConflict{UpdateAll: true}).Create(kycPackage)
	return kycPackage, result.Error
}

func (r *repo) GetByOrgID(_ context.Context, orgID string) ([]models.Package, error) {
	products := make([]models.Package, 0)

	tx := r.db.Where(&models.Package{OrgID: orgID}).Find(&products)
	if err := tx.Error; err != nil {
		return nil, err
	}

	return products, nil
}

func (r *repo) GetByID(_ context.Context, id string, orgID string) (*models.Package, error) {
	var kycPackage models.Package
	tx := r.db.Where(&models.Package{
		Model: model.Model{ID: id},
		OrgID: orgID,
	}).First(&kycPackage)
	if tx.Error != nil {
		return nil, fmt.Errorf("package with id %s not found", id)
	}

	return &kycPackage, nil
}

func (r *repo) UpdateByID(ctx context.Context, id string, updatedPackage *models.Package) (*models.Package, error) {
	updatedPackage.ID = id
	tx := r.db.Where(&models.Package{
		Model: model.Model{ID: id},
		OrgID: updatedPackage.OrgID,
	}).Updates(updatedPackage)
	if tx.Error != nil {
		return nil, fmt.Errorf("package with id %s not found", id)
	}

	return r.GetByID(ctx, id, updatedPackage.OrgID)
}

func (r *repo) DeleteByID(_ context.Context, id string, orgID string) error {
	obj := models.Package{
		Model: model.Model{ID: id},
		OrgID: orgID,
	}

	tx := r.db.Where(&obj).Delete(&obj)
	if err := tx.Error; err != nil {
		return err
	}

	if tx.RowsAffected < 1 {
		return fmt.Errorf("package with id %s not found", id)
	}

	return nil
}
