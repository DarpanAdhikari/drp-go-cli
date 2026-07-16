package {{.DomainName}}_test

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"{{.ModuleName}}/internal/{{.DomainName}}"
)

type mock{{.Name}}Service struct {
	mock.Mock
}

func (m *mock{{.Name}}Service) GetAll() ([]{{.DomainName}}.{{.Name}}, error) {
	args := m.Called()
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]{{.DomainName}}.{{.Name}}), args.Error(1)
}

func (m *mock{{.Name}}Service) GetByID(id int64) (*{{.DomainName}}.{{.Name}}, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*{{.DomainName}}.{{.Name}}), args.Error(1)
}

func (m *mock{{.Name}}Service) Create(mdl *{{.DomainName}}.{{.Name}}) (*{{.DomainName}}.{{.Name}}, error) {
	args := m.Called(mdl)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*{{.DomainName}}.{{.Name}}), args.Error(1)
}

func (m *mock{{.Name}}Service) Update(mdl *{{.DomainName}}.{{.Name}}) (*{{.DomainName}}.{{.Name}}, error) {
	args := m.Called(mdl)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*{{.DomainName}}.{{.Name}}), args.Error(1)
}

func (m *mock{{.Name}}Service) Delete(id int64) error {
	args := m.Called(id)
	return args.Error(0)
}

func new{{.Name}}TestHandler(svc *mock{{.Name}}Service) *{{.DomainName}}.{{.Name}}Handler {
	return {{.DomainName}}.New{{.Name}}Handler(svc)
}

func Test{{.Name}}Handler_Index(t *testing.T) {
	mockSvc := new(mock{{.Name}}Service)
	h := new{{.Name}}TestHandler(mockSvc)

	item1 := {{.DomainName}}.{{.Name}}{ID: 1}
	item2 := {{.DomainName}}.{{.Name}}{ID: 2}
	expected := []{{.DomainName}}.{{.Name}}{item1, item2}
	mockSvc.On("GetAll").Return(expected, nil).Once()

	req := httptest.NewRequest(http.MethodGet, "/{{.RouteName}}", nil)
	rr := httptest.NewRecorder()
	h.Index(rr, req)

	require.Equal(t, http.StatusOK, rr.Code)
	mockSvc.AssertExpectations(t)
}

func Test{{.Name}}Handler_Show_Success(t *testing.T) {
	mockSvc := new(mock{{.Name}}Service)
	h := new{{.Name}}TestHandler(mockSvc)

	expected := &{{.DomainName}}.{{.Name}}{ID: 1}
	mockSvc.On("GetByID", int64(1)).Return(expected, nil).Once()

	req := httptest.NewRequest(http.MethodGet, "/{{.RouteName}}/1", nil)
	rr := httptest.NewRecorder()
	h.Show(rr, req)

	require.Equal(t, http.StatusOK, rr.Code)
	mockSvc.AssertExpectations(t)
}

func Test{{.Name}}Handler_Show_InvalidID(t *testing.T) {
	h := new{{.Name}}TestHandler(nil)

	req := httptest.NewRequest(http.MethodGet, "/{{.RouteName}}/abc", nil)
	rr := httptest.NewRecorder()
	h.Show(rr, req)

	require.Equal(t, http.StatusBadRequest, rr.Code)
}

func Test{{.Name}}Handler_Store_Success(t *testing.T) {
	mockSvc := new(mock{{.Name}}Service)
	h := new{{.Name}}TestHandler(mockSvc)

	created := &{{.DomainName}}.{{.Name}}{ID: 1}
	mockSvc.On("Create", mock.Anything).Return(created, nil).Once()

	body := bytes.NewReader([]byte(`{}`))
	req := httptest.NewRequest(http.MethodPost, "/{{.RouteName}}", body)
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	h.Store(rr, req)

	require.Equal(t, http.StatusCreated, rr.Code)
	mockSvc.AssertExpectations(t)
}

func Test{{.Name}}Handler_Store_InvalidBody(t *testing.T) {
	h := new{{.Name}}TestHandler(nil)

	body := bytes.NewReader([]byte(`{invalid}`))
	req := httptest.NewRequest(http.MethodPost, "/{{.RouteName}}", body)
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	h.Store(rr, req)

	require.Equal(t, http.StatusBadRequest, rr.Code)
}

func Test{{.Name}}Handler_Update_Success(t *testing.T) {
	mockSvc := new(mock{{.Name}}Service)
	h := new{{.Name}}TestHandler(mockSvc)

	existing := &{{.DomainName}}.{{.Name}}{ID: 1}
	mockSvc.On("GetByID", int64(1)).Return(existing, nil).Once()
	mockSvc.On("Update", existing).Return(existing, nil).Once()

	body := bytes.NewReader([]byte(`{}`))
	req := httptest.NewRequest(http.MethodPut, "/{{.RouteName}}/1", body)
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	h.Update(rr, req)

	require.Equal(t, http.StatusOK, rr.Code)
	mockSvc.AssertExpectations(t)
}

func Test{{.Name}}Handler_Destroy_Success(t *testing.T) {
	mockSvc := new(mock{{.Name}}Service)
	h := new{{.Name}}TestHandler(mockSvc)

	mockSvc.On("Delete", int64(1)).Return(nil).Once()

	req := httptest.NewRequest(http.MethodDelete, "/{{.RouteName}}/1", nil)
	rr := httptest.NewRecorder()
	h.Destroy(rr, req)

	require.Equal(t, http.StatusNoContent, rr.Code)
	mockSvc.AssertExpectations(t)
}
