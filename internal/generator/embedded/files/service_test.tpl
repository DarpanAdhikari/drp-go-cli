package {{.DomainName}}_test

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"{{.ModuleName}}/internal/{{.DomainName}}"
)

type mock{{.Name}}Repository struct {
	mock.Mock
}

func (m *mock{{.Name}}Repository) FindAll() ([]{{.DomainName}}.{{.Name}}, error) {
	args := m.Called()
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]{{.DomainName}}.{{.Name}}), args.Error(1)
}

func (m *mock{{.Name}}Repository) FindByID(id int64) (*{{.DomainName}}.{{.Name}}, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*{{.DomainName}}.{{.Name}}), args.Error(1)
}

func (m *mock{{.Name}}Repository) Create(mdl *{{.DomainName}}.{{.Name}}) (*{{.DomainName}}.{{.Name}}, error) {
	args := m.Called(mdl)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*{{.DomainName}}.{{.Name}}), args.Error(1)
}

func (m *mock{{.Name}}Repository) Update(mdl *{{.DomainName}}.{{.Name}}) (*{{.DomainName}}.{{.Name}}, error) {
	args := m.Called(mdl)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*{{.DomainName}}.{{.Name}}), args.Error(1)
}

func (m *mock{{.Name}}Repository) Delete(id int64) error {
	args := m.Called(id)
	return args.Error(0)
}

func Test{{.Name}}Service_GetAll(t *testing.T) {
	mockRepo := new(mock{{.Name}}Repository)
	svc := {{.DomainName}}.New{{.Name}}Service(mockRepo)

	item1 := {{.DomainName}}.{{.Name}}{ID: 1}
	item2 := {{.DomainName}}.{{.Name}}{ID: 2}
	expected := []{{.DomainName}}.{{.Name}}{item1, item2}
	mockRepo.On("FindAll").Return(expected, nil).Once()

	result, err := svc.GetAll()
	require.NoError(t, err)
	require.Len(t, result, 2)
	mockRepo.AssertExpectations(t)
}

func Test{{.Name}}Service_GetByID_Success(t *testing.T) {
	mockRepo := new(mock{{.Name}}Repository)
	svc := {{.DomainName}}.New{{.Name}}Service(mockRepo)

	expected := &{{.DomainName}}.{{.Name}}{ID: 42}
	mockRepo.On("FindByID", int64(42)).Return(expected, nil).Once()

	result, err := svc.GetByID(42)
	require.NoError(t, err)
	require.NotNil(t, result)
	require.Equal(t, int64(42), result.ID)
	mockRepo.AssertExpectations(t)
}

func Test{{.Name}}Service_GetByID_NotFound(t *testing.T) {
	mockRepo := new(mock{{.Name}}Repository)
	svc := {{.DomainName}}.New{{.Name}}Service(mockRepo)

	mockRepo.On("FindByID", int64(999)).Return(nil, nil).Once()

	_, err := svc.GetByID(999)
	require.Error(t, err)
	mockRepo.AssertExpectations(t)
}

func Test{{.Name}}Service_Create(t *testing.T) {
	mockRepo := new(mock{{.Name}}Repository)
	svc := {{.DomainName}}.New{{.Name}}Service(mockRepo)

	input := &{{.DomainName}}.{{.Name}}{}
	mockRepo.On("Create", input).Return(input, nil).Once()

	result, err := svc.Create(input)
	require.NoError(t, err)
	require.NotNil(t, result)
	mockRepo.AssertExpectations(t)
}

func Test{{.Name}}Service_Update(t *testing.T) {
	mockRepo := new(mock{{.Name}}Repository)
	svc := {{.DomainName}}.New{{.Name}}Service(mockRepo)

	input := &{{.DomainName}}.{{.Name}}{ID: 1}
	mockRepo.On("Update", input).Return(input, nil).Once()

	result, err := svc.Update(input)
	require.NoError(t, err)
	require.NotNil(t, result)
	mockRepo.AssertExpectations(t)
}

func Test{{.Name}}Service_Delete(t *testing.T) {
	mockRepo := new(mock{{.Name}}Repository)
	svc := {{.DomainName}}.New{{.Name}}Service(mockRepo)

	mockRepo.On("Delete", int64(1)).Return(nil).Once()

	err := svc.Delete(1)
	require.NoError(t, err)
	mockRepo.AssertExpectations(t)
}

func Test{{.Name}}Service_Delete_Error(t *testing.T) {
	mockRepo := new(mock{{.Name}}Repository)
	svc := {{.DomainName}}.New{{.Name}}Service(mockRepo)

	mockRepo.On("Delete", int64(999)).Return(errors.New("not found")).Once()

	err := svc.Delete(999)
	require.Error(t, err)
	mockRepo.AssertExpectations(t)
}
