package core

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"

	validation "github.com/go-ozzo/ozzo-validation"
	"github.com/gorilla/mux"

	"github.com/Sinbad-HQ/kyc/core/components/kyc"
	"github.com/Sinbad-HQ/kyc/core/components/kyc/models"
)

type CreateKycRequest struct {
	FirstName   string `json:"first_name"`
	LastName    string `json:"last_name"`
	Nationality string `json:"nationality"`
	Address     string `json:"address"`
	Email       string `json:"email"`
	PhoneNumber string `json:"phone_number"`

	// bvn country specific
	BVN        string `json:"bvn"`
	ProviderID string `json:"provider_id"`
}

// TODO: we need more generic validation for these different providers
func (c CreateKycRequest) Validate() error {
	bvnRequired := false
	if strings.ToLower(c.Nationality) == "nigeria" {
		bvnRequired = true
	}

	return validation.ValidateStruct(&c,
		validation.Field(&c.FirstName, validation.Required),
		validation.Field(&c.LastName, validation.Required),
		validation.Field(&c.Nationality, validation.Required),
		validation.Field(&c.Address, validation.Required),
		validation.Field(&c.Email, validation.Required),
		validation.Field(&c.PhoneNumber, validation.Required),
		validation.Field(&c.ProviderID, validation.Required),
		validation.Field(&c.BVN, validation.By(func(value interface{}) error {
			if bvnRequired && value.(string) == "" {
				return errors.New("bvn is required for Nigeria")
			}
			return nil
		})),
	)
}

type UpdateKycRequest struct {
	Status string `json:"status"`
}

func (u UpdateKycRequest) Validate() error {
	u.Status = strings.ToLower(u.Status)
	return validation.ValidateStruct(&u,
		validation.Field(&u.Status,
			validation.Required,
			validation.In(kyc.QueStatus, kyc.AprovedStatus, kyc.RejectedStatus),
		),
	)
}

func (app *App) CreateKyc(w http.ResponseWriter, r *http.Request) {
	reqBody := CreateKycRequest{}
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

	createdProduct, err := app.KycComponent.Create(r.Context(), &models.Kyc{
		ID:          reqBody.BVN,
		ProductID:   mux.Vars(r)["productID"],
		ProviderID:  reqBody.ProviderID,
		FirstName:   reqBody.FirstName,
		LastName:    reqBody.LastName,
		Nationality: reqBody.Nationality,
		Email:       reqBody.Email,
		PhoneNumber: reqBody.PhoneNumber,
		Address:     reqBody.Address,
	})
	if err != nil {
		app.HandleAPIError(
			fmt.Errorf("failed to create kyc: %w", err), http.StatusInternalServerError, w,
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

func (app *App) GetKycByProduct(w http.ResponseWriter, r *http.Request) {
	productID := mux.Vars(r)["productID"]
	kyc, err := app.KycComponent.GetByProductID(r.Context(), productID)
	if err != nil {
		app.HandleAPIError(
			fmt.Errorf("failed to get kyc: %w", err), http.StatusInternalServerError, w,
		)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(kyc); err != nil {
		app.HandleAPIError(
			fmt.Errorf("failed to encode response: %w", err), http.StatusInternalServerError, w,
		)
		return
	}
}

func (app *App) GetKycByID(w http.ResponseWriter, r *http.Request) {
	kycID := mux.Vars(r)["kycID"]
	product, err := app.KycComponent.GetByID(r.Context(), kycID)
	if err != nil {
		app.HandleAPIError(
			fmt.Errorf("failed to get kyc: %w", err), http.StatusInternalServerError, w,
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

func (app *App) UpdateKycByID(w http.ResponseWriter, r *http.Request) {
	reqBody := UpdateKycRequest{}
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

	err := app.KycComponent.UpdateStatusByID(r.Context(), mux.Vars(r)["kycID"], reqBody.Status)
	if err != nil {
		app.HandleAPIError(
			fmt.Errorf("failed to get kyc: %w", err), http.StatusInternalServerError, w,
		)
		return
	}

	w.WriteHeader(http.StatusOK)
	return
}

func (app *App) CreditChekCallback(w http.ResponseWriter, r *http.Request) {
	userInfo, _, err := app.CreditChek.GetUserInfoFromCallback(r)
	if err != nil {
		app.HandleAPIError(
			fmt.Errorf("failed to decode body: %w", err), http.StatusInternalServerError, w,
		)
		return
	}

	// return ok status for events we do not handle
	if userInfo.KycID == "" {
		w.WriteHeader(http.StatusOK)
		return
	}

	err = app.KycComponent.UpdateByID(r.Context(), &userInfo)
	if err != nil {
		app.HandleAPIError(
			fmt.Errorf("failed to update kyc: %w", err), http.StatusInternalServerError, w,
		)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
}

func (app *App) OneBrickCallback(w http.ResponseWriter, r *http.Request) {
	userInfo, _, err := app.OneBrick.GetUserInfoFromCallback(r)
	if err != nil {
		app.HandleAPIError(
			fmt.Errorf("error while handling callback: %w", err), http.StatusInternalServerError, w,
		)
		return
	}

	err = app.KycComponent.UpdateByID(r.Context(), &userInfo)
	if err != nil {
		app.HandleAPIError(
			fmt.Errorf("failed to update kyc: %w", err), http.StatusInternalServerError, w,
		)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
}
