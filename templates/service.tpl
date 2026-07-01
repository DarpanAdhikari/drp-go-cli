package services

import (
	"fmt"

	"{{.ModuleName}}/internal/models"
	"{{.ModuleName}}/internal/repositories"
)

// {{.Name}}Service contains business logic for {{.Name}} operations.
type {{.Name}}Service struct {
	repo *repositories.{{.Name}}Repository
}

// New{{.Name}}Service constructs a new {{.Name}}Service.
func New{{.Name}}Service(repo *repositories.{{.Name}}Repository) *{{.Name}}Service {
	return &{{.Name}}Service{repo: repo}
}

// GetAll returns all {{.PluralName}}.
func (s *{{.Name}}Service) GetAll() ([]models.{{.Name}}, error) {
	return s.repo.FindAll()
}

// GetByID returns a single {{.Name}} by ID, or an error if not found.
func (s *{{.Name}}Service) GetByID(id int64) (*models.{{.Name}}, error) {
	m, err := s.repo.FindByID(id)
	if err != nil {
		return nil, err
	}
	if m == nil {
		return nil, fmt.Errorf("{{.LowerName}} with id %d not found", id)
	}
	return m, nil
}

// Create validates and creates a new {{.Name}}.
func (s *{{.Name}}Service) Create(m *models.{{.Name}}) (*models.{{.Name}}, error) {
	// TODO: add validation logic here
	return s.repo.Create(m)
}

// Update validates and updates an existing {{.Name}}.
func (s *{{.Name}}Service) Update(m *models.{{.Name}}) (*models.{{.Name}}, error) {
	// TODO: add validation logic here
	return s.repo.Update(m)
}

// Delete removes a {{.Name}} by ID.
func (s *{{.Name}}Service) Delete(id int64) error {
	return s.repo.Delete(id)
}
