package core

import (
	"context"
	"errors"
	"net/http"

	"github.com/Sinbad-HQ/kyc/core/components/usersession"
)

func (app *App) WithAuth(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		accessToken := r.Header.Get("Authorization")
		if accessToken == "" {
			app.HandleAPIError(
				errors.New("unauthorized"), http.StatusUnauthorized, w,
			)
			return
		}

		authContext, err := app.UserSessionComponent.GetAuthContextByAccessToken(accessToken)
		if err != nil {
			app.HandleAPIError(
				err, http.StatusUnauthorized, w,
			)
			return
		}

		ctx := context.WithValue(r.Context(), usersession.AuthCtxKey, &authContext)
		next.ServeHTTP(w, r.WithContext(ctx))
	}
}
