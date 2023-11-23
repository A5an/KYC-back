package server

import (
	"log/slog"

	"github.com/clerkinc/clerk-sdk-go/clerk"
	"github.com/spf13/viper"
	"gorm.io/gorm"

	"github.com/Sinbad-HQ/kyc/config"
	"github.com/Sinbad-HQ/kyc/core/components/kyc"
	"github.com/Sinbad-HQ/kyc/core/components/kyc/providers"
	"github.com/Sinbad-HQ/kyc/core/components/packages"
	"github.com/Sinbad-HQ/kyc/core/components/risk_parameters"
	"github.com/Sinbad-HQ/kyc/core/components/usersession"
	"github.com/Sinbad-HQ/kyc/db"
)

type App struct {
	DB                *gorm.DB
	shutdownCallbacks []func()

	// logger
	logger *slog.Logger

	// db repositories
	ProductRepo       packages.Repo
	KycRepo           kyc.Repo
	RiskParameterRepo risk_parameters.Repo

	// components handlers
	UserSessionComponent   usersession.Component
	PackageComponent       packages.Component
	KycComponent           kyc.Component
	RiskParameterComponent risk_parameters.Component

	// kyc providers for handling callbacks
	CreditCheck kyc.Provider
	OneBrick    kyc.Provider
	Idenfy      kyc.Provider
	Okra        kyc.Provider
}

func NewApp() (app *App, err error) {
	app = &App{}
	app.logger = slog.Default()
	if app.DB, err = db.Connect(app.logger, config.GetDatabaseConfig()); err != nil {
		return
	}

	app.shutdownCallbacks = []func(){}

	// repositories initialization
	if app.ProductRepo, err = packages.NewRepo(app.DB); err != nil {
		return app, err
	}
	if app.KycRepo, err = kyc.NewRepo(app.DB); err != nil {
		return app, err
	}
	if app.RiskParameterRepo, err = risk_parameters.NewRepo(app.DB); err != nil {
		return app, err
	}

	client, err := clerk.NewClient(viper.GetString("clerk.api-token"))
	if err != nil {
		return nil, err
	}

	idenfyConfig := config.GetIdenfyConfig()
	app.Idenfy = providers.NewIdenfyClient(idenfyConfig.BaseURL, idenfyConfig.ApiKey, idenfyConfig.ApiSecret)

	oneBrickConfig := config.GetOneBrickConfig()
	app.OneBrick = providers.NewOneBrickClient(
		oneBrickConfig.BaseURL,
		oneBrickConfig.ClientID,
		oneBrickConfig.ClientSecret,
	)

	creditCheckConfig := config.GetCreditCheckConfig()
	app.CreditCheck = providers.NewCreditChekClient(creditCheckConfig.BaseURL, creditCheckConfig.PublicKey)

	app.Okra = providers.NewOkraClient()

	kycProviders := map[string]kyc.Provider{
		kyc.OneBrickProvider:    app.OneBrick,
		kyc.CreditCheckProvider: app.CreditCheck,
		kyc.IdenfyProvider:      app.Idenfy,
		kyc.OkraProvider:        app.Okra,
	}

	// components initialization
	app.UserSessionComponent = usersession.NewComponent(client)
	app.RiskParameterComponent = risk_parameters.NewComponent(app.RiskParameterRepo, app.UserSessionComponent)
	app.PackageComponent = packages.NewComponent(app.ProductRepo, app.UserSessionComponent, app.RiskParameterComponent)
	app.KycComponent = kyc.NewComponent(app.KycRepo, app.PackageComponent, app.UserSessionComponent, kycProviders)

	return app, nil
}

func (app *App) Shutdown() {
	for _, callback := range app.shutdownCallbacks {
		callback()
	}
}
