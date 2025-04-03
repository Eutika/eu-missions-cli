package auth

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"github.com/eutika/eu-missions-cli/internal/config"
)

func PollForToken(deviceCode *DeviceCodeResponse) (*TokenResponse, error) {
	payload := map[string]string{
		"client_id":   config.NewConfig().GetClientID(),
		"device_code": deviceCode.DeviceCode,
		"grant_type":  "urn:ietf:params:oauth:grant-type:device_code",
	}

	deadline := time.Now().Add(time.Duration(deviceCode.ExpiresIn) * time.Second)
	interval := time.Duration(deviceCode.Interval) * time.Second

	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		if time.Now().After(deadline) {
			return nil, NewTokenPollingError(errors.New("tiempo de espera por la autenticación agotado"))
		}
		<-ticker.C
		jsonData, err := json.Marshal(payload)
		if err != nil {
			return nil, NewTokenPollingError(errors.New("error marshaling token request"))
		}

		ctx := context.Background()
		req, err := http.NewRequestWithContext(ctx, http.MethodPost,
			config.NewConfig().GetTokenURL(), bytes.NewBuffer(jsonData))
		if err != nil {
			return nil, NewTokenPollingError(errors.New("error creating token request"))
		}
		req.Header.Set("Content-Type", "application/json")

		client := &http.Client{}
		resp, err := client.Do(req)
		if err != nil {
			return nil, NewTokenPollingError(errors.New("error requesting token"))
		}
		defer resp.Body.Close()

		var errorResponse struct {
			Error string `json:"error"`
		}

		switch resp.StatusCode {
		case http.StatusOK:
			var token TokenResponse
			if decodeErr := json.NewDecoder(resp.Body).Decode(&token); decodeErr != nil {
				return nil, NewTokenPollingError(errors.New("error decoding token response"))
			}
			return &token, nil

		case http.StatusBadRequest:
			if decodeErr := json.NewDecoder(resp.Body).Decode(&errorResponse); decodeErr != nil {
				return nil, NewTokenPollingError(errors.New("error decoding error response"))
			}

			switch errorResponse.Error {
			case "authorization_pending":
				// Continue poolling.
				continue
			case "slow_down":
				// Increase interval.
				interval *= 2
				ticker.Reset(interval)
				continue
			case "expired_token":
				return nil, NewTokenPollingError(errors.New("el código de autorización ha expirado"))
			default:
				return nil, NewTokenPollingError(errors.New("error inesperado recuperando el token"))
			}

		default:
			return nil, NewTokenPollingError(errors.New("respuesta inesperada del servidor"))
		}
	}
}
