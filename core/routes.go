package core

import (
	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/rs/cors"
)

// NewHandler initialize all components http handler(s).
func (app *App) NewHandler() http.Handler {
	r := mux.NewRouter()

	r.HandleFunc("/info", app.ReplyInfo).Methods(http.MethodGet)

	// products endpoints
	r.HandleFunc("/products", app.WithAuth(app.GetProductsByProviderID)).Methods(http.MethodGet)
	r.HandleFunc("/products", app.WithAuth(app.CreateProduct)).Methods(http.MethodPost)
	r.HandleFunc("/products/{productID}", app.WithAuth(app.GetProductByID)).Methods(http.MethodGet)

	r.HandleFunc("/products/{productID}/kyc", app.CreateKyc).Methods(http.MethodPost)
	r.HandleFunc("/products/{productID}/kyc", app.WithAuth(app.GetKycByProduct)).Methods(http.MethodGet)
	r.HandleFunc("/products/{productID}/kyc/{kycID}", app.WithAuth(app.GetKycByID)).Methods(http.MethodGet)
	r.HandleFunc("/products/{productID}/kyc/{kycID}", app.WithAuth(app.UpdateKycByID)).Methods(http.MethodPut)
	//r.HandleFunc("/products/{id}", nil).Methods(http.MethodPut)
	//r.HandleFunc("/products/{id}", nil).Methods(http.MethodDelete)

	// risk-parameters
	r.HandleFunc("/risk-parameters", app.WithAuth(app.CreateRiskParameter)).Methods(http.MethodPost)
	r.HandleFunc("/risk-parameters", app.WithAuth(app.GetRiskParameters)).Methods(http.MethodGet)
	r.HandleFunc("/risk-parameters/{riskParameterID}", app.WithAuth(app.UpdateRiskParameter)).Methods(http.MethodPut)

	// kyc_submissions by provider callbacks
	r.HandleFunc("/creditcheck/callback", app.CreditChekCallback).Methods(http.MethodPost)
	r.HandleFunc("/onebrick/callback", app.OneBrickCallback).Methods(http.MethodPost)
	r.HandleFunc("/idenfy/callback", app.IdenfyCallback).Methods(http.MethodPost)

	corsHandler := cors.AllowAll()
	//corsHandler := cors.New(cors.Options{
	//	AllowedOrigins:   []string{"*"},
	//	AllowCredentials: true,
	//	AllowedMethods:   []string{"GET", "POST", "PUT", "PATCH", "DELETE"},
	//	AllowedHeaders:   []string{"Content-Type", "Authorization"},
	//})

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
