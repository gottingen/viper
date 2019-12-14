

package viper

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/gottingen/viper/vipercore"
)

// ServeHTTP is a simple JSON endpoint that can report on or change the current
// logging level.
//
// GET requests return a JSON description of the current logging level. PUT
// requests change the logging level and expect a payload like:
//   {"level":"info"}
//
// It's perfectly safe to change the logging level while a program is running.
func (lvl AtomicLevel) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	type errorResponse struct {
		Error string `json:"error"`
	}
	type payload struct {
		Level *vipercore.Level `json:"level"`
	}

	enc := json.NewEncoder(w)

	switch r.Method {

	case http.MethodGet:
		current := lvl.Level()
		enc.Encode(payload{Level: &current})

	case http.MethodPut:
		var req payload

		if errmess := func() string {
			if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
				return fmt.Sprintf("Request body must be well-formed JSON: %v", err)
			}
			if req.Level == nil {
				return "Must specify a logging level."
			}
			return ""
		}(); errmess != "" {
			w.WriteHeader(http.StatusBadRequest)
			enc.Encode(errorResponse{Error: errmess})
			return
		}

		lvl.SetLevel(*req.Level)
		enc.Encode(req)

	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
		enc.Encode(errorResponse{
			Error: "Only GET and PUT are supported.",
		})
	}
}
