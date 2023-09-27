package product

import (
	"context"
	"database/sql"
	"errors"

	"github.com/jmoiron/sqlx"

	"github.com/Sinbad-HQ/kyc/core/components/product/models"
)

type Repo interface {
	Create(ctx context.Context, product *models.Product) (*models.Product, error)
	GetByProviderID(ctx context.Context, providerID string) ([]models.Product, error)
	GetByID(ctx context.Context, id string, providerID string) (*models.Product, error)

	CreateRiskParameter(ctx context.Context, riskParameter *models.RiskParameter) (*models.RiskParameter, error)
	GetRiskParameters(ctx context.Context, providerID string) ([]models.RiskParameter, error)
	GetRiskParameterByCountry(ctx context.Context, country string) (*models.RiskParameter, error)
}

const (
	createProduct = `
	INSERT INTO products (id, name, description, image_url, provider_id) VALUES (:id, :name, :description, :image_url, :provider_id)
	                                                                    RETURNING id, name, description, image_url, provider_id
	`
	selectByProductID       = `SELECT id, name, description, image_url, provider_id FROM products WHERE provider_id = $1`
	selectByIDAndProviderID = `SELECT id, name, description, image_url, provider_id FROM products WHERE id = $1 AND provider_id = $2`

	createRiskParameter = `INSERT INTO risk_parameters(id, country, account_balance, average_salary, employment_status, provider_id) VALUES 
(:id, :country, :account_balance, :average_salary, :employment_status, :provider_id) RETURNING id, country, account_balance, average_salary, employment_status, provider_id`
	selectRiskParameters = `SELECT * FROM risk_parameters WHERE provider_id = $1`

	selectRiskParameterByName = "SELECT * FROM risk_parameters WHERE country = $1"
)

type repo struct {
	createProduct           *sqlx.NamedStmt
	selectByProductID       *sqlx.Stmt
	selectByIDAndProviderID *sqlx.Stmt

	// product risk parameters
	createRiskParameter       *sqlx.NamedStmt
	selectRiskParameters      *sqlx.Stmt
	selectRiskParameterByName *sqlx.Stmt
}

func NewRepo(db *sqlx.DB) (r *repo, err error) {
	r = &repo{}

	if r.createProduct, err = db.PrepareNamed(createProduct); err != nil {
		return
	}
	if r.selectByProductID, err = db.Preparex(selectByProductID); err != nil {
		return
	}
	if r.selectByIDAndProviderID, err = db.Preparex(selectByIDAndProviderID); err != nil {
		return
	}
	if r.createRiskParameter, err = db.PrepareNamed(createRiskParameter); err != nil {
		return
	}
	if r.selectRiskParameters, err = db.Preparex(selectRiskParameters); err != nil {
		return
	}
	if r.selectRiskParameterByName, err = db.Preparex(selectRiskParameterByName); err != nil {
		return
	}

	return
}

func (r *repo) Create(ctx context.Context, product *models.Product) (*models.Product, error) {
	var createdProduct models.Product
	if err := r.createProduct.GetContext(ctx, &createdProduct, product); err != nil {
		return nil, err
	}

	return &createdProduct, nil
}

func (r *repo) GetByProviderID(_ context.Context, providerID string) ([]models.Product, error) {
	products := make([]models.Product, 0)
	if err := r.selectByProductID.Select(&products, providerID); !errors.Is(err, sql.ErrNoRows) {
		return products, err
	}
	return products, nil
}

func (r *repo) GetByID(_ context.Context, id string, providerID string) (*models.Product, error) {
	var product models.Product
	err := r.selectByIDAndProviderID.Get(&product, id, providerID)
	if errors.Is(err, sql.ErrNoRows) {
		err = errors.New("product does not exist")
	}
	return &product, err
}

func (r *repo) CreateRiskParameter(ctx context.Context, riskParameter *models.RiskParameter) (*models.RiskParameter, error) {
	var createdRiskParameter models.RiskParameter
	if err := r.createRiskParameter.GetContext(ctx, &createdRiskParameter, riskParameter); err != nil {
		return nil, err
	}

	return &createdRiskParameter, nil
}

func (r *repo) GetRiskParameters(_ context.Context, providerID string) ([]models.RiskParameter, error) {
	riskProviders := make([]models.RiskParameter, 0)
	if err := r.selectRiskParameters.Select(&riskProviders, providerID); !errors.Is(err, sql.ErrNoRows) {
		return riskProviders, err
	}
	return riskProviders, nil
}

func (r *repo) GetRiskParameterByCountry(_ context.Context, country string) (*models.RiskParameter, error) {
	var riskParameter models.RiskParameter
	err := r.selectRiskParameterByName.Get(&riskParameter, country)
	if errors.Is(err, sql.ErrNoRows) {
		err = errors.New("risk parameter does not exist")
	}
	return &riskParameter, err
}
