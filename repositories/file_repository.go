package repositories

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"employee-management/models"
)

type fileData struct {
	NextEmployeeID   int                 `json:"nextEmployeeId"`
	NextDepartmentID int                 `json:"nextDepartmentId"`
	Employees        []models.Employee   `json:"employees"`
	Departments      []models.Department `json:"departments"`
}

type FileRepository struct {
	mu   sync.RWMutex
	path string
	data fileData
}

func NewFileRepository(path string) (*FileRepository, error) {
	r := &FileRepository{path: path}
	if err := r.load(); err != nil {
		return nil, err
	}
	return r, nil
}

func (r *FileRepository) load() error {
	r.mu.Lock()
	defer r.mu.Unlock()

	b, err := os.ReadFile(r.path)
	if err != nil {
		if os.IsNotExist(err) {
			r.data.NextEmployeeID = 1
			r.data.NextDepartmentID = 1
			r.data.Employees = []models.Employee{}
			r.data.Departments = []models.Department{}
			return nil
		}
		return fmt.Errorf("read storage file: %w", err)
	}
	if err := json.Unmarshal(b, &r.data); err != nil {
		return fmt.Errorf("decode storage file: %w", err)
	}
	if r.data.NextEmployeeID < 1 {
		r.data.NextEmployeeID = nextEmployeeID(r.data.Employees)
	}
	if r.data.NextDepartmentID < 1 {
		r.data.NextDepartmentID = nextDepartmentID(r.data.Departments)
	}
	if r.data.Employees == nil {
		r.data.Employees = []models.Employee{}
	}
	if r.data.Departments == nil {
		r.data.Departments = []models.Department{}
	}
	return nil
}

func (r *FileRepository) saveLocked() error {
	if err := os.MkdirAll(filepath.Dir(r.path), 0o755); err != nil {
		return fmt.Errorf("create storage directory: %w", err)
	}
	b, err := json.MarshalIndent(r.data, "", "  ")
	if err != nil {
		return fmt.Errorf("encode storage data: %w", err)
	}
	tmp, err := os.CreateTemp(filepath.Dir(r.path), ".employees-*.tmp")
	if err != nil {
		return fmt.Errorf("create temporary storage file: %w", err)
	}
	tmpName := tmp.Name()
	defer os.Remove(tmpName)
	if _, err = tmp.Write(b); err != nil {
		tmp.Close()
		return fmt.Errorf("write temporary storage file: %w", err)
	}
	if err = tmp.Close(); err != nil {
		return fmt.Errorf("close temporary storage file: %w", err)
	}
	if err = os.Rename(tmpName, r.path); err != nil {
		return fmt.Errorf("replace storage file: %w", err)
	}
	return nil
}

func (r *FileRepository) Create(ctx context.Context, employee models.Employee) (models.Employee, error) {
	if err := ctx.Err(); err != nil {
		return models.Employee{}, err
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	employee.ID = r.data.NextEmployeeID
	r.data.NextEmployeeID++
	r.data.Employees = append(r.data.Employees, employee)
	if err := r.saveLocked(); err != nil {
		r.data.Employees = r.data.Employees[:len(r.data.Employees)-1]
		r.data.NextEmployeeID--
		return models.Employee{}, err
	}
	return employee, nil
}

func (r *FileRepository) List(ctx context.Context, limit, offset int, departmentID *int) ([]models.Employee, int, error) {
	if err := ctx.Err(); err != nil {
		return nil, 0, err
	}
	r.mu.RLock()
	defer r.mu.RUnlock()
	filtered := make([]models.Employee, 0)
	for _, employee := range r.data.Employees {
		if departmentID == nil || employee.DepartmentID == *departmentID {
			filtered = append(filtered, employee)
		}
	}
	total := len(filtered)
	if offset >= total {
		return []models.Employee{}, total, nil
	}
	end := min(offset+limit, total)
	return append([]models.Employee(nil), filtered[offset:end]...), total, nil
}

func (r *FileRepository) GetByID(ctx context.Context, id int) (models.Employee, error) {
	if err := ctx.Err(); err != nil {
		return models.Employee{}, err
	}
	r.mu.RLock()
	defer r.mu.RUnlock()
	for _, employee := range r.data.Employees {
		if employee.ID == id {
			return employee, nil
		}
	}
	return models.Employee{}, ErrNotFound
}

func (r *FileRepository) Update(ctx context.Context, id int, update models.EmployeeUpdate) (models.Employee, error) {
	if err := ctx.Err(); err != nil {
		return models.Employee{}, err
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	for i := range r.data.Employees {
		if r.data.Employees[i].ID != id {
			continue
		}
		old := r.data.Employees[i]
		applyUpdate(&r.data.Employees[i], update)
		if err := r.saveLocked(); err != nil {
			r.data.Employees[i] = old
			return models.Employee{}, err
		}
		return r.data.Employees[i], nil
	}
	return models.Employee{}, ErrNotFound
}

func (r *FileRepository) Delete(ctx context.Context, id int) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	for i, employee := range r.data.Employees {
		if employee.ID != id {
			continue
		}
		old := append([]models.Employee(nil), r.data.Employees...)
		r.data.Employees = append(r.data.Employees[:i], r.data.Employees[i+1:]...)
		if err := r.saveLocked(); err != nil {
			r.data.Employees = old
			return err
		}
		return nil
	}
	return ErrNotFound
}

func (r *FileRepository) Search(ctx context.Context, keyword string) ([]models.Employee, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	r.mu.RLock()
	defer r.mu.RUnlock()
	keyword = strings.ToLower(keyword)
	result := make([]models.Employee, 0)
	for _, employee := range r.data.Employees {
		if strings.Contains(strings.ToLower(employee.Name), keyword) || strings.Contains(strings.ToLower(employee.Position), keyword) {
			result = append(result, employee)
		}
	}
	return result, nil
}

func (r *FileRepository) All(ctx context.Context) ([]models.Employee, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	r.mu.RLock()
	defer r.mu.RUnlock()
	return append([]models.Employee(nil), r.data.Employees...), nil
}

func (r *FileRepository) CreateDepartment(ctx context.Context, department models.Department) (models.Department, error) {
	if err := ctx.Err(); err != nil {
		return models.Department{}, err
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	department.ID = r.data.NextDepartmentID
	r.data.NextDepartmentID++
	r.data.Departments = append(r.data.Departments, department)
	if err := r.saveLocked(); err != nil {
		r.data.Departments = r.data.Departments[:len(r.data.Departments)-1]
		r.data.NextDepartmentID--
		return models.Department{}, err
	}
	return department, nil
}

func applyUpdate(employee *models.Employee, update models.EmployeeUpdate) {
	if update.Name != nil {
		employee.Name = *update.Name
	}
	if update.Age != nil {
		employee.Age = *update.Age
	}
	if update.Position != nil {
		employee.Position = *update.Position
	}
	if update.DepartmentID != nil {
		employee.DepartmentID = *update.DepartmentID
	}
	if update.Salary != nil {
		employee.Salary = *update.Salary
	}
}

func nextEmployeeID(employees []models.Employee) int {
	next := 1
	for _, employee := range employees {
		if employee.ID >= next {
			next = employee.ID + 1
		}
	}
	return next
}

func nextDepartmentID(departments []models.Department) int {
	next := 1
	for _, department := range departments {
		if department.ID >= next {
			next = department.ID + 1
		}
	}
	return next
}
