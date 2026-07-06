# Test Generation

When you run `drp init myapp --auth`, three test files are generated alongside
your application code. They are ready to run and serve both as a safety net
and as living examples of how to write tests for your own code.

---

## The three test files

| File | What it tests | Approach |
|---|---|---|
| `internal/user/repository_test.go` | Database operations (Create, FindByEmail, FindByID) | SQLite in-memory, real queries against a temporary table |
| `internal/user/service_test.go` | Business logic (Register, Authenticate, FindByID, password policy) | Mocked repository via `testify/mock` |
| `internal/auth/handler_test.go` | HTTP endpoints (Register, Login — validation errors, service errors, invalid credentials) | `net/http/httptest` with mocked service |

---

## Running the tests

```bash
# Run all tests in the project
go test ./...

# Run with verbose output to see each test name and duration
go test -v ./...

# Run only auth or user tests
go test -v ./internal/auth/
go test -v ./internal/user/

# Run a specific test function
go test -v -run TestHandler_Register_InvalidBody ./internal/auth/
```

Example output:

```
ok  github.com/myorg/myapp/internal/auth  0.010s
ok  github.com/myorg/myapp/internal/user  0.082s
```

If a test fails you will see a detailed message showing the expected vs actual
value, file, and line number:

```
    handler_test.go:56:
                Error Trace:    handler_test.go:56
                Error:          Not equal:
                                expected: 400
                                actual  : 422
                Test:           TestHandler_Register_ValidationError
```

---

## Repository tests (`internal/user/repository_test.go`)

These test your database layer against a real SQLite database running entirely
in memory — no external database needed.

**Key helpers:**

```go
// Creates an in-memory SQLite database with the users table
func setupTestDB(t *testing.T) *sql.DB

// Inserts a user row and returns its ID
func insertTestUser(t *testing.T, db *sql.DB, name, email, passwordHash string) int64
```

**Tests included:**

| Test | What it checks |
|---|---|
| `TestRepository_FindByEmail` | Finds a user by email, verifies name and email match |
| `TestRepository_FindByID` | Finds a user by primary key, verifies fields match |

**Pattern to follow when adding your own repository tests:**

```go
func TestMyRepo_FindByName(t *testing.T) {
    db := setupTestDB(t)
    defer db.Close()
    repo := NewRepository(db)

    id := insertTestUser(t, db, "Alice", "alice@example.com", "hash")

    // call your method and assert on the result
    u, err := repo.FindByID(id)
    require.NoError(t, err)
    require.Equal(t, "Alice", u.Name)
}
```

---

## Service tests (`internal/user/service_test.go`)

These test business logic without touching a database. A `mockRepository`
replaces the real database with programmed expectations.

**Mock methods available:**

```go
func (m *mockRepository) Create(u *User) error
func (m *mockRepository) FindByEmail(email string) (*User, error)
func (m *mockRepository) FindByID(id int64) (*User, error)
```

**Tests included:**

| Test | Scenario |
|---|---|
| `TestService_Register_Success` | Valid registration — user created successfully |
| `TestService_Register_WeakPassword` | Table-driven test: 4 weak passwords, each expecting a specific error message |
| `TestService_Authenticate_Success` | Email lookup succeeds (case-insensitive) |
| `TestService_Authenticate_UserNotFound` | Email not found — error returned |
| `TestService_FindByID` | Valid ID — user returned |

**Password policy enforced by the service:**

| Rule | Error message |
|---|---|
| Minimum 8 characters | `password must be at least 8 characters` |
| At least one uppercase letter | `password must contain an uppercase letter` |
| At least one lowercase letter | `password must contain a lowercase letter` |
| At least one digit | `password must contain a digit` |

**Pattern to follow when adding your own service tests:**

```go
func TestService_MyMethod(t *testing.T) {
    mockRepo := new(mockRepository)
    svc := NewService(mockRepo)

    // Program the mock
    mockRepo.On("FindByID", int64(1)).Return(&User{Name: "Alice"}, nil).Once()

    // Call the service
    result, err := svc.FindByID(1)

    // Assert
    require.NoError(t, err)
    require.Equal(t, "Alice", result.Name)
    mockRepo.AssertExpectations(t)
}
```

---

## Handler tests (`internal/auth/handler_test.go`)

These test HTTP endpoints using `httptest.NewRequest` and
`httptest.NewRecorder`. A `mockUserService` replaces the real user service.

**Tests included:**

| Test | What it checks |
|---|---|
| `TestHandler_Register_InvalidBody` | Malformed JSON body → 400 |
| `TestHandler_Register_ValidationError` | Missing/invalid fields → 400 |
| `TestHandler_Register_ServiceError` | Service returns error → 400 |
| `TestHandler_Login_ValidationError` | Empty email/password → 400 |
| `TestHandler_Login_InvalidCredentials` | Wrong credentials → 401 |

**Pattern to follow when adding your own handler tests:**

```go
func TestHandler_MyEndpoint(t *testing.T) {
    mockSvc := new(mockUserService)
    h := newTestHandler(mockSvc)

    // Program the mock
    mockSvc.On("Authenticate", "test@test.com", "pass").
        Return(&user.User{Name: "Alice"}, nil).Once()

    // Build the request
    body := user.LoginRequest{Email: "test@test.com", Password: "pass"}
    var buf bytes.Buffer
    json.NewEncoder(&buf).Encode(body)
    req := httptest.NewRequest(http.MethodPost, "/auth/login", &buf)
    req.Header.Set("Content-Type", "application/json")
    rr := httptest.NewRecorder()

    // Call the handler
    h.Login(rr, req)

    // Assert on the response
    require.Equal(t, http.StatusOK, rr.Code)
    mockSvc.AssertExpectations(t)
}
```

---

## Interpreting common test errors

| Error you see | Likely cause |
|---|---|
| `expected: 400, actual: 422` | Handler returned a different status code than expected — check validation logic |
| `mock: ... not called` | You programmed an expectation with `.Once()` but the code path didn't call it |
| `mock: ... unexpected call` | The code called a method you didn't program with `.On()` |
| `Error "not found" does not contain "invalid credentials"` | The error message from your service differs from what the test expects |

---

## Adding tests for CRUD-generated resources

When you run `drp create:crud product`, you get files in
`internal/product/` but no test files. You can add them yourself following
the same patterns:

**Repository test** (`internal/product/repository_test.go`):
Create a fresh SQLite table matching the migration, insert test rows using
raw SQL, then call your repository methods.

**Service test** (`internal/product/service_test.go`):
Define a `mockRepository` struct in the test file that implements the
repository interface, then test your service methods.

**Handler test** (`internal/product/handler_test.go`):
Use `httptest.NewRequest` and `httptest.NewRecorder` with your handler
directly.

---

## Quick reference

```bash
# Run everything
go test ./...

# Run with coverage
go test -cover ./...

# Run one package verbosely
go test -v ./internal/auth/

# Run a single test by name
go test -v -run TestService_Register_Success ./internal/user/
```
