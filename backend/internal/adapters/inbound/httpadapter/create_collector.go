package httpadapter

import (
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/TimoSto/tax-return-organizer/backend/internal/ports"
)

type createCollectorRequest struct {
	Name  string `json:"name"`
	Email string `json:"email"`
}

type createCollectorResponse struct {
	ID string `json:"id"`
}

func createCollectorHandler(log *slog.Logger, create ports.CreateCollector) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req createCollectorRequest

		dec := json.NewDecoder(http.MaxBytesReader(w, r.Body, maxBodyBytes))
		dec.DisallowUnknownFields()

		if err := dec.Decode(&req); err != nil {
			writeJSON(log, w, http.StatusBadRequest, errorResponse{Error: "invalid request body"})

			return
		}

		id, err := create.Execute(r.Context(), ports.CreateCollectorCmd{Name: req.Name, Email: req.Email})
		if err != nil {
			writeError(log, w, r, err)

			return
		}

		w.Header().Set("Location", "/collectors/"+id.String())
		writeJSON(log, w, http.StatusCreated, createCollectorResponse{ID: id.String()})
	}
}
