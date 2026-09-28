package auth

import (
	"context"
	"errors"
	"net/http"
	"strings"

	"github.com/coreos/go-oidc/v3/oidc"
	"github.com/mdflamingo/fgo-tracker-backend/internal/logger"
	"go.uber.org/zap"
)

type contextKey string

const userIDKey contextKey = "userID"

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
				http.Error(w, "Unauthorized: missing authorization header", http.StatusUnauthorized)
				return
			}

			tokenString := strings.TrimPrefix(authHeader, "Bearer ")

			userID, err := v.validateJWT(tokenString)
			if err != nil {
				logger.Log.Warn("invalid token", zap.Error(err))
				http.Error(w, "Unauthorized: invalid token", http.StatusUnauthorized)
				return
			}

			ctx := context.WithValue(r.Context(), userIDKey, userID)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func (v *OIDCValidator) validateJWT(tokenString string) (string, error) {
	idToken, err := v.verifier.Verify(context.Background(), tokenString)
	if err != nil {
		return "", err
	}

	var claims struct {
		Subject string `json:"sub"`
	}

	if err := idToken.Claims(&claims); err != nil {
		return "", errors.New("failed to parse claims")
	}

	if claims.Subject == "" {
		return "", errors.New("subject (sub) not found in token")
	}

	return claims.Subject, nil
}

func GetUserIDFromRequest(r *http.Request) (string, error) {
	ctx := r.Context()
	userIDValue := ctx.Value(userIDKey)
	if userIDValue == nil {
		return "", errors.New("userID not found in context")
	}

	userID, ok := userIDValue.(string)
	if !ok {
		return "", errors.New("userID is not a string")
	}

	return userID, nil
}
