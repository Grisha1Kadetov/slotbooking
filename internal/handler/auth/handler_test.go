package auth_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	handlerauth "github.com/Grisha1Kadetov/slotbooking/internal/handler/auth"
	"github.com/Grisha1Kadetov/slotbooking/internal/model/user"
	"github.com/google/uuid"
)

type authServiceMock struct {
	registerFunc           func(ctx context.Context, email, password string, role user.Role) (user.User, error)
	loginFunc              func(ctx context.Context, email, password string) (string, error)
	generateDummyTokenFunc func(role user.Role) (string, error)
}

func (m *authServiceMock) Register(ctx context.Context, email, password string, role user.Role) (user.User, error) {
	if m.registerFunc == nil {
		return user.User{}, errors.New("unexpected Register call")
	}
	return m.registerFunc(ctx, email, password, role)
}

func (m *authServiceMock) Login(ctx context.Context, email, password string) (string, error) {
	if m.loginFunc == nil {
		return "", errors.New("unexpected Login call")
	}
	return m.loginFunc(ctx, email, password)
}

func (m *authServiceMock) GenerateDummyToken(role user.Role) (string, error) {
	if m.generateDummyTokenFunc == nil {
		return "", errors.New("unexpected GenerateDummyToken call")
	}
	return m.generateDummyTokenFunc(role)
}

type errorResponse struct {
	Error struct {
		Code    string `json:"code"`
		Message string `json:"message"`
	} `json:"error"`
}

func TestAuthHandler_Register(t *testing.T) {
	t.Run("invalid json", func(t *testing.T) {
		h := handlerauth.New(&authServiceMock{}, nil)

		req := httptest.NewRequest(http.MethodPost, "/auth/register", bytes.NewBufferString("{"))
		rr := httptest.NewRecorder()

		h.Register(rr, req)

		if rr.Code != http.StatusBadRequest {
			t.Fatalf("status = %d, want %d", rr.Code, http.StatusBadRequest)
		}
	})

	t.Run("email is required", func(t *testing.T) {
		h := handlerauth.New(&authServiceMock{}, nil)

		body := `{"password":"123456","role":"user"}`
		req := httptest.NewRequest(http.MethodPost, "/auth/register", bytes.NewBufferString(body))
		rr := httptest.NewRecorder()

		h.Register(rr, req)

		if rr.Code != http.StatusBadRequest {
			t.Fatalf("status = %d, want %d", rr.Code, http.StatusBadRequest)
		}
	})

	t.Run("password is required", func(t *testing.T) {
		h := handlerauth.New(&authServiceMock{}, nil)

		body := `{"email":"test@example.com","role":"user"}`
		req := httptest.NewRequest(http.MethodPost, "/auth/register", bytes.NewBufferString(body))
		rr := httptest.NewRecorder()

		h.Register(rr, req)

		if rr.Code != http.StatusBadRequest {
			t.Fatalf("status = %d, want %d", rr.Code, http.StatusBadRequest)
		}
	})

	t.Run("invalid role", func(t *testing.T) {
		h := handlerauth.New(&authServiceMock{}, nil)

		body := `{"email":"test@example.com","password":"123456","role":"guest"}`
		req := httptest.NewRequest(http.MethodPost, "/auth/register", bytes.NewBufferString(body))
		rr := httptest.NewRecorder()

		h.Register(rr, req)

		if rr.Code != http.StatusBadRequest {
			t.Fatalf("status = %d, want %d", rr.Code, http.StatusBadRequest)
		}
	})

	t.Run("email already exists", func(t *testing.T) {
		h := handlerauth.New(&authServiceMock{
			registerFunc: func(ctx context.Context, email, password string, role user.Role) (user.User, error) {
				return user.User{}, user.ErrUnicEmail
			},
		}, nil)

		body := `{"email":"test@example.com","password":"123456","role":"user"}`
		req := httptest.NewRequest(http.MethodPost, "/auth/register", bytes.NewBufferString(body))
		rr := httptest.NewRecorder()

		h.Register(rr, req)

		if rr.Code != http.StatusBadRequest {
			t.Fatalf("status = %d, want %d", rr.Code, http.StatusBadRequest)
		}
	})

	t.Run("success", func(t *testing.T) {
		createdAt := time.Now().UTC()
		expectedID := uuid.New()
		h := handlerauth.New(&authServiceMock{
			registerFunc: func(ctx context.Context, email, password string, role user.Role) (user.User, error) {
				if email != "test@example.com" {
					t.Fatalf("email = %q, want %q", email, "test@example.com")
				}
				if password != "123456" {
					t.Fatalf("password = %q, want %q", password, "123456")
				}
				if role != user.RoleUser {
					t.Fatalf("role = %q, want %q", role, user.RoleUser)
				}
				return user.User{
					ID:        expectedID,
					Email:     email,
					Role:      role,
					CreatedAt: &createdAt,
				}, nil
			},
		}, nil)

		body := `{"email":"test@example.com","password":"123456","role":"user"}`
		req := httptest.NewRequest(http.MethodPost, "/auth/register", bytes.NewBufferString(body))
		rr := httptest.NewRecorder()

		h.Register(rr, req)

		if rr.Code != http.StatusCreated {
			t.Fatalf("status = %d, want %d", rr.Code, http.StatusCreated)
		}

		var resp handlerauth.RegisterResponse
		if err := json.Unmarshal(rr.Body.Bytes(), &resp); err != nil {
			t.Fatalf("unmarshal response: %v", err)
		}
		if resp.User.ID != expectedID {
			t.Fatalf("user id = %v, want %v", resp.User.ID, expectedID)
		}
		if resp.User.Email != "test@example.com" {
			t.Fatalf("user email = %q, want %q", resp.User.Email, "test@example.com")
		}
	})
}

