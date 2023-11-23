package server

import (
	"encoding/json"
	"fmt"
	"net/http"

	validation "github.com/go-ozzo/ozzo-validation"

	"github.com/Sinbad-HQ/kyc/core/components/usersession"
)

type AddMemberRequest struct {
	Email string `json:"email"`
	Role  string `json:"role"`
}

func (req AddMemberRequest) Validate() error {
	return validation.ValidateStruct(&req,
		validation.Field(&req.Email, validation.Required),
		validation.Field(&req.Role, validation.Required, validation.In(usersession.AdminRole, usersession.MemberRole)),
	)
}

type RemoveMemberRequest struct {
	UserID string `json:"user_id"`
}

func (req RemoveMemberRequest) Validate() error {
	return validation.ValidateStruct(&req,
		validation.Field(&req.UserID, validation.Required),
	)
}

func (app *App) ListMembers(w http.ResponseWriter, r *http.Request) {
	members, err := app.UserSessionComponent.GetOrgMembers(r.Context())
	if err != nil {
		app.HandleAPIError(
			fmt.Errorf("failed to get org members: %w", err), http.StatusInternalServerError, w,
		)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(members); err != nil {
		app.HandleAPIError(
			fmt.Errorf("failed to encode response: %w", err), http.StatusInternalServerError, w,
		)
		return
	}
}

func (app *App) AddOrgMember(w http.ResponseWriter, r *http.Request) {
	var reqBody AddMemberRequest
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

	err := app.UserSessionComponent.AddOrgMember(r.Context(), reqBody.Email, reqBody.Role)
	if err != nil {
		app.HandleAPIError(
			fmt.Errorf("failed to add org member: %w", err), http.StatusInternalServerError, w,
		)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
}

func (app *App) RemoveOrgMember(w http.ResponseWriter, r *http.Request) {
	var reqBody RemoveMemberRequest
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

	err := app.UserSessionComponent.RemoveOrgMember(r.Context(), reqBody.UserID)
	if err != nil {
		app.HandleAPIError(
			fmt.Errorf("failed to remove org member: %w", err), http.StatusInternalServerError, w,
		)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
}
