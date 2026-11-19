package main

import (
	"log"
	"net/http"
	"time"

	"employee-management/internal/handlers"
	"employee-management/repositories"
	"employee-management/services"
)

func main() {
	repository, err := repositories.NewFileRepository("data/storage.json")
	if err != nil {
		log.Fatal(err)
	}
	employeeService := services.NewEmployeeService(repository)
	departmentService := services.NewDepartmentService(repository)
	handler := handlers.New(employeeService, departmentService, 5*time.Second)

	server := &http.Server{
		Addr:              ":3000",
		Handler:           handler.Routes(),
		ReadHeaderTimeout: 5 * time.Second,
	}
	log.Printf("employee management server listening on %s", server.Addr)
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatal(err)
	}
}
