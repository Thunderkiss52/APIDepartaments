package department

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"
)

type Handler struct{ svc *Service }

func NewHandler(svc *Service) *Handler { return &Handler{svc: svc} }

func (h *Handler) Health(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.NotFound(w, r)
		return
	}
	h.writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (h *Handler) Index(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" || r.Method != http.MethodGet {
		http.NotFound(w, r)
		return
	}
	h.writeJSON(w, http.StatusOK, map[string]any{
		"service": "org-structure-api",
		"status":  "ok",
		"endpoints": []string{
			"GET /health",
			"GET /departments",
			"POST /departments",
			"GET /departments/{id}",
			"PATCH /departments/{id}",
			"DELETE /departments/{id}?mode=cascade",
			"DELETE /departments/{id}?mode=reassign&reassign_to_department_id={id}",
			"POST /departments/{id}/employees",
		},
	})
}

func (h *Handler) writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func (h *Handler) ListDepartments(w http.ResponseWriter, r *http.Request) {
	deps, err := h.svc.ListDepartments()
	if err != nil {
		h.writeErr(w, err)
		return
	}
	h.writeJSON(w, http.StatusOK, map[string]any{"departments": deps})
}

func (h *Handler) CreateDepartment(w http.ResponseWriter, r *http.Request) {
	var req CreateDepartmentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeJSON(w, 400, map[string]string{"error": "validation_error"})
		return
	}
	dep, err := h.svc.CreateDepartment(req)
	if err != nil {
		h.writeErr(w, err)
		return
	}
	h.writeJSON(w, http.StatusCreated, dep)
}

func (h *Handler) CreateEmployee(w http.ResponseWriter, r *http.Request, deptID int64) {
	var req CreateEmployeeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeJSON(w, 400, map[string]string{"error": "validation_error"})
		return
	}
	emp, err := h.svc.CreateEmployee(deptID, req)
	if err != nil {
		h.writeErr(w, err)
		return
	}
	h.writeJSON(w, http.StatusCreated, emp)
}

func (h *Handler) GetDepartment(w http.ResponseWriter, r *http.Request, id int64) {
	depth := 1
	if v := r.URL.Query().Get("depth"); v != "" {
		d, err := strconv.Atoi(v)
		if err != nil || d < 0 || d > 5 {
			h.writeJSON(w, 400, map[string]string{"error": "validation_error"})
			return
		}
		depth = d
	}
	include := true
	if v := r.URL.Query().Get("include_employees"); v != "" {
		switch v {
		case "true":
			include = true
		case "false":
			include = false
		default:
			h.writeJSON(w, 400, map[string]string{"error": "validation_error"})
			return
		}
	}
	out, err := h.svc.GetDepartment(id, depth, include)
	if err != nil {
		h.writeErr(w, err)
		return
	}
	h.writeJSON(w, 200, out)
}

func (h *Handler) UpdateDepartment(w http.ResponseWriter, r *http.Request, id int64) {
	var req UpdateDepartmentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeJSON(w, 400, map[string]string{"error": "validation_error"})
		return
	}
	dep, err := h.svc.UpdateDepartment(id, req)
	if err != nil {
		h.writeErr(w, err)
		return
	}
	h.writeJSON(w, 200, dep)
}

func (h *Handler) DeleteDepartment(w http.ResponseWriter, r *http.Request, id int64) {
	mode := strings.TrimSpace(r.URL.Query().Get("mode"))
	var reass *int64
	if v := r.URL.Query().Get("reassign_to_department_id"); v != "" {
		x, err := strconv.ParseInt(v, 10, 64)
		if err != nil {
			h.writeJSON(w, 400, map[string]string{"error": "validation_error"})
			return
		}
		reass = &x
	}
	err := h.svc.DeleteDepartment(id, mode, reass)
	if err != nil {
		h.writeErr(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) writeErr(w http.ResponseWriter, err error) {
	status := http.StatusInternalServerError
	switch {
	case errors.Is(err, ErrValidation):
		status = http.StatusBadRequest
	case errors.Is(err, ErrNotFound):
		status = http.StatusNotFound
	case errors.Is(err, ErrConflict):
		status = http.StatusConflict
	}
	h.writeJSON(w, status, map[string]string{"error": err.Error()})
}
