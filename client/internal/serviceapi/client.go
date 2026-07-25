package serviceapi

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

type Client struct {
	baseURL string
	http    *http.Client
}

type HTTPStatusError struct {
	StatusCode int
	Response   any
}

func (e *HTTPStatusError) Error() string {
	if e == nil {
		return ""
	}

	switch resp := e.Response.(type) {
	case QueryResponse:
		if resp.Error != nil {
			return fmt.Sprintf("service returned %d: %s", e.StatusCode, resp.Error.Message)
		}
	case *QueryResponse:
		if resp != nil && resp.Error != nil {
			return fmt.Sprintf("service returned %d: %s", e.StatusCode, resp.Error.Message)
		}
	}

	return fmt.Sprintf("service returned %d", e.StatusCode)
}

func New(baseURL string) *Client {
	return &Client{
		baseURL: strings.TrimRight(baseURL, "/"),
		http:    http.DefaultClient,
	}
}

func (c *Client) Health(ctx context.Context) (HealthResponse, error) {
	var response HealthResponse
	if err := c.get(ctx, "/healthz", &response); err != nil {
		return HealthResponse{}, err
	}

	return response, nil
}

func (c *Client) Ready(ctx context.Context) (ReadinessResponse, error) {
	var response ReadinessResponse
	status, err := c.doJSON(ctx, http.MethodGet, "/readyz", nil, &response)
	if err != nil {
		return ReadinessResponse{}, err
	}

	if status >= http.StatusBadRequest {
		return response, &HTTPStatusError{StatusCode: status, Response: response}
	}

	return response, nil
}

func (c *Client) Query(ctx context.Context, request QueryRequest) (QueryResponse, error) {
	var response QueryResponse
	status, err := c.doJSON(ctx, http.MethodPost, "/v1/query", request, &response)
	if err != nil {
		return QueryResponse{}, err
	}

	if status >= http.StatusBadRequest {
		return response, &HTTPStatusError{StatusCode: status, Response: response}
	}

	return response, nil
}

func (c *Client) StreamQuery(ctx context.Context, request QueryRequest) (*QueryStream, error) {
	httpRequest, err := c.newRequest(ctx, http.MethodPost, "/v1/query/stream", request)
	if err != nil {
		return nil, err
	}

	httpRequest.Header.Set("Accept", "text/event-stream")

	response, err := c.http.Do(httpRequest)
	if err != nil {
		return nil, err
	}

	if response.StatusCode >= http.StatusBadRequest {
		var queryResponse QueryResponse
		if err := json.NewDecoder(response.Body).Decode(&queryResponse); err != nil {
			response.Body.Close()
			return nil, err
		}
		response.Body.Close()
		return nil, &HTTPStatusError{StatusCode: response.StatusCode, Response: queryResponse}
	}

	stream := &QueryStream{response: response}
	stream.scanner = bufio.NewScanner(response.Body)
	stream.scanner.Buffer(make([]byte, 0, 64*1024), 1024*1024)

	return stream, nil
}

func (c *Client) Ingest(ctx context.Context, request IngestRequest) (IngestResponse, error) {
	var response IngestResponse
	status, err := c.doJSON(ctx, http.MethodPost, "/v1/documents/ingest", request, &response)
	if err != nil {
		return IngestResponse{}, err
	}

	if status >= http.StatusBadRequest {
		return response, &HTTPStatusError{StatusCode: status, Response: response}
	}

	return response, nil
}

func (c *Client) get(ctx context.Context, path string, dst any) error {
	status, err := c.doJSON(ctx, http.MethodGet, path, nil, dst)
	if err != nil {
		return err
	}

	if status >= http.StatusBadRequest {
		return fmt.Errorf("service returned %d", status)
	}

	return nil
}

func (c *Client) doJSON(ctx context.Context, method, path string, request any, dst any) (int, error) {
	httpRequest, err := c.newRequest(ctx, method, path, request)
	if err != nil {
		return 0, err
	}

	response, err := c.http.Do(httpRequest)
	if err != nil {
		return 0, err
	}
	defer response.Body.Close()

	if dst != nil {
		if err := json.NewDecoder(response.Body).Decode(dst); err != nil {
			return response.StatusCode, err
		}
	}

	return response.StatusCode, nil
}

func (c *Client) newRequest(ctx context.Context, method, path string, request any) (*http.Request, error) {
	var body *bytes.Reader
	if request == nil {
		body = bytes.NewReader(nil)
	} else {
		buf := new(bytes.Buffer)
		if err := json.NewEncoder(buf).Encode(request); err != nil {
			return nil, err
		}
		body = bytes.NewReader(buf.Bytes())
	}

	httpRequest, err := http.NewRequestWithContext(ctx, method, c.baseURL+path, body)
	if err != nil {
		return nil, err
	}

	if request != nil {
		httpRequest.Header.Set("Content-Type", "application/json")
	}

	return httpRequest, nil
}

type QueryStream struct {
	response *http.Response
	scanner  *bufio.Scanner
	done     bool
}

func (s *QueryStream) Close() error {
	if s == nil || s.response == nil {
		return nil
	}

	return s.response.Body.Close()
}

func (s *QueryStream) Next() (QueryResponse, error) {
	if s == nil || s.done {
		return QueryResponse{}, io.EOF
	}

	var dataLines []string
	for {
		if !s.scanner.Scan() {
			if err := s.scanner.Err(); err != nil {
				return QueryResponse{}, err
			}

			s.done = true
			if len(dataLines) == 0 {
				return QueryResponse{}, io.EOF
			}
			break
		}

		line := s.scanner.Text()
		if line == "" {
			if len(dataLines) == 0 {
				continue
			}
			break
		}

		if strings.HasPrefix(line, ":") {
			continue
		}

		if strings.HasPrefix(line, "data:") {
			dataLines = append(dataLines, strings.TrimSpace(strings.TrimPrefix(line, "data:")))
		}
	}

	if len(dataLines) == 0 {
		return QueryResponse{}, io.EOF
	}

	var response QueryResponse
	if err := json.Unmarshal([]byte(strings.Join(dataLines, "\n")), &response); err != nil {
		return QueryResponse{}, err
	}

	return response, nil
}
