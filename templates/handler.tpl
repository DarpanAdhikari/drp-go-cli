package {{.DomainName}}

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
)

// {{.Name}}Handler handles HTTP requests for the {{.Name}} resource.
type {{.Name}}Handler struct {
	service *{{.Name}}Service
}

// New{{.Name}}Handler constructs a new {{.Name}}Handler.
func New{{.Name}}Handler(service *{{.Name}}Service) *{{.Name}}Handler {
	return &{{.Name}}Handler{service: service}
}

// Index handles GET /{{.RouteName}}
func (h *{{.Name}}Handler) Index(w http.ResponseWriter, r *http.Request) {
	items, err := h.service.GetAll()
	if err != nil {
		jsonError(w, http.StatusInternalServerError, err.Error())
		return
	}
	jsonResponse(w, http.StatusOK, items)
}

// Show handles GET /{{.RouteName}}/{id}
func (h *{{.Name}}Handler) Show(w http.ResponseWriter, r *http.Request) {
	id, err := idFromPath(r.URL.Path)
	if err != nil {
		jsonError(w, http.StatusBadRequest, "invalid id")
		return
	}
	item, err := h.service.GetByID(id)
	if err != nil {
		jsonError(w, http.StatusNotFound, err.Error())
		return
	}
	jsonResponse(w, http.StatusOK, item)
}

// Store handles POST /{{.RouteName}}
func (h *{{.Name}}Handler) Store(w http.ResponseWriter, r *http.Request) {
	var input struct {
		// TODO: add request fields here
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		jsonError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}
	// TODO: map input fields onto model
	item, err := h.service.Create(nil)
	if err != nil {
		jsonError(w, http.StatusUnprocessableEntity, err.Error())
		return
	}
	jsonResponse(w, http.StatusCreated, item)
}

// Update handles PUT /{{.RouteName}}/{id}
func (h *{{.Name}}Handler) Update(w http.ResponseWriter, r *http.Request) {
	id, err := idFromPath(r.URL.Path)
	if err != nil {
		jsonError(w, http.StatusBadRequest, "invalid id")
		return
	}
	item, err := h.service.GetByID(id)
	if err != nil {
		jsonError(w, http.StatusNotFound, err.Error())
		return
	}
	// TODO: decode request body and update fields on item
	updated, err := h.service.Update(item)
	if err != nil {
		jsonError(w, http.StatusUnprocessableEntity, err.Error())
		return
	}
	jsonResponse(w, http.StatusOK, updated)
}

// Destroy handles DELETE /{{.RouteName}}/{id}
func (h *{{.Name}}Handler) Destroy(w http.ResponseWriter, r *http.Request) {
	id, err := idFromPath(r.URL.Path)
	if err != nil {
		jsonError(w, http.StatusBadRequest, "invalid id")
		return
	}
	if err := h.service.Delete(id); err != nil {
		jsonError(w, http.StatusNotFound, err.Error())
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// ── helpers ────────────────────────────────────────────────────────────────

func jsonResponse(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(payload)
}

func jsonError(w http.ResponseWriter, status int, message string) {
	jsonResponse(w, status, map[string]string{"error": message})
}

func idFromPath(path string) (int64, error) {
	parts := strings.Split(strings.Trim(path, "/"), "/")
	return strconv.ParseInt(parts[len(parts)-1], 10, 64)
}
