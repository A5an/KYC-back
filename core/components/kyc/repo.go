package kyc

import (
	"context"
	"database/sql"
	"errors"

	"github.com/jmoiron/sqlx"

	"github.com/Sinbad-HQ/kyc/core/components/kyc/models"
)

type Repo interface {
	Create(ctx context.Context, kyc *models.Kyc) (*models.Kyc, error)
	GetByProductID(ctx context.Context, providerID string, productID string) ([]models.Kyc, error)
	GetByIDAndProviderID(ctx context.Context, id string, providerID string) (*models.Kyc, error)
	GetByID(ctx context.Context, id string) (*models.Kyc, error)
	UpdateByID(ctx context.Context, kyc *models.Kyc) error
}

const (
	createKyc = `INSERT INTO products_kyc (id, product_id, provider_id, first_name, middle_name, last_name, dob, gender, country,link, status) 
VALUES(:id, :product_id, :provider_id, :first_name, :middle_name, :last_name, :dob, :gender, :country, :link, :status)RETURNING id, product_id, provider_id, first_name, middle_name, last_name, dob, gender, country, link, status`

	selectByProductIDAndProviderID = `SELECT * FROM products_kyc WHERE provider_id = $1 AND product_id = $2`
	selectByIDAndProviderID        = `SELECT * FROM products_kyc WHERE id = $1 AND provider_id = $2`
	selectByID                     = `SELECT * FROM products_kyc WHERE id = $1`
	updateByID                     = `UPDATE products_kyc SET
		status = :status,
		account_balance = :account_balance,
		average_salary = :average_salary,
		employment_status = :employment_status,
        identity_response = :identity_response,
        account_balance_risk_level = :account_balance_risk_level,
        average_salary_risk_level = :average_salary_risk_level,
        employment_risk_level = :employment_risk_level
        WHERE id = :id AND provider_id = :provider_id`

	PendingStatus  = "PENDING"
	AprovedStatus  = "APPROVED"
	RejectedStatus = "REJECTED"
)

type repo struct {
	createKyc                      *sqlx.NamedStmt
	selectByProductIDAndProviderID *sqlx.Stmt
	selectByIDAndProviderID        *sqlx.Stmt
	selectByID                     *sqlx.Stmt
	updateByID                     *sqlx.NamedStmt
	// update kyc fields for after receving callback
}

func NewRepo(db *sqlx.DB) (r *repo, err error) {
	r = &repo{}

	if r.createKyc, err = db.PrepareNamed(createKyc); err != nil {
		return
	}
	if r.selectByProductIDAndProviderID, err = db.Preparex(selectByProductIDAndProviderID); err != nil {
		return
	}
	if r.selectByIDAndProviderID, err = db.Preparex(selectByIDAndProviderID); err != nil {
		return
	}
	if r.selectByID, err = db.Preparex(selectByID); err != nil {
		return
	}
	if r.updateByID, err = db.PrepareNamed(updateByID); err != nil {
		return
	}

	return
}

func (r *repo) Create(ctx context.Context, kyc *models.Kyc) (*models.Kyc, error) {
	var createdKyc models.Kyc
	if err := r.createKyc.GetContext(ctx, &createdKyc, kyc); err != nil {
		return nil, err
	}

	return &createdKyc, nil
}

func (r *repo) GetByProductID(_ context.Context, providerID string, productID string) ([]models.Kyc, error) {
	kyc := make([]models.Kyc, 0)
	if err := r.selectByProductIDAndProviderID.Select(&kyc, providerID, productID); !errors.Is(err, sql.ErrNoRows) {
		return kyc, err
	}
	return kyc, nil
}

func (r *repo) GetByID(_ context.Context, id string) (*models.Kyc, error) {
	var kyc models.Kyc
	err := r.selectByID.Get(&kyc, id)
	if errors.Is(err, sql.ErrNoRows) {
		err = errors.New("kyc does not exist")
	}
	return &kyc, err
}

func (r *repo) GetByIDAndProviderID(_ context.Context, id string, providerID string) (*models.Kyc, error) {
	var kyc models.Kyc
	err := r.selectByIDAndProviderID.Get(&kyc, id, providerID)
	if errors.Is(err, sql.ErrNoRows) {
		err = errors.New("kyc does not exist")
	}
	return &kyc, err
}

func (r *repo) UpdateByID(_ context.Context, kyc *models.Kyc) error {
	_, err := r.updateByID.Exec(kyc)
	if err != nil {
		return err
	}
	return nil
}
