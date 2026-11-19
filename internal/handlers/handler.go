package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strconv"
	"time"

	"employee-management/models"
	"employee-management/repositories"
	"employee-management/services"
	"employee-management/utils"
)

type Handler struct {
	employees   *services.EmployeeService
	departments *services.DepartmentService
	timeout     time.Duration
}

func New(employeeService *services.EmployeeService, departmentService *services.DepartmentService, timeout time.Duration) *Handler {
	return &Handler{employees: employeeService, departments: departmentService, timeout: timeout}
}

func (h *Handler) Routes() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("POST /employees", h.createEmployee)
	mux.HandleFunc("GET /employees", h.listEmployees)
	mux.HandleFunc("GET /employees/search", h.searchEmployees)
	mux.HandleFunc("GET /employees/{id}", h.getEmployee)
	mux.HandleFunc("PUT /employees/{id}", h.updateEmployee)
	mux.HandleFunc("DELETE /employees/{id}", h.deleteEmployee)
	mux.HandleFunc("POST /departments", h.createDepartment)
	return h.withTimeout(mux)
}

func (h *Handler) withTimeout(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), h.timeout)
		defer cancel()
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func (h *Handler) createEmployee(w http.ResponseWriter, r *http.Request) {
	var employee models.Employee
	if err := decodeJSON(r, &employee); err != nil {
		utils.WriteError(w, http.StatusBadRequest, "Invalid request body")
		return
	}
	created, err := h.employees.Create(r.Context(), employee)
	if err != nil {
		writeServiceError(w, err, "Employee not found")
		return
	}
	utils.WriteJSON(w, http.StatusCreated, created)
}

func (h *Handler) listEmployees(w http.ResponseWriter, r *http.Request) {
	limit, err := queryInt(r, "limit", 10)
	if err != nil {
		utils.WriteError(w, http.StatusBadRequest, "Invalid limit")
		return
	}
	offset, err := queryInt(r, "offset", 0)
	if err != nil {
		utils.WriteError(w, http.StatusBadRequest, "Invalid offset")
		return
	}
	departmentID, err := optionalQueryInt(r, "departmentId")
	if err != nil {
		utils.WriteError(w, http.StatusBadRequest, "Invalid departmentId")
		return
	}
	result, err := h.employees.List(r.Context(), limit, offset, departmentID)
	if err != nil {
		writeServiceError(w, err, "Employee not found")
		return
	}
	utils.WriteJSON(w, http.StatusOK, result)
}

func (h *Handler) getEmployee(w http.ResponseWriter, r *http.Request) {
	id, ok := pathID(w, r)
	if !ok {
		return
	}
	employee, err := h.employees.GetByID(r.Context(), id)
	if err != nil {
		writeServiceError(w, err, "Employee not found")
		return
	}
	utils.WriteJSON(w, http.StatusOK, employee)
}

func (h *Handler) updateEmployee(w http.ResponseWriter, r *http.Request) {
	id, ok := pathID(w, r)
	if !ok {
		return
	}
	var update models.EmployeeUpdate
	if err := decodeJSON(r, &update); err != nil {
		utils.WriteError(w, http.StatusBadRequest, "Invalid request body")
		return
	}
	employee, err := h.employees.Update(r.Context(), id, update)
	if err != nil {
		writeServiceError(w, err, "Employee not found")
		return
	}
	utils.WriteJSON(w, http.StatusOK, employee)
}

func (h *Handler) deleteEmployee(w http.ResponseWriter, r *http.Request) {
	id, ok := pathID(w, r)
	if !ok {
		return
	}
	if err := h.employees.Delete(r.Context(), id); err != nil {
		writeServiceError(w, err, "Employee not found")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) searchEmployees(w http.ResponseWriter, r *http.Request) {
	employees, err := h.employees.Search(r.Context(), r.URL.Query().Get("keyword"))
	if err != nil {
		writeServiceError(w, err, "Employee not found")
		return
	}
	utils.WriteJSON(w, http.StatusOK, employees)
}

func (h *Handler) createDepartment(w http.ResponseWriter, r *http.Request) {
	var department models.Department
	if err := decodeJSON(r, &department); err != nil {
		utils.WriteError(w, http.StatusBadRequest, "Invalid request body")
		return
	}
	created, err := h.departments.Create(r.Context(), department)
	if err != nil {
		writeServiceError(w, err, "Department not found")
		return
	}
	utils.WriteJSON(w, http.StatusCreated, created)
}

func decodeJSON(r *http.Request, destination any) error {
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(destination); err != nil {
		return err
	}
	if err := decoder.Decode(&struct{}{}); !errors.Is(err, io.EOF) {
		return errors.New("request body must contain one JSON value")
	}
	return nil
}

func queryInt(r *http.Request, key string, defaultValue int) (int, error) {
	value := r.URL.Query().Get(key)
	if value == "" {
		return defaultValue, nil
	}
	return strconv.Atoi(value)
}

func optionalQueryInt(r *http.Request, key string) (*int, error) {
	value := r.URL.Query().Get(key)
	if value == "" {
		return nil, nil
	}
	parsed, err := strconv.Atoi(value)
	return &parsed, err
}

func pathID(w http.ResponseWriter, r *http.Request) (int, bool) {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil || id <= 0 {
		utils.WriteError(w, http.StatusBadRequest, "Invalid employee ID")
		return 0, false
	}
	return id, true
}

func writeServiceError(w http.ResponseWriter, err error, notFoundMessage string) {
	switch {
	case errors.Is(err, services.ErrInvalidInput):
		utils.WriteError(w, http.StatusBadRequest, err.Error())
	case errors.Is(err, repositories.ErrNotFound):
		utils.WriteError(w, http.StatusNotFound, notFoundMessage)
	case errors.Is(err, context.DeadlineExceeded), errors.Is(err, context.Canceled):
		utils.WriteError(w, http.StatusGatewayTimeout, "Request timeout")
	default:
		utils.WriteError(w, http.StatusInternalServerError, "Internal server error")
	}
}
