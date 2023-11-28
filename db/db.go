package db

import (
	"log/slog"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"github.com/Sinbad-HQ/kyc/config"
	kycModels "github.com/Sinbad-HQ/kyc/core/components/kyc/models"
	"github.com/Sinbad-HQ/kyc/core/components/packages/models"
	riskModels "github.com/Sinbad-HQ/kyc/core/components/risk_parameters/models"
)

var registeredModels = []interface{}{
	&riskModels.RiskParameter{},
	&models.Package{},
	&kycModels.KycSubmission{},
	&kycModels.UserInfo{},
	&kycModels.PassportInfo{},
	&kycModels.EmploymentInfo{},
	&kycModels.BankInfo{},
	&kycModels.AddressInfo{},
}

// Connect establishes a connection to the database using the provided database configuration.
func Connect(logger *slog.Logger, cfg config.DatabaseConfig) (*gorm.DB, error) {
	db, err := gorm.Open(postgres.Open(cfg.URL()), &gorm.Config{
		NowFunc: func() time.Time {
			return time.Now().UTC()
		},
		PrepareStmt: true,
	})
	if err != nil {
		return nil, err
	}

	logger.Info("Connected to postgres on address: " + cfg.URL())

	if err := db.AutoMigrate(registeredModels...); err != nil {
		return nil, err
	}

	return db, nil
}
