package server

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

type CreateKycSubmissionRequest struct {
	FirstName     string `json:"first_name"`
	LastName      string `json:"last_name"`
	Nationality   string `json:"nationality"`
	Address       string `json:"address"`
	Email         string `json:"email"`
	PhoneNumber   string `json:"phone_number"`
	SignatureLink string `json:"signature_link"`

	// bvn country specific
	BVN string `json:"bvn"`
}

func (c CreateKycSubmissionRequest) Validate() error {
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
		//validation.Field(&c.SignatureLink, validation.Required),
		validation.Field(&c.BVN, validation.By(func(value interface{}) error {
			if bvnRequired && value.(string) == "" {
				return errors.New("bvn is required for Nigeria")
			}
			return nil
		})),
	)
}

type UpdateKycSubmissionRequest struct {
	Status string `json:"status"`
}

func (u UpdateKycSubmissionRequest) Validate() error {
	u.Status = strings.ToLower(u.Status)
	return validation.ValidateStruct(&u,
		validation.Field(&u.Status,
			validation.Required,
			validation.In(kyc.QueueStatus, kyc.AcceptedStatus, kyc.RejectedStatus),
		),
	)
}

func (app *App) CreateKyc(w http.ResponseWriter, r *http.Request) {
	var reqBody CreateKycSubmissionRequest
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

	createdProduct, err := app.KycComponent.Create(r.Context(), &models.KycSubmission{
		PackageID: mux.Vars(r)["packageID"],
		UserInfo: models.UserInfo{
			FirstName:     reqBody.FirstName,
			LastName:      reqBody.LastName,
			Nationality:   reqBody.Nationality,
			Address:       reqBody.Address,
			Email:         reqBody.Email,
			PhoneNumber:   reqBody.PhoneNumber,
			IDNumber:      reqBody.BVN,
			SignatureLink: reqBody.SignatureLink,
		},
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
	packageID := mux.Vars(r)["packageID"]
	kyc, err := app.KycComponent.GetByProductID(r.Context(), packageID)
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

func (app *App) GetKycByOrg(w http.ResponseWriter, r *http.Request) {
	kycSubmissions, err := app.KycComponent.GetByOrgID(r.Context())
	if err != nil {
		app.HandleAPIError(
			fmt.Errorf("failed to get kyc submissions: %w", err), http.StatusInternalServerError, w,
		)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(kycSubmissions); err != nil {
		app.HandleAPIError(
			fmt.Errorf("failed to encode response: %w", err), http.StatusInternalServerError, w,
		)
		return
	}
}

func (app *App) GetKycByID(w http.ResponseWriter, r *http.Request) {
	kycID := mux.Vars(r)["kycID"]
	packageID := mux.Vars(r)["packageID"]
	product, err := app.KycComponent.GetByID(r.Context(), kycID, packageID)
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
	reqBody := UpdateKycSubmissionRequest{}
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

	kycID := mux.Vars(r)["kycID"]
	packageID := mux.Vars(r)["packageID"]
	err := app.KycComponent.UpdateStatusByID(r.Context(), kycID, packageID, reqBody.Status)
	if err != nil {
		app.HandleAPIError(
			fmt.Errorf("failed to get kyc: %w", err), http.StatusInternalServerError, w,
		)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (app *App) CreditChekCallback(w http.ResponseWriter, r *http.Request) {
	providerCallback, err := app.CreditCheck.GetProviderCallback(r)
	if err != nil {
		app.HandleAPIError(
			fmt.Errorf("failed to decode body: %w", err), http.StatusInternalServerError, w,
		)
		return
	}

	// return ok status for events we do not handle
	if providerCallback.KycSubmissionID == "" && providerCallback.UserIDNumber == "" {
		w.WriteHeader(http.StatusOK)
		return
	}

	err = app.KycComponent.UpdateByProviderInfo(r.Context(), &providerCallback)
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
	providerCallback, err := app.OneBrick.GetProviderCallback(r)
	if err != nil {
		app.HandleAPIError(
			fmt.Errorf("error while handling callback: %w", err), http.StatusInternalServerError, w,
		)
		return
	}

	err = app.KycComponent.UpdateByProviderInfo(r.Context(), &providerCallback)
	if err != nil {
		app.HandleAPIError(
			fmt.Errorf("failed to update kyc: %w", err), http.StatusInternalServerError, w,
		)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
}

func (app *App) IdenfyCallback(w http.ResponseWriter, r *http.Request) {
	providerCallback, err := app.Idenfy.GetProviderCallback(r)
	if err != nil {
		app.HandleAPIError(
			fmt.Errorf("error while handling callback: %w", err), http.StatusInternalServerError, w,
		)
		return
	}

	err = app.KycComponent.UpdateByProviderInfo(r.Context(), &providerCallback)
	if err != nil {
		app.HandleAPIError(
			fmt.Errorf("failed to update kyc: %w", err), http.StatusInternalServerError, w,
		)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
}

func (app *App) OkraCallBack(w http.ResponseWriter, r *http.Request) {
	providerCallback, err := app.Okra.GetProviderCallback(r)
	if err != nil {
		app.HandleAPIError(
			fmt.Errorf("error while handling callback: %w", err), http.StatusInternalServerError, w,
		)
		return
	}

	err = app.KycComponent.UpdateByProviderInfo(r.Context(), &providerCallback)
	if err != nil {
		app.HandleAPIError(
			fmt.Errorf("failed to update kyc: %w", err), http.StatusInternalServerError, w,
		)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
}
