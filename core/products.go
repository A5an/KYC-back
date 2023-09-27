package core

import (
	"encoding/json"
	"fmt"
	"net/http"

	validation "github.com/go-ozzo/ozzo-validation"
	"github.com/gorilla/mux"

	"github.com/Sinbad-HQ/kyc/core/components/product/models"
)

type CreateProductRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	ImageURL    string `json:"image_url"`
}

func (c CreateProductRequest) Validate() error {
	return validation.ValidateStruct(&c,
		validation.Field(&c.Name, validation.Required),
		//validation.Field(&c.ImageURL, validation.Required),
	)
}

func (app *App) CreateProduct(w http.ResponseWriter, r *http.Request) {
	reqBody := CreateProductRequest{}
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

	createdProduct, err := app.ProductComponent.Create(r.Context(), &models.Product{
		Name:        reqBody.Name,
		Description: reqBody.Description,
		ImageURL:    reqBody.ImageURL,
	})
	if err != nil {
		app.HandleAPIError(
			fmt.Errorf("failed to create product: %w", err), http.StatusInternalServerError, w,
		)
		return
	}

	if err := json.NewEncoder(w).Encode(createdProduct); err != nil {
		app.HandleAPIError(
			fmt.Errorf("failed to encode response: %w", err), http.StatusInternalServerError, w,
		)
		return
	}
}

func (app *App) GetProductByID(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["productID"]
	product, err := app.ProductComponent.GetByID(r.Context(), id)
	if err != nil {
		app.HandleAPIError(
			fmt.Errorf("failed to get provider: %w", err), http.StatusInternalServerError, w,
		)
		return
	}

	if err := json.NewEncoder(w).Encode(product); err != nil {
		app.HandleAPIError(
			fmt.Errorf("failed to encode response: %w", err), http.StatusInternalServerError, w,
		)
		return
	}
}

func (app *App) GetProductsByProviderID(w http.ResponseWriter, r *http.Request) {
	products, err := app.ProductComponent.GetByProviderID(r.Context())
	if err != nil {
		app.HandleAPIError(
			fmt.Errorf("failed to get provider products: %w", err), http.StatusInternalServerError, w,
		)
		return
	}

	if err := json.NewEncoder(w).Encode(products); err != nil {
		app.HandleAPIError(
			fmt.Errorf("failed to encode response: %w", err), http.StatusInternalServerError, w,
		)
		return
	}
}
