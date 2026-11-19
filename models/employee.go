package models

type Employee struct {
	ID           int     `json:"id"`
	Name         string  `json:"name"`
	Age          int     `json:"age"`
	Position     string  `json:"position"`
	DepartmentID int     `json:"departmentId"`
	Salary       float64 `json:"salary"`
}

type EmployeeUpdate struct {
	Name         *string  `json:"name"`
	Age          *int     `json:"age"`
	Position     *string  `json:"position"`
	DepartmentID *int     `json:"departmentId"`
	Salary       *float64 `json:"salary"`
}

type EmployeeList struct {
	TotalCount int        `json:"totalCount"`
	Employees  []Employee `json:"employees"`
}
