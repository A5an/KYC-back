package server

import (
	"encoding/json"
	"io/fs"
	"log"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/rs/cors"

	"github.com/Sinbad-HQ/kyc/server/api"
)

// NewHandler initialize all components http handler(s).
func (app *App) NewHandler() http.Handler {
	r := mux.NewRouter()

	subFS, err := fs.Sub(api.SwaggerUI, "swagger-ui")
	if err != nil {
		log.Fatal(err)
	}

	swaggerUI := http.StripPrefix("/swagger-ui/", http.FileServer(http.FS(subFS)))
	r.PathPrefix("/swagger-ui/").Handler(swaggerUI)

	// Serve OpenAPI spec
	r.HandleFunc("/swagger.yml", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Content-Type", "application/x-yaml")
		_, _ = w.Write([]byte(api.Swagger))
	})

	// packages endpoints
	r.HandleFunc("/packages", app.WithAuth(app.GetPackageByProviderID)).Methods(http.MethodGet)
	r.HandleFunc("/packages", app.WithAuth(app.CreateProduct)).Methods(http.MethodPost)
	r.HandleFunc("/packages/{packageID}", app.WithAuth(app.GetPackageByID)).Methods(http.MethodGet)
	r.HandleFunc("/packages/{packageID}", app.WithAuth(app.UpdatePackageByID)).Methods(http.MethodPut)
	r.HandleFunc("/packages/{packageID}", app.WithAuth(app.DeletePackageByID)).Methods(http.MethodDelete)

	r.HandleFunc("/packages/{packageID}/kyc-submissions", app.CreateKyc).Methods(http.MethodPost)
	r.HandleFunc("/packages/{packageID}/kyc-submissions", app.WithAuth(app.GetKycByProduct)).Methods(http.MethodGet)
	r.HandleFunc("/packages/{packageID}/kyc-submissions/{kycID}", app.WithAuth(app.GetKycByID)).Methods(http.MethodGet)
	r.HandleFunc("/packages/{packageID}/kyc-submissions/{kycID}", app.WithAuth(app.UpdateKycByID)).Methods(http.MethodPut)

	r.HandleFunc("/kyc-submissions", app.WithAuth(app.GetKycByOrg)).Methods(http.MethodGet)

	// risk parameters endpoints
	r.HandleFunc("/risk-parameters", app.WithAuth(app.GetRiskParameterByProviderID)).Methods(http.MethodGet)
	r.HandleFunc("/risk-parameters", app.WithAuth(app.CreateRiskParameter)).Methods(http.MethodPost)
	r.HandleFunc("/risk-parameters/{id}", app.WithAuth(app.GetRiskParameterByID)).Methods(http.MethodGet)
	r.HandleFunc("/risk-parameters/{id}", app.WithAuth(app.UpdateRiskParameterByID)).Methods(http.MethodPut)
	r.HandleFunc("/risk-parameters/{id}", app.WithAuth(app.DeleteRiskParameterByID)).Methods(http.MethodDelete)

	// kyc_submissions by provider callbacks
	r.HandleFunc("/creditcheck/callback", app.CreditChekCallback).Methods(http.MethodPost)
	r.HandleFunc("/onebrick/callback", app.OneBrickCallback).Methods(http.MethodPost)
	r.HandleFunc("/idenfy/callback", app.IdenfyCallback).Methods(http.MethodPost)
	r.HandleFunc("/okra/callback", app.OkraCallBack).Methods(http.MethodPost)

	// membership endpoints
	r.HandleFunc("/members", app.WithAuth(app.ListMembers)).Methods(http.MethodGet)
	r.HandleFunc("/add-member", app.WithAuth(app.AddOrgMember)).Methods(http.MethodPost)
	r.HandleFunc("/remove-member", app.WithAuth(app.RemoveOrgMember)).Methods(http.MethodPost)

	corsHandler := cors.AllowAll()
	return corsHandler.Handler(r)
}

func (app *App) HandleAPIError(err error, statusCode int, w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	if err = json.NewEncoder(w).Encode(map[string]interface{}{
		"message": err.Error(),
	}); err != nil {
		w.Write([]byte("Failed to encode error message"))
		return
	}
}
