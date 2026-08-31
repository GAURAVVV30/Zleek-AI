package http

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"github.com/hcl-backend/services/api-go/internal/knowledge/application"
	"github.com/hcl-backend/services/api-go/internal/knowledge/domain"
	"github.com/hcl-backend/services/api-go/internal/platform/httpx"
)

type Handler struct {
	knowledgeService *application.KnowledgeService
}

func NewHandler(knowledgeService *application.KnowledgeService) *Handler {
	return &Handler{knowledgeService: knowledgeService}
}

func (h *Handler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /domains", h.handleListDomains)
	mux.HandleFunc("GET /concepts/{id}", h.handleGetConcept)

	mux.HandleFunc("GET /curator/knowledge-structures", h.requireAdminOrCurator(h.handleListStructures))
	mux.HandleFunc("POST /curator/knowledge-structures", h.requireAdminOrCurator(h.handleCreateStructure))
	mux.HandleFunc("PATCH /curator/knowledge-structures", h.requireAdminOrCurator(h.handleUpdateStructureStatus))
	mux.HandleFunc("POST /curator/knowledge-structures/validate", h.requireAdminOrCurator(h.handleValidateStructure))
}

func roleOK(r *http.Request) bool {
	role := r.Header.Get("X-User-Role")
	return role == "admin" || role == "curator"
}

func (h *Handler) requireAdminOrCurator(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if !roleOK(r) {
			httpx.Error(w, http.StatusForbidden, "Forbidden")
			return
		}
		next(w, r)
	}
}

func (h *Handler) handleListDomains(w http.ResponseWriter, r *http.Request) {
	domains, err := h.knowledgeService.ListDomains(r.Context())
	if err != nil {
		httpx.Error(w, http.StatusInternalServerError, "Internal server error")
		return
	}
	httpx.Envelope(w, http.StatusOK, domains)
}

func (h *Handler) handleGetConcept(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	learnerID := r.Header.Get("X-User-ID")
	view, err := h.knowledgeService.GetConceptView(r.Context(), learnerID, id)
	if err != nil {
		if errors.Is(err, domain.ErrConceptNotFound) {
			httpx.Error(w, http.StatusNotFound, "Concept not found")
			return
		}
		if strings.Contains(err.Error(), "is locked") {
			httpx.Error(w, http.StatusForbidden, err.Error())
			return
		}
		httpx.Error(w, http.StatusInternalServerError, "Internal server error")
		return
	}
	httpx.Envelope(w, http.StatusOK, view)
}

func (h *Handler) handleListStructures(w http.ResponseWriter, r *http.Request) {
	structures, err := h.knowledgeService.ListStructures(r.Context())
	if err != nil {
		httpx.Error(w, http.StatusInternalServerError, "Internal server error")
		return
	}
	httpx.Envelope(w, http.StatusOK, structures)
}

func (h *Handler) handleCreateStructure(w http.ResponseWriter, r *http.Request) {
	var req struct {
		DomainID string `json:"domainId"`
		Status   string `json:"status"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpx.Error(w, http.StatusBadRequest, "Invalid request body")
		return
	}
	actor := r.Header.Get("X-User-ID")
	structure := &domain.KnowledgeStructure{DomainID: req.DomainID, Status: req.Status}
	if err := h.knowledgeService.CreateStructure(r.Context(), structure, actor); err != nil {
		httpx.Error(w, http.StatusBadRequest, err.Error())
		return
	}
	httpx.Envelope(w, http.StatusCreated, structure)
}

func (h *Handler) handleUpdateStructureStatus(w http.ResponseWriter, r *http.Request) {
	var req struct {
		ID     string `json:"id"`
		Status string `json:"status"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpx.Error(w, http.StatusBadRequest, "Invalid request body")
		return
	}
	if err := h.knowledgeService.UpdateStructure(r.Context(), req.ID, req.Status); err != nil {
		httpx.Error(w, http.StatusBadRequest, err.Error())
		return
	}
	httpx.Envelope(w, http.StatusOK, map[string]string{"message": "Knowledge structure updated."})
}

func (h *Handler) handleValidateStructure(w http.ResponseWriter, r *http.Request) {
	var req struct {
		ID string `json:"id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpx.Error(w, http.StatusBadRequest, "Invalid request body")
		return
	}
	valid, message, err := h.knowledgeService.ValidateStructure(r.Context(), req.ID)
	if err != nil {
		httpx.Error(w, http.StatusInternalServerError, "Validation error")
		return
	}
	httpx.Envelope(w, http.StatusOK, map[string]any{
		"valid":   valid,
		"message": message,
	})
}
