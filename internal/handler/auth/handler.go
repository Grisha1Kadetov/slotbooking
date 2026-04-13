package auth

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/Grisha1Kadetov/slotbooking/internal/model/user"
	errrend "github.com/Grisha1Kadetov/slotbooking/internal/pkg/errorrenderer"
	"github.com/Grisha1Kadetov/slotbooking/internal/pkg/log"
	"github.com/go-chi/render"
)

type AuthHandler struct {
	service authService
	l       log.Logger
}

func New(s authService, l log.Logger) *AuthHandler {
	return &AuthHandler{service: s, l: l}
}

func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	var request RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		errrend.RenderBadRequestError(w, r, "invalid JSON")
		return
	}

	if request.Email == "" {
		errrend.RenderBadRequestError(w, r, "email is required")
		return
	}
	if request.Password == "" {
		errrend.RenderBadRequestError(w, r, "password is required")
		return
	}
	if request.Role != user.RoleAdmin && request.Role != user.RoleUser {
		errrend.RenderBadRequestError(w, r, "role must be admin or user")
		return
	}

	u, err := h.service.Register(r.Context(), request.Email, request.Password, request.Role)
	if err != nil {
		if errors.Is(err, user.ErrUnicEmail) {
			errrend.RenderBadRequestError(w, r, "email already exists")
			return
		}
		errrend.RenderInternalError(w, r)
		h.l.Error("failed to register user", log.Err(err))
		return
	}

	response := ToRegisterResponse(u)
	render.Status(r, http.StatusCreated)
	render.JSON(w, r, response)
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var request LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		errrend.RenderBadRequestError(w, r, "invalid JSON")
		return
	}

	if request.Email == "" {
		errrend.RenderBadRequestError(w, r, "email is required")
		return
	}
	if request.Password == "" {
		errrend.RenderBadRequestError(w, r, "password is required")
		return
	}

	token, err := h.service.Login(r.Context(), request.Email, request.Password)
	if err != nil {
		if errors.Is(err, user.ErrNotFound) || errors.Is(err, user.ErrIncorrectPassword) {
			errrend.RenderError(w, r, http.StatusUnauthorized, string(errrend.StatusUnauthorized), "invalid email or password")
			return
		}
		errrend.RenderInternalError(w, r)
		h.l.Error("failed to login user", log.Err(err))
		return
	}

	response := ToTokenResponse(token)
	render.JSON(w, r, response)
}

func (h *AuthHandler) DummyLogin(w http.ResponseWriter, r *http.Request) {
	var request DummyLoginRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		errrend.RenderBadRequestError(w, r, "invalid JSON")
		return
	}

	if request.Role != user.RoleAdmin && request.Role != user.RoleUser {
		errrend.RenderBadRequestError(w, r, "invalid role value")
		return
	}

	token, err := h.service.GenerateDummyToken(request.Role)
	if err != nil {
		errrend.RenderInternalError(w, r)
		h.l.Error("failed to generate dummy token", log.Err(err))
		return
	}

	response := ToTokenResponse(token)
	render.JSON(w, r, response)
}
