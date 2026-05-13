package router

import (
	"net/http"
	"strconv"
	"strings"

	"org-structure-api/internal/department"
)

func New(h *department.Handler) http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/", h.Index)
	mux.HandleFunc("/health", h.Health)
	mux.HandleFunc("/departments/", func(w http.ResponseWriter, r *http.Request) {
		path := strings.TrimPrefix(r.URL.Path, "/departments/")
		parts := strings.Split(strings.Trim(path, "/"), "/")
		if len(parts) == 1 && parts[0] == "" {
			switch r.Method {
			case http.MethodGet:
				h.ListDepartments(w, r)
			case http.MethodPost:
				h.CreateDepartment(w, r)
			default:
				http.NotFound(w, r)
			}
			return
		}
		if len(parts) == 1 && parts[0] != "" {
			id, err := parseID(parts[0])
			if err != nil {
				writeValidationError(w)
				return
			}
			switch r.Method {
			case http.MethodGet:
				h.GetDepartment(w, r, id)
			case http.MethodPatch:
				h.UpdateDepartment(w, r, id)
			case http.MethodDelete:
				h.DeleteDepartment(w, r, id)
			default:
				http.NotFound(w, r)
			}
			return
		}
		if len(parts) == 2 && parts[1] == "employees" && r.Method == http.MethodPost {
			id, err := parseID(parts[0])
			if err != nil {
				writeValidationError(w)
				return
			}
			h.CreateEmployee(w, r, id)
			return
		}
		if r.Method == http.MethodPost && strings.HasSuffix(r.URL.Path, "/departments/") {
			h.CreateDepartment(w, r)
			return
		}
		http.NotFound(w, r)
	})
	mux.HandleFunc("/departments", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			h.ListDepartments(w, r)
		case http.MethodPost:
			h.CreateDepartment(w, r)
		default:
			http.NotFound(w, r)
		}
	})
	return mux
}

func parseID(raw string) (int64, error) {
	id, err := strconv.ParseInt(raw, 10, 64)
	if err != nil || id <= 0 {
		return 0, strconv.ErrSyntax
	}
	return id, nil
}

func writeValidationError(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusBadRequest)
	_, _ = w.Write([]byte(`{"error":"validation_error"}` + "\n"))
}
