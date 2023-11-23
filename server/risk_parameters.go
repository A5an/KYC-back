package server

import (
	"encoding/json"
	"fmt"
	"net/http"

	validation "github.com/go-ozzo/ozzo-validation"
	"github.com/gorilla/mux"

	"github.com/Sinbad-HQ/kyc/core/components/risk_parameters/models"
)

type riskParameterRequest struct {
	Name             string  `json:"name"`
	AccountBalance   float64 `json:"account_balance"`
	AverageSalary    float64 `json:"average_salary"`
	EmploymentStatus bool    `json:"employment_status"`
}

func (r riskParameterRequest) Validate() error {
	return validation.ValidateStruct(&r,
		validation.Field(&r.Name, validation.Required),
		validation.Field(&r.AverageSalary, validation.Required),
		validation.Field(&r.AccountBalance, validation.Required),
	)
}

func (app *App) CreateRiskParameter(w http.ResponseWriter, r *http.Request) {
	reqBody := riskParameterRequest{}
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

	createdRiskParameter, err := app.RiskParameterComponent.Create(r.Context(), &models.RiskParameter{
		Name:             reqBody.Name,
		AccountBalance:   reqBody.AccountBalance,
		AverageSalary:    reqBody.AverageSalary,
		EmploymentStatus: &reqBody.EmploymentStatus,
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

func (app *App) GetRiskParameterByID(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["id"]
	riskParameters, err := app.RiskParameterComponent.GetByID(r.Context(), id)
	if err != nil {
		app.HandleAPIError(
			fmt.Errorf("failed to get risk parameter: %w", err), http.StatusInternalServerError, w,
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

func (app *App) GetRiskParameterByProviderID(w http.ResponseWriter, r *http.Request) {
	riskParameters, err := app.RiskParameterComponent.GetByProviderID(r.Context())
	if err != nil {
		app.HandleAPIError(
			fmt.Errorf("failed to get provider risk parameters: %w", err), http.StatusInternalServerError, w,
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

func (app *App) UpdateRiskParameterByID(w http.ResponseWriter, r *http.Request) {
	reqBody := riskParameterRequest{}
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

	id := mux.Vars(r)["id"]
	updatedProduct, err := app.RiskParameterComponent.UpdateByID(r.Context(), id, &models.RiskParameter{
		Name:             reqBody.Name,
		AccountBalance:   reqBody.AccountBalance,
		AverageSalary:    reqBody.AverageSalary,
		EmploymentStatus: &reqBody.EmploymentStatus,
	})
	if err != nil {
		app.HandleAPIError(
			fmt.Errorf("failed to update risk parameter: %w", err), http.StatusInternalServerError, w,
		)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(updatedProduct); err != nil {
		app.HandleAPIError(
			fmt.Errorf("failed to encode response: %w", err), http.StatusInternalServerError, w,
		)
		return
	}
}

func (app *App) DeleteRiskParameterByID(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["id"]
	err := app.RiskParameterComponent.DeleteByID(r.Context(), id)
	if err != nil {
		app.HandleAPIError(
			fmt.Errorf("failed to delete risk parameter: %w", err), http.StatusInternalServerError, w,
		)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
}
