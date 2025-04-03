package auth

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/eutika/eu-missions-cli/internal/config"
)

type DeviceCodeResponse struct {
	DeviceCode              string `json:"device_code"`
	UserCode                string `json:"user_code"`
	VerificationURI         string `json:"verification_uri"`
	VerificationURIComplete string `json:"verification_uri_complete"`
	ExpiresIn               int    `json:"expires_in"`
	Interval                int    `json:"interval"`
}

func RequestDeviceCode() (*DeviceCodeResponse, error) {
	payload := map[string]string{
		"client_id": config.NewConfig().GetClientID(),
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		return nil, NewDeviceCodeError(fmt.Errorf("error creating device code request JSON: %w", err))
	}

	const requestTimeout = 10 * time.Second
	ctx, cancel := context.WithTimeout(context.Background(), requestTimeout)
	defer cancel()

	deviceCodeURL := config.NewConfig().GetDeviceCodeURL()
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, deviceCodeURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, NewDeviceCodeError(fmt.Errorf("error creating device code request: %w", err))
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, NewDeviceCodeError(fmt.Errorf("error sending device code request: %w", err))
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, NewDeviceCodeError(fmt.Errorf("device code request failed with status: %d", resp.StatusCode))
	}

	var deviceCode DeviceCodeResponse
	if decodeErr := json.NewDecoder(resp.Body).Decode(&deviceCode); decodeErr != nil {
		return nil, NewDeviceCodeError(fmt.Errorf("error decoding device code response: %w", decodeErr))
	}

	return &deviceCode, nil
}
