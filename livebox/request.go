package livebox

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
)

type apiRequest struct {
	Method     string         `json:"method"`
	Service    string         `json:"service"`
	Parameters map[string]any `json:"parameters"`
}

type apiResponse struct {
	Status json.RawMessage `json:"status"`
	Errors json.RawMessage `json:"errors"`
}

func (c *Client) doReq(r *apiRequest) (json.RawMessage, error) {
	b, err := json.Marshal(r)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest(http.MethodPost, c.host+"/ws", bytes.NewReader(b))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "X-Sah "+c.token)
	req.Header.Set("Content-Type", "application/x-sah-ws-4-call+json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()

	var apiResp apiResponse
	if err = json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
		return nil, err
	}

	if len(apiResp.Errors) > 0 {
		return nil, fmt.Errorf("api error: %s", apiResp.Errors)
	}

	return apiResp.Status, nil
}

func (c *Client) doAuthReq(r *apiRequest) (*http.Response, error) {
	b, err := json.Marshal(r)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest(http.MethodPost, c.host+"/ws", bytes.NewReader(b))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "X-Sah-Login")
	req.Header.Set("Content-Type", "application/x-sah-ws-4-call+json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}

	return resp, nil
}
