package server

import (
	"encoding/json"
	"fmt"
	"net/http"

	validation "github.com/go-ozzo/ozzo-validation"
	"github.com/gorilla/mux"

	"github.com/Sinbad-HQ/kyc/core/components/packages/models"
)

type packageRequest struct {
	Name            string `json:"name"`
	Description     string `json:"description"`
	LogoURL         string `json:"logo_url"`
	RiskParameterID string `json:"risk_parameter_id"`
}

func (c packageRequest) Validate() error {
	return validation.ValidateStruct(&c,
		validation.Field(&c.Name, validation.Required),
	)
}

func (app *App) CreateProduct(w http.ResponseWriter, r *http.Request) {
	reqBody := packageRequest{}
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

	createdProduct, err := app.PackageComponent.Create(r.Context(), &models.Package{
		Name:            reqBody.Name,
		Description:     reqBody.Description,
		LogoURL:         reqBody.LogoURL,
		RiskParameterID: reqBody.RiskParameterID,
	})
	if err != nil {
		app.HandleAPIError(
			fmt.Errorf("failed to create packages: %w", err), http.StatusInternalServerError, w,
		)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(createdProduct); err != nil {
		app.HandleAPIError(
			fmt.Errorf("failed to encode response: %w", err), http.StatusInternalServerError, w,
		)
		return
	}
}

func (app *App) GetPackageByID(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["packageID"]
	product, err := app.PackageComponent.GetByID(r.Context(), id)
	if err != nil {
		app.HandleAPIError(
			fmt.Errorf("failed to get packages: %w", err), http.StatusInternalServerError, w,
		)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(product); err != nil {
		app.HandleAPIError(
			fmt.Errorf("failed to encode response: %w", err), http.StatusInternalServerError, w,
		)
		return
	}
}

func (app *App) GetPackageByProviderID(w http.ResponseWriter, r *http.Request) {
	products, err := app.PackageComponent.GetByProviderID(r.Context())
	if err != nil {
		app.HandleAPIError(
			fmt.Errorf("failed to get provider products: %w", err), http.StatusInternalServerError, w,
		)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(products); err != nil {
		app.HandleAPIError(
			fmt.Errorf("failed to encode response: %w", err), http.StatusInternalServerError, w,
		)
		return
	}
}

func (app *App) UpdatePackageByID(w http.ResponseWriter, r *http.Request) {
	reqBody := packageRequest{}
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

	id := mux.Vars(r)["packageID"]
	updatedProduct, err := app.PackageComponent.UpdateByID(r.Context(), id, &models.Package{
		Name:            reqBody.Name,
		Description:     reqBody.Description,
		LogoURL:         reqBody.LogoURL,
		RiskParameterID: reqBody.RiskParameterID,
	})
	if err != nil {
		app.HandleAPIError(
			fmt.Errorf("failed to update packages: %w", err), http.StatusInternalServerError, w,
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

func (app *App) DeletePackageByID(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["packageID"]
	err := app.PackageComponent.DeleteByID(r.Context(), id)
	if err != nil {
		app.HandleAPIError(
			fmt.Errorf("failed to delete packages: %w", err), http.StatusInternalServerError, w,
		)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
}
