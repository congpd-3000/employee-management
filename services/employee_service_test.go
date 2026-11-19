package services

import (
	"context"
	"encoding/csv"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"testing"

	"employee-management/models"
	"employee-management/repositories"
)

type employeeRepositoryStub struct {
	employees []models.Employee
	created   bool
}

func (r *employeeRepositoryStub) Create(_ context.Context, employee models.Employee) (models.Employee, error) {
	r.created = true
	employee.ID = 1
	return employee, nil
}

func (r *employeeRepositoryStub) List(_ context.Context, _, _ int, _ *int) ([]models.Employee, int, error) {
	return r.employees, len(r.employees), nil
}

func (r *employeeRepositoryStub) GetByID(_ context.Context, id int) (models.Employee, error) {
	for _, employee := range r.employees {
		if employee.ID == id {
			return employee, nil
		}
	}
	return models.Employee{}, repositories.ErrNotFound
}

func (r *employeeRepositoryStub) Update(_ context.Context, id int, update models.EmployeeUpdate) (models.Employee, error) {
	employee, err := r.GetByID(context.Background(), id)
	if err != nil {
		return models.Employee{}, err
	}
	if update.Name != nil {
		employee.Name = *update.Name
	}
	return employee, nil
}

func (r *employeeRepositoryStub) Delete(_ context.Context, id int) error {
	_, err := r.GetByID(context.Background(), id)
	return err
}

func (r *employeeRepositoryStub) Search(_ context.Context, _ string) ([]models.Employee, error) {
	return r.employees, nil
}

func (r *employeeRepositoryStub) All(_ context.Context) ([]models.Employee, error) {
	return r.employees, nil
}

func TestCreateValidatesAgeAndSalary(t *testing.T) {
	repository := &employeeRepositoryStub{}
	service := NewEmployeeService(repository)

	_, err := service.Create(context.Background(), models.Employee{Name: "A", Age: 0, Salary: 100})
	if !errors.Is(err, ErrInvalidInput) {
		t.Fatalf("expected ErrInvalidInput, got %v", err)
	}
	if repository.created {
		t.Fatal("repository must not be called for invalid data")
	}

	created, err := service.Create(context.Background(), models.Employee{Name: "A", Age: 28, Salary: 2000})
	if err != nil {
		t.Fatalf("create valid employee: %v", err)
	}
	if created.ID != 1 || !repository.created {
		t.Fatalf("unexpected created employee: %+v", created)
	}
}

func TestExportWritesJSONAndCSV(t *testing.T) {
	repository := &employeeRepositoryStub{employees: []models.Employee{{
		ID: 1, Name: "Nguyen Van A", Age: 28, Position: "Backend Engineer", DepartmentID: 1, Salary: 2000,
	}}}
	service := NewEmployeeService(repository)
	directory := t.TempDir()
	jsonPath := filepath.Join(directory, "employees.json")
	csvPath := filepath.Join(directory, "employees.csv")

	if err := service.Export(context.Background(), jsonPath, csvPath); err != nil {
		t.Fatalf("export: %v", err)
	}

	jsonFile, err := os.Open(jsonPath)
	if err != nil {
		t.Fatalf("open JSON export: %v", err)
	}
	defer jsonFile.Close()
	var employees []models.Employee
	if err := json.NewDecoder(jsonFile).Decode(&employees); err != nil {
		t.Fatalf("decode JSON export: %v", err)
	}
	if len(employees) != 1 || employees[0].Name != "Nguyen Van A" {
		t.Fatalf("unexpected JSON export: %+v", employees)
	}

	csvFile, err := os.Open(csvPath)
	if err != nil {
		t.Fatalf("open CSV export: %v", err)
	}
	defer csvFile.Close()
	records, err := csv.NewReader(csvFile).ReadAll()
	if err != nil {
		t.Fatalf("read CSV export: %v", err)
	}
	if len(records) != 2 || records[1][1] != "Nguyen Van A" {
		t.Fatalf("unexpected CSV export: %+v", records)
	}
}
