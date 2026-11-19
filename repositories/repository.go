package repositories

import (
	"context"

	"employee-management/models"
)

type EmployeeRepository interface {
	Create(ctx context.Context, employee models.Employee) (models.Employee, error)
	List(ctx context.Context, limit, offset int, departmentID *int) ([]models.Employee, int, error)
	GetByID(ctx context.Context, id int) (models.Employee, error)
	Update(ctx context.Context, id int, update models.EmployeeUpdate) (models.Employee, error)
	Delete(ctx context.Context, id int) error
	Search(ctx context.Context, keyword string) ([]models.Employee, error)
	All(ctx context.Context) ([]models.Employee, error)
}

type DepartmentRepository interface {
	CreateDepartment(ctx context.Context, department models.Department) (models.Department, error)
}
