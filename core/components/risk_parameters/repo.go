package risk_parameters

import (
	"context"
	"fmt"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/Sinbad-HQ/kyc/core/components/risk_parameters/models"
	"github.com/Sinbad-HQ/kyc/db/model"
)

type Repo interface {
	Create(ctx context.Context, riskParameter *models.RiskParameter) (*models.RiskParameter, error)
	GetByID(ctx context.Context, id string, orgID string) (*models.RiskParameter, error)
	GetByOrgID(ctx context.Context, orgID string) ([]models.RiskParameter, error)
	UpdateByID(ctx context.Context, id string, updatedRiskParameter *models.RiskParameter) (*models.RiskParameter, error)
	DeleteByID(ctx context.Context, id string, orgID string) error
}

type repo struct {
	db *gorm.DB
}

func NewRepo(db *gorm.DB) (r *repo, err error) {
	r = &repo{db: db}
	return
}

func (r *repo) Create(ctx context.Context, riskParameter *models.RiskParameter) (*models.RiskParameter, error) {
	result := r.db.WithContext(ctx).Clauses(clause.OnConflict{UpdateAll: true}).Create(riskParameter)
	return riskParameter, result.Error
}

func (r *repo) GetByOrgID(_ context.Context, orgID string) ([]models.RiskParameter, error) {
	riskParameters := make([]models.RiskParameter, 0)

	tx := r.db.Where(&models.RiskParameter{OrgID: orgID}).Find(&riskParameters)
	if err := tx.Error; err != nil {
		return nil, err
	}

	return riskParameters, nil
}

func (r *repo) GetByID(_ context.Context, id string, orgID string) (*models.RiskParameter, error) {
	var riskParameter models.RiskParameter
	tx := r.db.Where(&models.RiskParameter{
		Model: model.Model{ID: id},
		OrgID: orgID,
	}).First(&riskParameter)
	if tx.Error != nil {
		return nil, fmt.Errorf("risk parameter with id %s not found", id)
	}

	return &riskParameter, nil
}

func (r *repo) UpdateByID(ctx context.Context, id string, updatedRiskParameter *models.RiskParameter) (*models.RiskParameter, error) {
	updatedRiskParameter.ID = id
	tx := r.db.Where(&models.RiskParameter{
		Model: model.Model{ID: id},
		OrgID: updatedRiskParameter.OrgID,
	}).Updates(updatedRiskParameter)
	if tx.Error != nil {
		return nil, fmt.Errorf("risk parameter with id %s not found", id)
	}

	return r.GetByID(ctx, id, updatedRiskParameter.OrgID)
}

func (r *repo) DeleteByID(_ context.Context, id string, orgID string) error {
	obj := models.RiskParameter{
		Model: model.Model{ID: id},
		OrgID: orgID,
	}

	tx := r.db.Unscoped().Delete(&obj)
	if err := tx.Error; err != nil {
		return err
	}

	if tx.RowsAffected < 1 {
		return fmt.Errorf("risk parameter with id %s not found", id)
	}

	return nil
}
