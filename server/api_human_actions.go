package server

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	humanActionService "github.com/mottibechhofer/otel-ai-engineer/server/service/humanaction"
	"github.com/mottibechhofer/otel-ai-engineer/server/storage"
)

// HandleListHumanActions handles GET /api/human-actions
func (s *Server) HandleListHumanActions(w http.ResponseWriter, r *http.Request) {
	// Parse query parameters
	limitStr := r.URL.Query().Get("limit")
	offsetStr := r.URL.Query().Get("offset")
	runIDStr := r.URL.Query().Get("run_id")
	statusStr := r.URL.Query().Get("status")
	requestTypeStr := r.URL.Query().Get("request_type")

	limit := 100
	if limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil {
			limit = l
		}
	}

	offset := 0
	if offsetStr != "" {
		if o, err := strconv.Atoi(offsetStr); err == nil {
			offset = o
		}
	}

	opts := storage.HumanActionListOptions{
		Limit:  limit,
		Offset: offset,
	}

	if runIDStr != "" {
		opts.RunID = &runIDStr
	}

	if statusStr != "" {
		status := storage.HumanActionStatus(statusStr)
		opts.Status = &status
	}

	if requestTypeStr != "" {
		opts.RequestType = &requestTypeStr
	}

	actions, err := s.humanActionService.ListHumanActions(r.Context(), opts)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(actions)
}

// HandleGetPendingHumanActions handles GET /api/human-actions/pending
func (s *Server) HandleGetPendingHumanActions(w http.ResponseWriter, r *http.Request) {
	actions, err := s.humanActionService.GetPendingHumanActions(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(actions)
}

// HandleGetHumanAction handles GET /api/human-actions/:actionId
func (s *Server) HandleGetHumanAction(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	actionID := vars["actionId"]

	action, err := s.humanActionService.GetHumanAction(r.Context(), actionID)
	if err != nil {
		if err.Error() == "action ID cannot be empty" {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(action)
}

// HandleRespondToHumanAction handles POST /api/human-actions/:actionId/respond
func (s *Server) HandleRespondToHumanAction(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	actionID := vars["actionId"]

	var req humanActionService.RespondToHumanActionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	action, err := s.humanActionService.RespondToHumanAction(r.Context(), actionID, req)
	if err != nil {
		if err.Error() == "action ID cannot be empty" || err.Error() == "response is required" {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		if err.Error() == "failed to get human action: human action not found" {
			http.Error(w, "Human action not found", http.StatusNotFound)
			return
		}
		if err.Error() == "human action is not pending" {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(action)
}

// HandleResumeFromHumanAction handles POST /api/human-actions/:actionId/resume
func (s *Server) HandleResumeFromHumanAction(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	actionID := vars["actionId"]

	action, err := s.humanActionService.ResumeFromHumanAction(r.Context(), actionID)
	if err != nil {
		if err.Error() == "action ID cannot be empty" {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		if err.Error() == "failed to get human action: human action not found" {
			http.Error(w, "Human action not found", http.StatusNotFound)
			return
		}
		if err.Error() == "human action must be responded to before resuming" ||
			err.Error() == "no response found on human action" {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(action)
}

// HandleDeleteHumanAction handles DELETE /api/human-actions/:actionId
func (s *Server) HandleDeleteHumanAction(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	actionID := vars["actionId"]

	if err := s.humanActionService.DeleteHumanAction(r.Context(), actionID); err != nil {
		if err.Error() == "action ID cannot be empty" {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		if err.Error()[:28] == "failed to delete human action" {
			http.Error(w, "Human action not found", http.StatusNotFound)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"status":    "deleted",
		"action_id": actionID,
	})
}

