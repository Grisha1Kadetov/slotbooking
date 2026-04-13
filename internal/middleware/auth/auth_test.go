package auth_test

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	middlewareauth "github.com/Grisha1Kadetov/slotbooking/internal/middleware/auth"
	authmodel "github.com/Grisha1Kadetov/slotbooking/internal/model/auth"
	"github.com/Grisha1Kadetov/slotbooking/internal/model/user"
	"github.com/google/uuid"
)

type serviceMock struct {
	parseTokenFunc func(ctx context.Context, tokenString string) (uuid.UUID, user.Role, error)
}

func (m *serviceMock) ParseToken(ctx context.Context, tokenString string) (uuid.UUID, user.Role, error) {
	if m.parseTokenFunc == nil {
		return uuid.Nil, "", errors.New("unexpected ParseToken call")
	}
	return m.parseTokenFunc(ctx, tokenString)
}

type errorResponse struct {
	Error struct {
		Code    string `json:"code"`
		Message string `json:"message"`
	} `json:"error"`
}

func TestAuthChecker_Middleware(t *testing.T) {
	t.Run("missing authorization header", func(t *testing.T) {
		checker := middlewareauth.New(&serviceMock{}, user.RoleUser)

		h := checker.Middleware()(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			t.Fatal("next handler must not be called")
		}))

		req := httptest.NewRequest(http.MethodGet, "/", nil)
		rr := httptest.NewRecorder()

		h.ServeHTTP(rr, req)

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

	t.Run("invalid authorization header prefix", func(t *testing.T) {
		checker := middlewareauth.New(&serviceMock{}, user.RoleUser)

		h := checker.Middleware()(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			t.Fatal("next handler must not be called")
		}))

		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.Header.Set("Authorization", "Token abc")
		rr := httptest.NewRecorder()

		h.ServeHTTP(rr, req)

		if rr.Code != http.StatusUnauthorized {
			t.Fatalf("status = %d, want %d", rr.Code, http.StatusUnauthorized)
		}
	})

	t.Run("invalid token", func(t *testing.T) {
		checker := middlewareauth.New(&serviceMock{
			parseTokenFunc: func(ctx context.Context, tokenString string) (uuid.UUID, user.Role, error) {
				if tokenString != "bad-token" {
					t.Fatalf("tokenString = %q, want %q", tokenString, "bad-token")
				}
				return uuid.Nil, "", errors.New("invalid token")
			},
		}, user.RoleUser)

		h := checker.Middleware()(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			t.Fatal("next handler must not be called")
		}))

		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.Header.Set("Authorization", "Bearer bad-token")
		rr := httptest.NewRecorder()

		h.ServeHTTP(rr, req)

		if rr.Code != http.StatusUnauthorized {
			t.Fatalf("status = %d, want %d", rr.Code, http.StatusUnauthorized)
		}
	})

	t.Run("forbidden role", func(t *testing.T) {
		checker := middlewareauth.New(&serviceMock{
			parseTokenFunc: func(ctx context.Context, tokenString string) (uuid.UUID, user.Role, error) {
				return uuid.New(), user.RoleUser, nil
			},
		}, user.RoleAdmin)

		h := checker.Middleware()(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			t.Fatal("next handler must not be called")
		}))

		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.Header.Set("Authorization", "Bearer valid-token")
		rr := httptest.NewRecorder()

		h.ServeHTTP(rr, req)

		if rr.Code != http.StatusForbidden {
			t.Fatalf("status = %d, want %d", rr.Code, http.StatusForbidden)
		}
	})

	t.Run("success", func(t *testing.T) {
		expectedUserID := uuid.New()
		checker := middlewareauth.New(&serviceMock{
			parseTokenFunc: func(ctx context.Context, tokenString string) (uuid.UUID, user.Role, error) {
				if tokenString != "valid-token" {
					t.Fatalf("tokenString = %q, want %q", tokenString, "valid-token")
				}
				return expectedUserID, user.RoleUser, nil
			},
		}, user.RoleUser)

		called := false
		h := checker.Middleware()(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			called = true

			actor, ok := authmodel.ActorFromContext(r.Context())
			if !ok {
				t.Fatal("actor not found in context")
			}
			if actor.UserId != expectedUserID {
				t.Fatalf("actor.UserId = %v, want %v", actor.UserId, expectedUserID)
			}
			if actor.Role != user.RoleUser {
				t.Fatalf("actor.Role = %q, want %q", actor.Role, user.RoleUser)
			}

			w.WriteHeader(http.StatusOK)
		}))

		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.Header.Set("Authorization", "Bearer valid-token")
		rr := httptest.NewRecorder()

		h.ServeHTTP(rr, req)

		if !called {
			t.Fatal("next handler was not called")
		}
		if rr.Code != http.StatusOK {
			t.Fatalf("status = %d, want %d", rr.Code, http.StatusOK)
		}
	})
}
