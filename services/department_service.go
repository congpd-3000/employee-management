package services

import (
	"context"
	"fmt"
	"strings"

	"employee-management/models"
	"employee-management/repositories"
)

type DepartmentService struct {
	repository repositories.DepartmentRepository
}

func NewDepartmentService(repository repositories.DepartmentRepository) *DepartmentService {
	return &DepartmentService{repository: repository}
}

func (s *DepartmentService) Create(ctx context.Context, department models.Department) (models.Department, error) {
	if strings.TrimSpace(department.Name) == "" {
		return models.Department{}, fmt.Errorf("%w: name is required", ErrInvalidInput)
	}
	return s.repository.CreateDepartment(ctx, department)
}
