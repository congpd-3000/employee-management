package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"
	"time"

	"employee-management/models"
	"employee-management/repositories"
	"employee-management/services"
	"employee-management/utils"
)

func TestEmployeeHTTPFlow(t *testing.T) {
	repository, err := repositories.NewFileRepository(filepath.Join(t.TempDir(), "storage.json"))
	if err != nil {
		t.Fatalf("new repository: %v", err)
	}
	routes := New(
		services.NewEmployeeService(repository),
		services.NewDepartmentService(repository),
		time.Second,
	).Routes()

	create := request(t, routes, http.MethodPost, "/employees", `{
		"name":"Nguyen Van A","age":28,"position":"Backend Engineer","departmentId":1,"salary":2000
	}`)
	if create.Code != http.StatusCreated {
		t.Fatalf("create status=%d body=%s", create.Code, create.Body.String())
	}
	var created models.Employee
	decodeResponse(t, create, &created)
	if created.ID != 1 {
		t.Fatalf("created employee ID=%d, want 1", created.ID)
	}

	list := request(t, routes, http.MethodGet, "/employees?limit=10&offset=0&departmentId=1", "")
	if list.Code != http.StatusOK {
		t.Fatalf("list status=%d body=%s", list.Code, list.Body.String())
	}
	var result models.EmployeeList
	decodeResponse(t, list, &result)
	if result.TotalCount != 1 || len(result.Employees) != 1 {
		t.Fatalf("unexpected list: %+v", result)
	}

	update := request(t, routes, http.MethodPut, "/employees/1", `{"salary":2500}`)
	if update.Code != http.StatusOK {
		t.Fatalf("update status=%d body=%s", update.Code, update.Body.String())
	}
	var updated models.Employee
	decodeResponse(t, update, &updated)
	if updated.Salary != 2500 || updated.Name != created.Name {
		t.Fatalf("unexpected update: %+v", updated)
	}

	search := request(t, routes, http.MethodGet, "/employees/search?keyword=backend", "")
	if search.Code != http.StatusOK {
		t.Fatalf("search status=%d body=%s", search.Code, search.Body.String())
	}

	missing := request(t, routes, http.MethodGet, "/employees/999", "")
	if missing.Code != http.StatusNotFound {
		t.Fatalf("missing status=%d body=%s", missing.Code, missing.Body.String())
	}
	var responseError utils.ErrorResponse
	decodeResponse(t, missing, &responseError)
	if responseError.Error != "Employee not found" || responseError.Code != http.StatusNotFound {
		t.Fatalf("unexpected error response: %+v", responseError)
	}

	deleted := request(t, routes, http.MethodDelete, "/employees/1", "")
	if deleted.Code != http.StatusNoContent {
		t.Fatalf("delete status=%d body=%s", deleted.Code, deleted.Body.String())
	}
}

func request(t *testing.T, handler http.Handler, method, target, body string) *httptest.ResponseRecorder {
	t.Helper()
	request := httptest.NewRequest(method, target, bytes.NewBufferString(body))
	if body != "" {
		request.Header.Set("Content-Type", "application/json")
	}
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, request)
	return response
}

func decodeResponse(t *testing.T, response *httptest.ResponseRecorder, destination any) {
	t.Helper()
	if err := json.NewDecoder(response.Body).Decode(destination); err != nil {
		t.Fatalf("decode response: %v", err)
	}
}