func TestAuthHandler_Login(t *testing.T) {
	t.Run("invalid json", func(t *testing.T) {
		h := handlerauth.New(&authServiceMock{}, nil)

		req := httptest.NewRequest(http.MethodPost, "/auth/login", bytes.NewBufferString("{"))
		rr := httptest.NewRecorder()

		h.Login(rr, req)

		if rr.Code != http.StatusBadRequest {
			t.Fatalf("status = %d, want %d", rr.Code, http.StatusBadRequest)
		}
	})

	t.Run("email is required", func(t *testing.T) {
		h := handlerauth.New(&authServiceMock{}, nil)

		body := `{"password":"123456"}`
		req := httptest.NewRequest(http.MethodPost, "/auth/login", bytes.NewBufferString(body))
		rr := httptest.NewRecorder()

		h.Login(rr, req)

		if rr.Code != http.StatusBadRequest {
			t.Fatalf("status = %d, want %d", rr.Code, http.StatusBadRequest)
		}
	})

	t.Run("password is required", func(t *testing.T) {
		h := handlerauth.New(&authServiceMock{}, nil)

		body := `{"email":"test@example.com"}`
		req := httptest.NewRequest(http.MethodPost, "/auth/login", bytes.NewBufferString(body))
		rr := httptest.NewRecorder()

		h.Login(rr, req)

		if rr.Code != http.StatusBadRequest {
			t.Fatalf("status = %d, want %d", rr.Code, http.StatusBadRequest)
		}
	})

	t.Run("invalid credentials", func(t *testing.T) {
		h := handlerauth.New(&authServiceMock{
			loginFunc: func(ctx context.Context, email, password string) (string, error) {
				return "", user.ErrIncorrectPassword
			},
		}, nil)

		body := `{"email":"test@example.com","password":"123456"}`
		req := httptest.NewRequest(http.MethodPost, "/auth/login", bytes.NewBufferString(body))
		rr := httptest.NewRecorder()

		h.Login(rr, req)

		if rr.Code != http.StatusUnauthorized {
			t.Fatalf("status = %d, want %d", rr.Code, http.StatusUnauthorized)
		}

		var resp errorResponse
		if err := json.Unmarshal(rr.Body.Bytes(), &resp); err != nil {
			t.Fatalf("unmarshal response: %v", err)
		}
		if resp.Error.Code != "UNAUTHORIZED" {
			t.Fatalf("error code = %q, want %q", resp.Error.Code, "UNAUTHORIZED")
		}
	})

	t.Run("success", func(t *testing.T) {
		h := handlerauth.New(&authServiceMock{
			loginFunc: func(ctx context.Context, email, password string) (string, error) {
				if email != "test@example.com" {
					t.Fatalf("email = %q, want %q", email, "test@example.com")
				}
				if password != "123456" {
					t.Fatalf("password = %q, want %q", password, "123456")
				}
				return "token-123", nil
			},
		}, nil)

		body := `{"email":"test@example.com","password":"123456"}`
		req := httptest.NewRequest(http.MethodPost, "/auth/login", bytes.NewBufferString(body))
		rr := httptest.NewRecorder()

		h.Login(rr, req)

		if rr.Code != http.StatusOK {
			t.Fatalf("status = %d, want %d", rr.Code, http.StatusOK)
		}

		var resp handlerauth.TokenResponse
		if err := json.Unmarshal(rr.Body.Bytes(), &resp); err != nil {
			t.Fatalf("unmarshal response: %v", err)
		}
		if resp.Token != "token-123" {
			t.Fatalf("token = %q, want %q", resp.Token, "token-123")
		}
	})
}

func TestAuthHandler_DummyLogin(t *testing.T) {
	t.Run("invalid json", func(t *testing.T) {
		h := handlerauth.New(&authServiceMock{}, nil)

		req := httptest.NewRequest(http.MethodPost, "/auth/dummy-login", bytes.NewBufferString("{"))
		rr := httptest.NewRecorder()

		h.DummyLogin(rr, req)

		if rr.Code != http.StatusBadRequest {
			t.Fatalf("status = %d, want %d", rr.Code, http.StatusBadRequest)
		}
	})

	t.Run("invalid role", func(t *testing.T) {
		h := handlerauth.New(&authServiceMock{}, nil)

		body := `{"role":"guest"}`
		req := httptest.NewRequest(http.MethodPost, "/auth/dummy-login", bytes.NewBufferString(body))
		rr := httptest.NewRecorder()

		h.DummyLogin(rr, req)

		if rr.Code != http.StatusBadRequest {
			t.Fatalf("status = %d, want %d", rr.Code, http.StatusBadRequest)
		}
	})

	t.Run("success", func(t *testing.T) {
		h := handlerauth.New(&authServiceMock{
			generateDummyTokenFunc: func(role user.Role) (string, error) {
				if role != user.RoleAdmin {
					t.Fatalf("role = %q, want %q", role, user.RoleAdmin)
				}
				return "dummy-token", nil
			},
		}, nil)

		body := `{"role":"admin"}`
		req := httptest.NewRequest(http.MethodPost, "/auth/dummy-login", bytes.NewBufferString(body))
		rr := httptest.NewRecorder()

		h.DummyLogin(rr, req)

		if rr.Code != http.StatusOK {
			t.Fatalf("status = %d, want %d", rr.Code, http.StatusOK)
		}

		var resp handlerauth.TokenResponse
		if err := json.Unmarshal(rr.Body.Bytes(), &resp); err != nil {
			t.Fatalf("unmarshal response: %v", err)
		}
		if resp.Token != "dummy-token" {
			t.Fatalf("token = %q, want %q", resp.Token, "dummy-token")
		}
	})
}
