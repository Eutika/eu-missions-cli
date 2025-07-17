package services

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/eutika/eu-missions-cli/internal/auth"
	"github.com/eutika/eu-missions-cli/internal/config"
	"github.com/eutika/eu-missions-cli/pkg/types"
)

type RemoteService struct {
	config *config.Config
	client *http.Client
}

func NewRemoteService(cfg *config.Config) *RemoteService {
	return &RemoteService{
		config: cfg,
		client: &http.Client{},
	}
}

// createAuthenticatedRequest creates an HTTP request with authentication token.
func (s *RemoteService) createAuthenticatedRequest(
	ctx context.Context, method, url string, body io.Reader,
) (*http.Request, error) {
	req, err := http.NewRequestWithContext(ctx, method, url, body)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	token, err := auth.GetCurrentToken()
	if err != nil {
		return nil, fmt.Errorf("authentication token not found: %w", err)
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))
	if method == http.MethodPost {
		req.Header.Set("Content-Type", "application/json")
	}

	return req, nil
}

// executeRequest executes an HTTP request and handles common response processing.
func (s *RemoteService) executeRequest(req *http.Request) ([]byte, error) {
	resp, err := s.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
		body, errStatusCode := io.ReadAll(resp.Body)
		if errStatusCode != nil {
			return nil,
				fmt.Errorf("request failed with status %d, could not read response body: %w", resp.StatusCode, errStatusCode)
		}

		var headers strings.Builder
		for key, values := range resp.Header {
			headers.WriteString(fmt.Sprintf("%s: %s\n", key, strings.Join(values, ", ")))
		}

		return nil, fmt.Errorf("request failed with status %d\nBody: %s\nHeaders:\n%s", resp.StatusCode, string(body), headers.String())
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	return body, nil
}

// createCommandResultPayload creates a JSON payload for command results.
func (s *RemoteService) createCommandResultPayload(id string, results []string) ([]byte, error) {
	resultPayload := types.CommandResult{
		ID:      id,
		Results: results,
	}

	jsonPayload, err := json.Marshal(resultPayload)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal result payload: %w", err)
	}

	return jsonPayload, nil
}

// unmarshalResponse unmarshals the response body into the provided interface.
func (s *RemoteService) unmarshalResponse(data []byte, v interface{}) error {
	if err := json.Unmarshal(data, v); err != nil {
		return fmt.Errorf("failed to unmarshal response: %w", err)
	}
	return nil
}

func (s *RemoteService) FetchCommands(id string) ([]string, error) {
	url := fmt.Sprintf("%s/commands/%s", s.config.GetRemoteURL(), id)
	req, err := s.createAuthenticatedRequest(context.Background(), http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}

	body, err := s.executeRequest(req)
	if err != nil {
		return nil, err
	}

	var cmd types.Command
	if errUnmarshalFetch := s.unmarshalResponse(body, &cmd); err != nil {
		return nil, errUnmarshalFetch
	}

	return cmd.Commands, nil
}

func (s *RemoteService) SendCommandResult(command string, id string, results []string) (map[string]interface{}, error) {
	jsonPayload, err := s.createCommandResultPayload(id, results)
	if err != nil {
		return nil, err
	}

	url := fmt.Sprintf("%s/%s", s.config.GetRemoteURL(), command)
	req, err := s.createAuthenticatedRequest(context.Background(), http.MethodPost, url, bytes.NewBuffer(jsonPayload))
	if err != nil {
		return nil, err
	}

	body, err := s.executeRequest(req)
	if err != nil {
		return nil, err
	}

	var response map[string]interface{}
	errUnmarshalSend := s.unmarshalResponse(body, &response)
	if errUnmarshalSend != nil {
		return nil, errUnmarshalSend
	}

	return response, nil
}

func (s *RemoteService) ValidateCommandResult(id string, results []string) (map[string]interface{}, error) {
	jsonPayload, err := s.createCommandResultPayload(id, results)
	if err != nil {
		return nil, err
	}

	url := fmt.Sprintf("%s/validate", s.config.GetRemoteURL())
	req, err := s.createAuthenticatedRequest(context.Background(), http.MethodPost, url, bytes.NewBuffer(jsonPayload))
	if err != nil {
		return nil, err
	}

	body, err := s.executeRequest(req)
	if err != nil {
		return nil, err
	}

	var response map[string]interface{}
	if errUnmarshalValidate := s.unmarshalResponse(body, &response); err != nil {
		return nil, errUnmarshalValidate
	}

	return response, nil
}
