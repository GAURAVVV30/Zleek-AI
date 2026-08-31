package http

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/hcl-backend/services/api-go/internal/knowledge/application"
	"github.com/hcl-backend/services/api-go/internal/knowledge/domain"
)

type Handler struct {
	knowledgeService *application.KnowledgeService
}

func NewHandler(knowledgeService *application.KnowledgeService) *Handler {
	return &Handler{
		knowledgeService: knowledgeService,
	}
}

func (h *Handler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /domains", h.handleListDomains)
	mux.HandleFunc("GET /concepts/{id}", h.handleGetConcept) // GET /concepts/{id}
	mux.HandleFunc("GET /curator/knowledge-structures", h.handleListKnowledgeStructures)
	mux.HandleFunc("POST /curator/knowledge-structures", h.handleCreateKnowledgeStructure)
	mux.HandleFunc("PATCH /curator/knowledge-structures", h.handleUpdateKnowledgeStructure)
	mux.HandleFunc("POST /curator/knowledge-structures/validate", h.handleValidateKnowledgeStructure)
}

func (h *Handler) handleListDomains(w http.ResponseWriter, r *http.Request) {
	domains, err := h.knowledgeService.ListDomains(r.Context())
	if err != nil {
		h.errorResponse(w, http.StatusInternalServerError, "Internal server error")
		return
	}
	h.successResponse(w, http.StatusOK, domains)
}

func (h *Handler) handleGetConcept(w http.ResponseWriter, r *http.Request) {
	parts := strings.Split(r.URL.Path, "/")
	if len(parts) != 3 {
		http.NotFound(w, r)
		return
	}
	id := parts[2]

	concept, err := h.knowledgeService.GetConcept(r.Context(), id)
	if err != nil {
		if errors.Is(err, domain.ErrConceptNotFound) {
			h.errorResponse(w, http.StatusNotFound, "Concept not found")
			return
		}
		h.errorResponse(w, http.StatusInternalServerError, "Internal server error")
		return
	}
	h.successResponse(w, http.StatusOK, concept)
}

func (h *Handler) handleListKnowledgeStructures(w http.ResponseWriter, r *http.Request) {
	structures, err := h.knowledgeService.ListKnowledgeStructures(r.Context())
	if err != nil {
		h.errorResponse(w, http.StatusInternalServerError, "Internal server error")
		return
	}
	h.successResponse(w, http.StatusOK, structures)
}

func (h *Handler) handleCreateKnowledgeStructure(w http.ResponseWriter, r *http.Request) {
	var req struct {
		DomainID    string `json:"domainId"`
		Title       string `json:"title"`
		Description string `json:"description"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.errorResponse(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	structure := &domain.KnowledgeStructure{
		ID:          uuid.New().String(),
		DomainID:    req.DomainID,
		Title:       req.Title,
		Description: req.Description,
		Version:     1,
		Status:      "draft",
		CreatedAt:   time.Now().UTC(),
	}

	if err := h.knowledgeService.CreateKnowledgeStructure(r.Context(), structure); err != nil {
		h.errorResponse(w, http.StatusInternalServerError, "Internal server error")
		return
	}

	h.successResponse(w, http.StatusCreated, structure)
}

func (h *Handler) handleUpdateKnowledgeStructure(w http.ResponseWriter, r *http.Request) {
	var req struct {
		ID          string `json:"id"`
		Title       string `json:"title"`
		Description string `json:"description"`
		Status      string `json:"status"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.errorResponse(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	structure := &domain.KnowledgeStructure{
		ID:          req.ID,
		Title:       req.Title,
		Description: req.Description,
		Status:      req.Status,
	}

	if err := h.knowledgeService.UpdateKnowledgeStructure(r.Context(), structure); err != nil {
		h.errorResponse(w, http.StatusInternalServerError, "Internal server error")
		return
	}

	h.successResponse(w, http.StatusOK, structure)
}

func (h *Handler) handleValidateKnowledgeStructure(w http.ResponseWriter, r *http.Request) {
	var req struct {
		ID string `json:"id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.errorResponse(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	isValid, issues, err := h.knowledgeService.ValidateKnowledgeStructure(r.Context(), req.ID)
	if err != nil {
		h.errorResponse(w, http.StatusInternalServerError, "Validation error")
		return
	}

	h.successResponse(w, http.StatusOK, map[string]interface{}{
		"isValid": isValid,
		"issues":  issues,
	})
}

func (h *Handler) successResponse(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func (h *Handler) errorResponse(w http.ResponseWriter, status int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"error": map[string]interface{}{
			"code":    status,
			"message": message,
		},
	})
}
