package core

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	validation "github.com/go-ozzo/ozzo-validation"
	"github.com/gorilla/mux"

	"github.com/Sinbad-HQ/kyc/core/components/product"
	"github.com/Sinbad-HQ/kyc/core/components/product/models"
)

type CreateRiskParameterRequest struct {
	Country          string  `json:"country"`
	AccountBalance   float64 `json:"account_balance"`
	AverageSalary    float64 `json:"average_salary"`
	EmploymentStatus bool    `json:"employment_status"`
}

func (c CreateRiskParameterRequest) Validate() error {
	return validation.ValidateStruct(&c,
		validation.Field(&c.Country, validation.Required),
		validation.Field(&c.AccountBalance, validation.Required),
		validation.Field(&c.AverageSalary, validation.Required),
		validation.Field(&c.Country, validation.By(func(value interface{}) error {
			var isSupported bool
			country := strings.ToLower(value.(string))
			for _, supportedCountry := range product.SupportedCountries {
				if country == supportedCountry {
					isSupported = true
				}
			}

			if !isSupported {
				return fmt.Errorf("%s is not supported", country)
			}
			return nil
		})),
	)
}

func (app *App) CreateRiskParameter(w http.ResponseWriter, r *http.Request) {
	reqBody := CreateRiskParameterRequest{}
	if err := json.NewDecoder(r.Body).Decode(&reqBody); err != nil {
		app.HandleAPIError(
			fmt.Errorf("failed to decode body: %w", err), http.StatusInternalServerError, w,
		)
		return
	}

	if err := reqBody.Validate(); err != nil {
		app.HandleAPIError(
			err, http.StatusBadRequest, w,
		)
		return
	}

	createdRiskParameter, err := app.ProductComponent.CreateRiskParameter(r.Context(), &models.RiskParameter{
		Country:          reqBody.Country,
		AccountBalance:   reqBody.AccountBalance,
		AverageSalary:    reqBody.AverageSalary,
		EmploymentStatus: reqBody.EmploymentStatus,
	})
	if err != nil {
		app.HandleAPIError(
			fmt.Errorf("failed to create risk parameter: %w", err), http.StatusInternalServerError, w,
		)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(createdRiskParameter); err != nil {
		app.HandleAPIError(
			fmt.Errorf("failed to encode response: %w", err), http.StatusInternalServerError, w,
		)
		return
	}
}

func (app *App) GetRiskParameters(w http.ResponseWriter, r *http.Request) {
	riskParameters, err := app.ProductComponent.GetRiskParameters(r.Context())
	if err != nil {
		app.HandleAPIError(
			fmt.Errorf("failed to get risk parameters: %w", err), http.StatusInternalServerError, w,
		)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(riskParameters); err != nil {
		app.HandleAPIError(
			fmt.Errorf("failed to encode response: %w", err), http.StatusInternalServerError, w,
		)
		return
	}
}

func (app *App) UpdateRiskParameter(w http.ResponseWriter, r *http.Request) {
	reqBody := CreateRiskParameterRequest{}
	if err := json.NewDecoder(r.Body).Decode(&reqBody); err != nil {
		app.HandleAPIError(
			fmt.Errorf("failed to decode body: %w", err), http.StatusInternalServerError, w,
		)
		return
	}

	if err := reqBody.Validate(); err != nil {
		app.HandleAPIError(
			err, http.StatusBadRequest, w,
		)
		return
	}

	createdRiskParameter, err := app.ProductComponent.UpdateRiskParameterByID(r.Context(), &models.RiskParameter{
		ID:               mux.Vars(r)["riskParameterID"],
		Country:          reqBody.Country,
		AccountBalance:   reqBody.AccountBalance,
		AverageSalary:    reqBody.AverageSalary,
		EmploymentStatus: reqBody.EmploymentStatus,
	})
	if err != nil {
		app.HandleAPIError(
			fmt.Errorf("failed to update risk parameter: %w", err), http.StatusInternalServerError, w,
		)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(createdRiskParameter); err != nil {
		app.HandleAPIError(
			fmt.Errorf("failed to encode response: %w", err), http.StatusInternalServerError, w,
		)
		return
	}
}
