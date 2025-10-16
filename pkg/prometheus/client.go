package prometheus

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"
)

// Client is a simple Prometheus API client
type Client struct {
	baseURL string
	client  *http.Client
}

// QueryResult represents a Prometheus query result
type QueryResult struct {
	Status string `json:"status"`
	Data   struct {
		ResultType string `json:"resultType"`
		Result     []struct {
			Metric map[string]string `json:"metric"`
			Value  []interface{}     `json:"value"`
		} `json:"result"`
	} `json:"data"`
}

// NewClient creates a new Prometheus client
func NewClient(prometheusURL string) *Client {
	return &Client{
		baseURL: prometheusURL,
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

// Query executes a PromQL query
func (c *Client) Query(ctx context.Context, query string) (*QueryResult, error) {
	params := url.Values{}
	params.Add("query", query)
	params.Add("time", fmt.Sprintf("%d", time.Now().Unix()))

	queryURL := fmt.Sprintf("%s/api/v1/query?%s", c.baseURL, params.Encode())

	req, err := http.NewRequestWithContext(ctx, "GET", queryURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute query: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("prometheus returned status %d: %s", resp.StatusCode, string(body))
	}

	var result QueryResult
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	if result.Status != "success" {
		return nil, fmt.Errorf("query failed with status: %s", result.Status)
	}

	return &result, nil
}

// GetScalarValue extracts a scalar value from query result
func (r *QueryResult) GetScalarValue() (float64, error) {
	if len(r.Data.Result) == 0 {
		return 0, nil
	}

	if len(r.Data.Result[0].Value) < 2 {
		return 0, fmt.Errorf("invalid value format")
	}

	// Value is [timestamp, "value_as_string"]
	valueStr, ok := r.Data.Result[0].Value[1].(string)
	if !ok {
		return 0, fmt.Errorf("value is not a string")
	}

	var value float64
	if _, err := fmt.Sscanf(valueStr, "%f", &value); err != nil {
		return 0, fmt.Errorf("failed to parse value: %w", err)
	}

	return value, nil
}

// GetVectorValues extracts vector values with labels
func (r *QueryResult) GetVectorValues() (map[string]float64, error) {
	values := make(map[string]float64)

	for _, result := range r.Data.Result {
		if len(result.Value) < 2 {
			continue
		}

		valueStr, ok := result.Value[1].(string)
		if !ok {
			continue
		}

		var value float64
		if _, err := fmt.Sscanf(valueStr, "%f", &value); err != nil {
			continue
		}

		// Use job label as key, or combine multiple labels
		key := result.Metric["job"]
		if key == "" {
			key = result.Metric["instance"]
		}
		if key == "" {
			key = "unknown"
		}

		values[key] = value
	}

	return values, nil
}
