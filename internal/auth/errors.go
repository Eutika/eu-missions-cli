package auth

import "fmt"

// AuthenticationError represents specific authentication-related errors.
type AuthenticationError struct {
	Code    string
	Message string
	Err     error
}

func (e *AuthenticationError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("%s: %s (underlying error: %v)", e.Code, e.Message, e.Err)
	}
	return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

// Error codes for authentication.
const (
	ErrDeviceCodeRequest = "DEVICE_CODE_REQUEST_ERROR"
	ErrAuthPolling       = "ERROR_AUTH_POLLING"
	ErrSaveToken         = "SAVE_TOKEN_ERROR"
	ErrTokenRefresh      = "TOKEN_REFRESH_ERROR" // #nosec G101
)

// NewDeviceCodeError creates a new authentication error for device code request failures.
func NewDeviceCodeError(err error) *AuthenticationError {
	return &AuthenticationError{
		Code:    ErrDeviceCodeRequest,
		Message: "Failed to request device code",
		Err:     err,
	}
}

// NewTokenPollingError creates a new authentication error for token polling failures.
func NewTokenPollingError(err error) *AuthenticationError {
	return &AuthenticationError{
		Code:    ErrAuthPolling,
		Message: "Failed to poll for authentication token",
		Err:     err,
	}
}

// NewTokenSavingError creates a new authentication error for token saving failures.
func NewTokenSavingError(err error) *AuthenticationError {
	return &AuthenticationError{
		Code:    ErrSaveToken,
		Message: "Failed to save authentication tokens",
		Err:     err,
	}
}

// NewTokenRefreshError creates a new authentication error for token refresh failures.
func NewTokenRefreshError(err error) *AuthenticationError {
	return &AuthenticationError{
		Code:    ErrTokenRefresh,
		Message: "Failed to refresh authentication token",
		Err:     err,
	}
}
