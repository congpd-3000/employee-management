package services

import (
	"context"
	"encoding/csv"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"

	"employee-management/models"
	"employee-management/repositories"
)

type EmployeeService struct {
	repository repositories.EmployeeRepository
}

func NewEmployeeService(repository repositories.EmployeeRepository) *EmployeeService {
	return &EmployeeService{repository: repository}
}

func (s *EmployeeService) Create(ctx context.Context, employee models.Employee) (models.Employee, error) {
	if err := validateEmployee(employee); err != nil {
		return models.Employee{}, err
	}
	return s.repository.Create(ctx, employee)
}

func (s *EmployeeService) List(ctx context.Context, limit, offset int, departmentID *int) (models.EmployeeList, error) {
	if limit <= 0 || offset < 0 || (departmentID != nil && *departmentID <= 0) {
		return models.EmployeeList{}, fmt.Errorf("%w: pagination or departmentId is invalid", ErrInvalidInput)
	}
	employees, total, err := s.repository.List(ctx, limit, offset, departmentID)
	return models.EmployeeList{TotalCount: total, Employees: employees}, err
}

func (s *EmployeeService) GetByID(ctx context.Context, id int) (models.Employee, error) {
	if id <= 0 {
		return models.Employee{}, fmt.Errorf("%w: id must be positive", ErrInvalidInput)
	}
	return s.repository.GetByID(ctx, id)
}

func (s *EmployeeService) Update(ctx context.Context, id int, update models.EmployeeUpdate) (models.Employee, error) {
	if id <= 0 {
		return models.Employee{}, fmt.Errorf("%w: id must be positive", ErrInvalidInput)
	}
	if update.Name == nil && update.Age == nil && update.Position == nil && update.DepartmentID == nil && update.Salary == nil {
		return models.Employee{}, fmt.Errorf("%w: at least one field is required", ErrInvalidInput)
	}
	if update.Name != nil && strings.TrimSpace(*update.Name) == "" {
		return models.Employee{}, fmt.Errorf("%w: name is required", ErrInvalidInput)
	}
	if update.Age != nil && *update.Age <= 0 {
		return models.Employee{}, fmt.Errorf("%w: age must be greater than zero", ErrInvalidInput)
	}
	if update.Salary != nil && *update.Salary <= 0 {
		return models.Employee{}, fmt.Errorf("%w: salary must be greater than zero", ErrInvalidInput)
	}
	return s.repository.Update(ctx, id, update)
}

func (s *EmployeeService) Delete(ctx context.Context, id int) error {
	if id <= 0 {
		return fmt.Errorf("%w: id must be positive", ErrInvalidInput)
	}
	return s.repository.Delete(ctx, id)
}

func (s *EmployeeService) Search(ctx context.Context, keyword string) ([]models.Employee, error) {
	keyword = strings.TrimSpace(keyword)
	if keyword == "" {
		return nil, fmt.Errorf("%w: keyword is required", ErrInvalidInput)
	}
	return s.repository.Search(ctx, keyword)
}

func (s *EmployeeService) Export(ctx context.Context, jsonPath, csvPath string) error {
	employees, err := s.repository.All(ctx)
	if err != nil {
		return err
	}

	var wg sync.WaitGroup
	var mu sync.Mutex
	var exportErrors []error
	run := func(export func() error) {
		defer wg.Done()
		if err := export(); err != nil {
			mu.Lock()
			exportErrors = append(exportErrors, err)
			mu.Unlock()
		}
	}
	wg.Add(2)
	go run(func() error { return exportJSON(ctx, jsonPath, employees) })
	go run(func() error { return exportCSV(ctx, csvPath, employees) })
	wg.Wait()
	return errors.Join(exportErrors...)
}

func validateEmployee(employee models.Employee) error {
	if strings.TrimSpace(employee.Name) == "" {
		return fmt.Errorf("%w: name is required", ErrInvalidInput)
	}
	if employee.Age <= 0 {
		return fmt.Errorf("%w: age must be greater than zero", ErrInvalidInput)
	}
	if employee.Salary <= 0 {
		return fmt.Errorf("%w: salary must be greater than zero", ErrInvalidInput)
	}
	return nil
}

func exportJSON(ctx context.Context, path string, employees []models.Employee) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	b, err := json.MarshalIndent(employees, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, b, 0o644)
}

func exportCSV(ctx context.Context, path string, employees []models.Employee) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	file, err := os.Create(path)
	if err != nil {
		return err
	}
	writer := csv.NewWriter(file)
	records := [][]string{{"id", "name", "age", "position", "departmentId", "salary"}}
	for _, employee := range employees {
		records = append(records, []string{
			strconv.Itoa(employee.ID), employee.Name, strconv.Itoa(employee.Age), employee.Position,
			strconv.Itoa(employee.DepartmentID), strconv.FormatFloat(employee.Salary, 'f', -1, 64),
		})
	}
	writeErr := writer.WriteAll(records)
	closeErr := file.Close()
	return errors.Join(writeErr, closeErr)
}
