package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"

	"github.com/google/uuid"
	"github.com/urfave/cli/v3"
)

const (
	EnvVarEndpoint = "GREENER_INGRESS_ENDPOINT"
	EnvVarAPIKey   = "GREENER_INGRESS_API_KEY"
)

type TestcaseStatus string

const (
	StatusPass  TestcaseStatus = "pass"
	StatusFail  TestcaseStatus = "fail"
	StatusError TestcaseStatus = "error"
	StatusSkip  TestcaseStatus = "skip"
)

type Label struct {
	Key   string  `json:"key"`
	Value *string `json:"value,omitempty"`
}

type SessionRequest struct {
	Id          *string        `json:"id,omitempty"`
	Description *string        `json:"description,omitempty"`
	Baggage     map[string]any `json:"baggage,omitempty"`
	Labels      []Label        `json:"labels,omitempty"`
}

type SessionResponse struct {
	Id string `json:"id"`
}

type TestcaseRequest struct {
	SessionId         string         `json:"sessionId"`
	TestcaseName      string         `json:"testcaseName"`
	TestcaseClassname *string        `json:"testcaseClassname,omitempty"`
	TestcaseFile      *string        `json:"testcaseFile,omitempty"`
	Testsuite         *string        `json:"testsuite,omitempty"`
	Status            TestcaseStatus `json:"status"`
	Output            *string        `json:"output,omitempty"`
	Baggage           map[string]any `json:"baggage,omitempty"`
}

type TestcasesRequest struct {
	Testcases []TestcaseRequest `json:"testcases"`
}

type ErrorResponse struct {
	Detail string `json:"detail"`
}

type Client struct {
	httpClient *http.Client
	endpoint   string
	apiKey     string
}

func NewClient(endpoint, apiKey string) *Client {
	return &Client{
		httpClient: &http.Client{},
		endpoint:   endpoint,
		apiKey:     apiKey,
	}
}

