package auth

import (
	"context"
	"errors"
	"time"

	"github.com/Grisha1Kadetov/slotbooking/internal/model/user"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

var (
	dummyAdminID = uuid.MustParse("11111111-1111-1111-1111-111111111111")
	dummyUserID  = uuid.MustParse("22222222-2222-2222-2222-222222222222")
)

type AuthService struct {
	userRepo  userRepository
	secretKey []byte
}

type claims struct {
	UserID uuid.UUID `json:"user_id"`
	Role   user.Role `json:"role"`
	jwt.RegisteredClaims
}

func New(secretKey []byte, userRepo userRepository) *AuthService {
	return &AuthService{
		userRepo:  userRepo,
		secretKey: secretKey,
	}
}

func (s *AuthService) Register(ctx context.Context, email, password string, role user.Role) (user.User, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return user.User{}, err
	}
	u := user.User{
		Email:        email,
		PasswordHash: string(hash),
		Role:         role,
	}
	return s.userRepo.Create(ctx, u)
}

func (s *AuthService) Login(ctx context.Context, email, password string) (string, error) {
	u, err := s.userRepo.GetByEmail(ctx, email)
	if err != nil {
		return "", err
	}

	if bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte(password)) != nil {
		return "", user.ErrIncorrectPassword
	}

	tokenString, err := generateToken(u.ID, u.Role, s.secretKey)
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

func (s *AuthService) ParseToken(ctx context.Context, tokenString string) (uuid.UUID, user.Role, error) {
	claims := &claims{}
	_, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		return s.secretKey, nil
	})
	if err != nil {
		return uuid.Nil, "", err
	}

	return claims.UserID, claims.Role, nil
}

func (s *AuthService) GenerateDummyToken(role user.Role) (string, error) {
	var userID uuid.UUID
	switch role {
	case user.RoleAdmin:
		userID = dummyAdminID
	case user.RoleUser:
		userID = dummyUserID
	default:
		return "", errors.New("invalid role")
	}

	tokenString, err := generateToken(userID, role, s.secretKey)
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

func generateToken(uuid uuid.UUID, role user.Role, secretKey []byte) (string, error) {
	claims := claims{
		UserID: uuid,
		Role:   role,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   uuid.String(),
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(secretKey)
}
