package repositories

import (
	"context"
	"fmt"
	"path/filepath"
	"sync"
	"testing"

	"employee-management/models"
)

func TestFileRepositoryCRUDPaginationAndReload(t *testing.T) {
	path := filepath.Join(t.TempDir(), "storage.json")
	repository, err := NewFileRepository(path)
	if err != nil {
		t.Fatalf("new repository: %v", err)
	}
	ctx := context.Background()

	for i := 0; i < 3; i++ {
		_, err := repository.Create(ctx, models.Employee{
			Name: fmt.Sprintf("Employee %d", i), Age: 20 + i, Position: "Engineer", DepartmentID: i%2 + 1, Salary: 1000,
		})
		if err != nil {
			t.Fatalf("create employee: %v", err)
		}
	}

	departmentID := 1
	employees, total, err := repository.List(ctx, 1, 1, &departmentID)
	if err != nil {
		t.Fatalf("list employees: %v", err)
	}
	if total != 2 || len(employees) != 1 || employees[0].ID != 3 {
		t.Fatalf("unexpected page: total=%d employees=%+v", total, employees)
	}

	name := "Updated"
	updated, err := repository.Update(ctx, 1, models.EmployeeUpdate{Name: &name})
	if err != nil || updated.Name != name || updated.Age != 20 {
		t.Fatalf("partial update failed: employee=%+v error=%v", updated, err)
	}
	if err := repository.Delete(ctx, 2); err != nil {
		t.Fatalf("delete employee: %v", err)
	}

	reloaded, err := NewFileRepository(path)
	if err != nil {
		t.Fatalf("reload repository: %v", err)
	}
	all, err := reloaded.All(ctx)
	if err != nil {
		t.Fatalf("get all after reload: %v", err)
	}
	if len(all) != 2 || all[0].Name != name {
		t.Fatalf("unexpected persisted employees: %+v", all)
	}
}

func TestFileRepositoryConcurrentCreate(t *testing.T) {
	repository, err := NewFileRepository(filepath.Join(t.TempDir(), "storage.json"))
	if err != nil {
		t.Fatalf("new repository: %v", err)
	}

	const count = 20
	var wg sync.WaitGroup
	errorsChannel := make(chan error, count)
	for i := 0; i < count; i++ {
		wg.Add(1)
		go func(index int) {
			defer wg.Done()
			_, err := repository.Create(context.Background(), models.Employee{
				Name: fmt.Sprintf("Employee %d", index), Age: 20, Position: "Engineer", DepartmentID: 1, Salary: 1000,
			})
			errorsChannel <- err
		}(i)
	}
	wg.Wait()
	close(errorsChannel)
	for err := range errorsChannel {
		if err != nil {
			t.Fatalf("concurrent create: %v", err)
		}
	}

	employees, err := repository.All(context.Background())
	if err != nil {
		t.Fatalf("get all: %v", err)
	}
	if len(employees) != count {
		t.Fatalf("got %d employees, want %d", len(employees), count)
	}
}
