package auth

import (
	"context"
	"errors"
	"net/http"
	"strings"

	"github.com/coreos/go-oidc/v3/oidc"
	"github.com/google/uuid"
	"github.com/mdflamingo/fgo-tracker-backend/internal/apiresponse"
	"github.com/mdflamingo/fgo-tracker-backend/internal/logger"
	"github.com/mdflamingo/fgo-tracker-backend/internal/model"
	"go.uber.org/zap"
)

type contextKey string

const identityKey contextKey = "identity"

type OIDCValidator struct {
	verifier *oidc.IDTokenVerifier
}

func NewOIDCValidator(providerURL, clientID string) (*OIDCValidator, error) {
	provider, err := oidc.NewProvider(context.Background(), providerURL)
	if err != nil {
		return nil, err
	}

	config := &oidc.Config{
		ClientID: clientID,
	}

	return &OIDCValidator{
		verifier: provider.Verifier(config),
	}, nil
}

func (v *OIDCValidator) AuthMiddleware() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				logger.Log.Debug("no authorization header found")
				apiresponse.ResponseWithError(w, r, http.StatusUnauthorized, "Unauthorized: missing authorization header")
				return
			}

			tokenString := strings.TrimPrefix(authHeader, "Bearer ")

			identity, err := v.validateJWT(tokenString)
			if err != nil {
				logger.Log.Warn("invalid token", zap.Error(err))
				apiresponse.ResponseWithError(w, r, http.StatusUnauthorized, "Unauthorized: invalid token")
				return
			}

			ctx := context.WithValue(r.Context(), identityKey, identity)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func (v *OIDCValidator) validateJWT(tokenString string) (model.Identity, error) {
	idToken, err := v.verifier.Verify(context.Background(), tokenString)
	if err != nil {
		return model.Identity{}, err
	}

	var claims struct {
		Subject           string `json:"sub"`
		PreferredUsername string `json:"preferred_username"`
		Email             string `json:"email"`
	}

	if err := idToken.Claims(&claims); err != nil {
		return model.Identity{}, errors.New("failed to parse claims")
	}

	if claims.Subject == "" {
		return model.Identity{}, errors.New("subject (sub) not found in token")
	}

	id, err := uuid.Parse(claims.Subject)
	if err != nil {
		return model.Identity{}, errors.New("subject (sub) is not a valid UUID")
	}

	username := claims.PreferredUsername
	if username == "" {
		username = claims.Subject
	}

	var email *string
	if claims.Email != "" {
		email = &claims.Email
	}

	return model.Identity{
		ID:       id,
		Username: username,
		Email:    email,
	}, nil
}

// IdentityFromRequest возвращает атрибуты владельца токена из контекста запроса.
func IdentityFromRequest(r *http.Request) (model.Identity, error) {
	identityValue := r.Context().Value(identityKey)
	if identityValue == nil {
		return model.Identity{}, errors.New("identity not found in context")
	}

	identity, ok := identityValue.(model.Identity)
	if !ok {
		return model.Identity{}, errors.New("identity has unexpected type")
	}

	return identity, nil
}

func GetUserIDFromRequest(r *http.Request) (string, error) {
	identity, err := IdentityFromRequest(r)
	if err != nil {
		return "", err
	}

	return identity.ID.String(), nil
}
