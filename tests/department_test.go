package tests

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"org-structure-api/internal/department"
	"org-structure-api/internal/router"
)

func TestCreateDepartmentHandler(t *testing.T) {
	h := department.NewHandler(department.NewService(department.NewRepository(nil)))
	req := httptest.NewRequest(http.MethodPost, "/departments/", bytes.NewBufferString(`{"name":"IT"}`))
	rr := httptest.NewRecorder()
	router.New(h).ServeHTTP(rr, req)
	if rr.Code != http.StatusCreated {
		t.Fatalf("unexpected status: %d", rr.Code)
	}
}

func TestHealthHandler(t *testing.T) {
	h := department.NewHandler(department.NewService(department.NewRepository(nil)))
	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	rr := httptest.NewRecorder()

	router.New(h).ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("unexpected status: %d", rr.Code)
	}
	if !strings.Contains(rr.Body.String(), `"status":"ok"`) {
		t.Fatalf("unexpected body: %s", rr.Body.String())
	}
}

func TestListDepartmentsHandler(t *testing.T) {
	h := department.NewHandler(department.NewService(department.NewRepository(nil)))
	req := httptest.NewRequest(http.MethodGet, "/departments", nil)
	rr := httptest.NewRecorder()

	router.New(h).ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("unexpected status: %d", rr.Code)
	}
	if !strings.Contains(rr.Body.String(), `"departments"`) {
		t.Fatalf("unexpected body: %s", rr.Body.String())
	}
}

func TestInvalidDepartmentIDReturnsBadRequest(t *testing.T) {
	h := department.NewHandler(department.NewService(department.NewRepository(nil)))
	req := httptest.NewRequest(http.MethodGet, "/departments/not-a-number", nil)
	rr := httptest.NewRecorder()

	router.New(h).ServeHTTP(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("unexpected status: %d", rr.Code)
	}
}

func TestInvalidIncludeEmployeesReturnsBadRequest(t *testing.T) {
	h := department.NewHandler(department.NewService(department.NewRepository(nil)))
	req := httptest.NewRequest(http.MethodGet, "/departments/1?include_employees=yes", nil)
	rr := httptest.NewRecorder()

	router.New(h).ServeHTTP(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("unexpected status: %d", rr.Code)
	}
}
