package auth_test

import (
	"context"
	"errors"
	"testing"

	"github.com/Grisha1Kadetov/slotbooking/internal/model/user"
	"github.com/Grisha1Kadetov/slotbooking/internal/service/auth"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

type userRepositoryMock struct {
	createFunc     func(ctx context.Context, u user.User) (user.User, error)
	getByEmailFunc func(ctx context.Context, email string) (user.User, error)
}

func (m *userRepositoryMock) Create(ctx context.Context, u user.User) (user.User, error) {
	if m.createFunc == nil {
		return user.User{}, errors.New("unexpected Create call")
	}
	return m.createFunc(ctx, u)
}

func (m *userRepositoryMock) GetByEmail(ctx context.Context, email string) (user.User, error) {
	if m.getByEmailFunc == nil {
		return user.User{}, errors.New("unexpected GetByEmail call")
	}
	return m.getByEmailFunc(ctx, email)
}

func TestAuthService_Register(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	var createdUser user.User

	repo := &userRepositoryMock{
		createFunc: func(ctx context.Context, u user.User) (user.User, error) {
			createdUser = u
			u.ID = uuid.New()
			return u, nil
		},
	}

	service := auth.New([]byte("secret"), repo)

	got, err := service.Register(ctx, "test@example.com", "qwerty123", user.RoleUser)
	if err != nil {
		t.Fatalf("Register() error = %v", err)
	}

	if got.Email != "test@example.com" {
		t.Fatalf("Register() email = %q, want %q", got.Email, "test@example.com")
	}
	if got.Role != user.RoleUser {
		t.Fatalf("Register() role = %q, want %q", got.Role, user.RoleUser)
	}
	if got.PasswordHash == "" {
		t.Fatal("Register() password hash is empty")
	}
	if createdUser.PasswordHash == "qwerty123" {
		t.Fatal("Register() password was not hashed")
	}
	if err := bcrypt.CompareHashAndPassword([]byte(createdUser.PasswordHash), []byte("qwerty123")); err != nil {
		t.Fatalf("Register() saved invalid bcrypt hash: %v", err)
	}
}

func TestAuthService_Login_Success(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	password := "qwerty123"
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		t.Fatalf("GenerateFromPassword() error = %v", err)
	}

	expectedID := uuid.New()

	repo := &userRepositoryMock{
		getByEmailFunc: func(ctx context.Context, email string) (user.User, error) {
			if email != "test@example.com" {
				t.Fatalf("GetByEmail() email = %q, want %q", email, "test@example.com")
			}
			return user.User{
				ID:           expectedID,
				Email:        email,
				PasswordHash: string(hash),
				Role:         user.RoleUser,
			}, nil
		},
	}

	service := auth.New([]byte("secret"), repo)

	token, err := service.Login(ctx, "test@example.com", password)
	if err != nil {
		t.Fatalf("Login() error = %v", err)
	}
	if token == "" {
		t.Fatal("Login() returned empty token")
	}
}

func TestAuthService_Login_UserNotFound(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	repo := &userRepositoryMock{
		getByEmailFunc: func(ctx context.Context, email string) (user.User, error) {
			return user.User{}, user.ErrNotFound
		},
	}

	service := auth.New([]byte("secret"), repo)

	_, err := service.Login(ctx, "test@example.com", "qwerty123")
	if !errors.Is(err, user.ErrNotFound) {
		t.Fatalf("Login() error = %v, want %v", err, user.ErrNotFound)
	}
}

func TestAuthService_Login_IncorrectPassword(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	hash, err := bcrypt.GenerateFromPassword([]byte("correct-password"), bcrypt.DefaultCost)
	if err != nil {
		t.Fatalf("GenerateFromPassword() error = %v", err)
	}

	repo := &userRepositoryMock{
		getByEmailFunc: func(ctx context.Context, email string) (user.User, error) {
			return user.User{
				ID:           uuid.New(),
				Email:        email,
				PasswordHash: string(hash),
				Role:         user.RoleUser,
			}, nil
		},
	}

	service := auth.New([]byte("secret"), repo)

	_, err = service.Login(ctx, "test@example.com", "wrong-password")
	if !errors.Is(err, user.ErrIncorrectPassword) {
		t.Fatalf("Login() error = %v, want %v", err, user.ErrIncorrectPassword)
	}
}

func TestAuthService_ParseToken(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	secret := []byte("secret")
	service := auth.New(secret, &userRepositoryMock{})

	token, err := service.GenerateDummyToken(user.RoleAdmin)
	if err != nil {
		t.Fatalf("GenerateDummyToken() error = %v", err)
	}

	gotID, gotRole, err := service.ParseToken(ctx, token)
	if err != nil {
		t.Fatalf("ParseToken() error = %v", err)
	}

	expectedID := uuid.MustParse("11111111-1111-1111-1111-111111111111")
	if gotID != expectedID {
		t.Fatalf("ParseToken() userID = %v, want %v", gotID, expectedID)
	}
	if gotRole != user.RoleAdmin {
		t.Fatalf("ParseToken() role = %q, want %q", gotRole, user.RoleAdmin)
	}
}

func TestAuthService_ParseToken_InvalidToken(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	service := auth.New([]byte("secret"), &userRepositoryMock{})

	_, _, err := service.ParseToken(ctx, "invalid-token")
	if err == nil {
		t.Fatal("ParseToken() expected error, got nil")
	}
}

func TestAuthService_GenerateDummyToken(t *testing.T) {
	t.Parallel()

	secret := []byte("secret")
	service := auth.New(secret, &userRepositoryMock{})

	tests := []struct {
		name       string
		role       user.Role
		expectedID uuid.UUID
		wantErr    bool
	}{
		{
			name:       "admin",
			role:       user.RoleAdmin,
			expectedID: uuid.MustParse("11111111-1111-1111-1111-111111111111"),
		},
		{
			name:       "user",
			role:       user.RoleUser,
			expectedID: uuid.MustParse("22222222-2222-2222-2222-222222222222"),
		},
		{
			name:    "invalid role",
			role:    user.Role("guest"),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			token, err := service.GenerateDummyToken(tt.role)
			if tt.wantErr {
				if err == nil {
					t.Fatal("GenerateDummyToken() expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Fatalf("GenerateDummyToken() error = %v", err)
			}
			if token == "" {
				t.Fatal("GenerateDummyToken() returned empty token")
			}

			gotID, gotRole, err := service.ParseToken(context.Background(), token)
			if err != nil {
				t.Fatalf("ParseToken() error = %v", err)
			}

			if gotID != tt.expectedID {
				t.Fatalf("dummy token userID = %v, want %v", gotID, tt.expectedID)
			}
			if gotRole != tt.role {
				t.Fatalf("dummy token role = %q, want %q", gotRole, tt.role)
			}
		})
	}
}
