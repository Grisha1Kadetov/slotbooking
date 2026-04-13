package auth

import (
	"net/http"
	"slices"
	"strings"

	"github.com/Grisha1Kadetov/slotbooking/internal/model/auth"
	"github.com/Grisha1Kadetov/slotbooking/internal/model/user"
	"github.com/Grisha1Kadetov/slotbooking/internal/pkg/errorrenderer"
)

type AuthChecker struct {
	service service
	roles   []user.Role
}

func New(service service, roles ...user.Role) *AuthChecker {
	return &AuthChecker{
		service: service,
		roles:   roles,
	}
}

func (ac *AuthChecker) Middleware() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				errorrenderer.RenderError(w, r, http.StatusUnauthorized, string(errorrenderer.StatusUnauthorized), "authorization header is required")
				return
			}

			const prefix = "Bearer "
			if !strings.HasPrefix(authHeader, prefix) {
				errorrenderer.RenderError(w, r, http.StatusUnauthorized, string(errorrenderer.StatusUnauthorized), "invalid authorization header must start with Bearer")
				return
			}
			tokenString := strings.TrimPrefix(authHeader, prefix)

			userID, role, err := ac.service.ParseToken(r.Context(), tokenString)
			if err != nil {
				errorrenderer.RenderError(w, r, http.StatusUnauthorized, string(errorrenderer.StatusUnauthorized), "invalid authorization token")
				return
			}

			if !slices.Contains(ac.roles, role) {
				errorrenderer.RenderError(w, r, http.StatusForbidden, string(errorrenderer.StatusForbidden), "forbidden")
				return
			}

			ctx := auth.WithActor(r.Context(), auth.Actor{
				UserId: userID,
				Role:   role,
			})

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