func (c *Client) CreateSession(req SessionRequest) (string, error) {
	url := fmt.Sprintf("%s/api/v1/ingress/sessions", c.endpoint)

	jsonData, err := json.Marshal(req)
	if err != nil {
		return "", fmt.Errorf("failed to marshal session request: %w", err)
	}

	httpReq, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return "", fmt.Errorf("failed to create session request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("X-API-Key", c.apiKey)

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return "", fmt.Errorf("failed to send session request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		var errorResp ErrorResponse
		body, _ := io.ReadAll(resp.Body)
		if err := json.Unmarshal(body, &errorResp); err != nil {
			return "", fmt.Errorf("failed session request (%d): %s", resp.StatusCode, string(body))
		}
		return "", fmt.Errorf("failed session request (%d): %s", resp.StatusCode, errorResp.Detail)
	}

	var sessionResp SessionResponse
	if err := json.NewDecoder(resp.Body).Decode(&sessionResp); err != nil {
		return "", fmt.Errorf("failed to parse session response: %w", err)
	}

	return sessionResp.Id, nil
}

func (c *Client) CreateTestcases(testcases []TestcaseRequest) error {
	url := fmt.Sprintf("%s/api/v1/ingress/testcases", c.endpoint)

	req := TestcasesRequest{Testcases: testcases}

	jsonData, err := json.Marshal(req)
	if err != nil {
		return fmt.Errorf("failed to marshal testcases request: %w", err)
	}

	httpReq, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to create testcases request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("X-API-Key", c.apiKey)

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return fmt.Errorf("failed to send testcases request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		var errorResp ErrorResponse
		body, _ := io.ReadAll(resp.Body)
		if err := json.Unmarshal(body, &errorResp); err != nil {
			return fmt.Errorf("failed testcases request (%d): %s", resp.StatusCode, string(body))
		}
		return fmt.Errorf("failed testcases request (%d): %s", resp.StatusCode, errorResp.Detail)
	}

	return nil
}

func parseLabels(labelStrings []string) ([]Label, error) {
	if len(labelStrings) == 0 {
		return nil, nil
	}

	labels := make([]Label, 0, len(labelStrings))
	keysSet := make(map[string]bool)

	for _, labelStr := range labelStrings {
		var key string
		var value *string

		if before, after, ok := strings.Cut(labelStr, "="); ok {
			key = before
			val := after
			value = &val
		} else {
			key = labelStr
		}

		if key == "" {
			return nil, fmt.Errorf("label key cannot be empty")
		}

		if keysSet[key] {
			return nil, fmt.Errorf("duplicate label key: %s", key)
		}
		keysSet[key] = true

		labels = append(labels, Label{Key: key, Value: value})
	}

	return labels, nil
}

func parseBaggage(baggageStr string) (map[string]any, error) {
	if baggageStr == "" {
		return nil, nil
	}

	var baggage map[string]any
	if err := json.Unmarshal([]byte(baggageStr), &baggage); err != nil {
		return nil, fmt.Errorf("invalid baggage JSON: %w", err)
	}

	return baggage, nil
}

func validateStatus(status string) (TestcaseStatus, error) {
	s := TestcaseStatus(status)
	switch s {
	case StatusPass, StatusFail, StatusError, StatusSkip:
		return s, nil
	default:
		return "", fmt.Errorf("invalid status: %s. Valid values: pass, fail, error, skip", status)
	}
}

func getRequiredGlobalFlag(cmd *cli.Command, name string) (string, error) {
	value := cmd.String(name)
	if value == "" {
		return "", fmt.Errorf("--%s is required (or set %s environment variable)", name, name)
	}
	return value, nil
}

func createSessionAction(ctx context.Context, cmd *cli.Command) error {
	endpoint, err := getRequiredGlobalFlag(cmd, "endpoint")
	if err != nil {
		return err
	}

	apiKey, err := getRequiredGlobalFlag(cmd, "api-key")
	if err != nil {
		return err
	}

	client := NewClient(endpoint, apiKey)

	var sessionID *string
	if id := cmd.String("id"); id != "" {
		sessionID = &id
	}

	var description *string
	if desc := cmd.String("description"); desc != "" {
		description = &desc
	}

	var baggage map[string]any
	if baggageStr := cmd.String("baggage"); baggageStr != "" {
		baggage, err = parseBaggage(baggageStr)
		if err != nil {
			return err
		}
	}

	var labels []Label
	if labelStrings := cmd.StringSlice("label"); len(labelStrings) > 0 {
		labels, err = parseLabels(labelStrings)
		if err != nil {
			return err
		}
	}

	req := SessionRequest{
		Id:          sessionID,
		Description: description,
		Baggage:     baggage,
		Labels:      labels,
	}

	id, err := client.CreateSession(req)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to create session: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Created session ID: %s\n", id)
	return nil
}

func createTestcaseAction(ctx context.Context, cmd *cli.Command) error {
	endpoint, err := getRequiredGlobalFlag(cmd, "endpoint")
	if err != nil {
		return err
	}

	apiKey, err := getRequiredGlobalFlag(cmd, "api-key")
	if err != nil {
		return err
	}

	sessionID := cmd.String("session-id")
	if sessionID == "" {
		return fmt.Errorf("--session-id is required")
	}

	if _, err := uuid.Parse(sessionID); err != nil {
		return fmt.Errorf("invalid session ID format: %w", err)
	}

	name := cmd.String("name")
	if name == "" {
		return fmt.Errorf("--name is required")
	}

	statusStr := cmd.String("status")
	status, err := validateStatus(statusStr)
	if err != nil {
		return err
	}

	client := NewClient(endpoint, apiKey)

	var output *string
	if out := cmd.String("output"); out != "" {
		output = &out
	}

	var classname *string
	if cn := cmd.String("classname"); cn != "" {
		classname = &cn
	}

	var file *string
	if f := cmd.String("file"); f != "" {
		file = &f
	}

	var testsuite *string
	if ts := cmd.String("testsuite"); ts != "" {
		testsuite = &ts
	}

	var baggage map[string]any
	if baggageStr := cmd.String("baggage"); baggageStr != "" {
		baggage, err = parseBaggage(baggageStr)
		if err != nil {
			return err
		}
	}

	testcase := TestcaseRequest{
		SessionId:         sessionID,
		TestcaseName:      name,
		TestcaseClassname: classname,
		TestcaseFile:      file,
		Testsuite:         testsuite,
		Status:            status,
		Output:            output,
		Baggage:           baggage,
	}

	if err := client.CreateTestcases([]TestcaseRequest{testcase}); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to create testcase: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("Successfully reported testcase")
	return nil
}

func main() {
	cmd := &cli.Command{
		Name:  "greener-reporter-cli",
		Usage: "CLI tool for Greener reporting",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:    "endpoint",
				Usage:   "Greener ingress endpoint URL",
				Sources: cli.EnvVars(EnvVarEndpoint),
			},
			&cli.StringFlag{
				Name:    "api-key",
				Usage:   "API key for authentication",
				Sources: cli.EnvVars(EnvVarAPIKey),
			},
		},
		Commands: []*cli.Command{
			{
				Name:  "create",
				Usage: "Create results",
				Commands: []*cli.Command{
					{
						Name:   "session",
						Usage:  "Create session",
						Action: createSessionAction,
						Flags: []cli.Flag{
							&cli.StringFlag{
								Name:  "id",
								Usage: "ID for the session",
							},
							&cli.StringFlag{
								Name:  "description",
								Usage: "Description of the session",
							},
							&cli.StringFlag{
								Name:  "baggage",
								Usage: "Additional metadata as JSON",
							},
							&cli.StringSliceFlag{
								Name:  "label",
								Usage: "Labels in `key` or `key=value` format",
							},
						},
					},
					{
						Name:   "testcase",
						Usage:  "Create test case",
						Action: createTestcaseAction,
						Flags: []cli.Flag{
							&cli.StringFlag{
								Name:     "session-id",
								Usage:    "Session ID for the test case",
								Required: true,
							},
							&cli.StringFlag{
								Name:     "name",
								Usage:    "Name of the test case",
								Required: true,
							},
							&cli.StringFlag{
								Name:  "output",
								Usage: "Output from the test case",
							},
							&cli.StringFlag{
								Name:  "classname",
								Usage: "Class name of the test case",
							},
							&cli.StringFlag{
								Name:  "file",
								Usage: "File path of the test case",
							},
							&cli.StringFlag{
								Name:  "testsuite",
								Usage: "Test suite name",
							},
							&cli.StringFlag{
								Name:  "status",
								Usage: "Test case status (pass, fail, error, skip)",
								Value: "pass",
							},
							&cli.StringFlag{
								Name:  "baggage",
								Usage: "Additional metadata as JSON",
							},
						},
					},
				},
			},
		},
	}

	if err := cmd.Run(context.Background(), os.Args); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
