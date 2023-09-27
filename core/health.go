package core

import (
	"encoding/json"
	"net/http"
)

type Info struct {
	Environment string
}

// ReplyInfo returns the server information.
func (app *App) ReplyInfo(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"env": app.Info.Environment,
	})
}
