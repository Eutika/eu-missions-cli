package auth

import (
	"errors"
	"fmt"
	"time"

	"github.com/zalando/go-keyring"

	"github.com/eutika/eu-missions-cli/internal/config"
)

type TokenResponse struct {
	AccessToken  string `json:"access_token"`
	TokenType    string `json:"token_type"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int    `json:"expires_in"`
}

func SaveTokens(token *TokenResponse) error {
	// Save access token.
	if err := keyring.Set(config.NewConfig().GetKeyringService(), "access_token", token.AccessToken); err != nil {
		return NewTokenSavingError(fmt.Errorf("error saving access token: %w", err))
	}

	// Save refresh token.
	if err := keyring.Set(config.NewConfig().GetKeyringService(), "refresh_token", token.RefreshToken); err != nil {
		if delErr := keyring.Delete(config.NewConfig().GetKeyringService(), "access_token"); delErr != nil {
			return NewTokenSavingError(fmt.Errorf("error deleting access token: %w", delErr))
		}
		return NewTokenSavingError(fmt.Errorf("error saving refresh token: %w", err))
	}

	// Save token expiration.
	expiresAt := time.Now().Add(time.Duration(token.ExpiresIn) * time.Second)
	if err := keyring.Set(config.NewConfig().GetKeyringService(),
		"token_expires_at", expiresAt.Format(time.RFC3339)); err != nil {
		if delErr := keyring.Delete(config.NewConfig().GetKeyringService(), "access_token"); delErr != nil {
			return NewTokenSavingError(fmt.Errorf("error deleting access token: %w", delErr))
		}
		if delErr := keyring.Delete(config.NewConfig().GetKeyringService(), "refresh_token"); delErr != nil {
			return NewTokenSavingError(fmt.Errorf("error deleting refresh token: %w", delErr))
		}
		return NewTokenSavingError(fmt.Errorf("error saving token expiration: %w", err))
	}

	return nil
}

const tokenExpiryBufferMinutes = 5

func IsTokenExpired() (bool, error) {
	expiresAtStr, err := keyring.Get(config.NewConfig().GetKeyringService(), "token_expires_at")
	if err != nil {
		return true, err // Si no podemos obtener la fecha, asumimos que expiró
	}

	expiresAt, err := time.Parse(time.RFC3339, expiresAtStr)
	if err != nil {
		return true, err // Si no podemos parsear la fecha, asumimos que expiró
	}

	// Considerar el token expirado tokenExpiryBufferMinutes minutos antes de su expiración real
	return time.Now().Add(tokenExpiryBufferMinutes * time.Minute).After(expiresAt), nil
}

func GetCurrentToken() (string, error) {
	expired, err := IsTokenExpired()
	if err != nil {
		return "", err
	}

	if expired {
		if refreshErr := RefreshToken(); refreshErr != nil {
			return "", err
		}
	}

	return keyring.Get(config.NewConfig().GetKeyringService(), "access_token")
}

func RefreshToken() error {
	_, err := keyring.Get(config.NewConfig().GetKeyringService(), "refresh_token")
	if err != nil {
		return NewTokenRefreshError(errors.New("no refresh token found, please login again"))
	}

	// TODO: Implement refresh token login. For now, we force a new login.
	return NewTokenRefreshError(errors.New("session expired, please login again"))
}
