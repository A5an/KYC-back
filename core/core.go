package core

import (
	"log/slog"

	"github.com/jmoiron/sqlx"
	"github.com/nedpals/supabase-go"
	"github.com/spf13/viper"

	"github.com/Sinbad-HQ/kyc/config"
	"github.com/Sinbad-HQ/kyc/core/components/kyc"
	"github.com/Sinbad-HQ/kyc/core/components/kyc/providers"
	"github.com/Sinbad-HQ/kyc/core/components/product"
	"github.com/Sinbad-HQ/kyc/core/components/usersession"
	"github.com/Sinbad-HQ/kyc/db"
)

type App struct {
	Info              Info
	DB                *sqlx.DB
	shutdownCallbacks []func()

	// logger
	logger *slog.Logger

	// db repositories
	ProductRepo product.Repo
	KycRepo     kyc.Repo

	// components handlers
	UserSessionComponent usersession.Component
	ProductComponent     product.Component
	KycComponent         kyc.Component

	// kyc providers for handling callbacks
	CreditChek kyc.Provider
	OneBrick   kyc.Provider
}

func NewApp() (app *App, err error) {
	app = &App{
		Info: Info{
			Environment: viper.GetString("env"),
		},
	}

	app.logger = slog.Default()
	if app.DB, err = db.Connect(app.logger, config.GetDatabaseConfig()); err != nil {
		return
	}

	app.shutdownCallbacks = []func(){
		func() { _ = app.DB.Close },
	}

	// repositories initialization
	if app.ProductRepo, err = product.NewRepo(app.DB); err != nil {
		return app, err
	}
	if app.KycRepo, err = kyc.NewRepo(app.DB); err != nil {
		return app, err
	}

	sbConfig := config.GetSupabaseConfig()
	client := supabase.CreateClient(sbConfig.BaseURL, sbConfig.ApiKey)

	oneBrickConfig := config.GetOneBrickConfig()
	app.OneBrick = providers.NewOneBrickClient(
		oneBrickConfig.BaseURL,
		oneBrickConfig.ClientID,
		oneBrickConfig.ClientSecret,
	)

	creditCheckConfig := config.GetCreditCheckConfig()
	app.CreditChek = providers.NewCreditChekClient(creditCheckConfig.BaseURL, creditCheckConfig.PublicKey)

	kycProviders := map[string]kyc.Provider{
		"indonesia": app.OneBrick,
		"nigeria":   app.CreditChek,
	}

	// components initialization
	app.UserSessionComponent = usersession.NewComponent(client)
	app.ProductComponent = product.NewComponent(app.ProductRepo, app.UserSessionComponent)
	app.KycComponent = kyc.NewComponent(app.KycRepo, app.ProductComponent, app.UserSessionComponent, kycProviders)

	return app, nil
}

func (app *App) Shutdown() {
	for _, callback := range app.shutdownCallbacks {
		callback()
	}
}
