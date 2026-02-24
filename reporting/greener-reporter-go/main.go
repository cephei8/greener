package main

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/urfave/cli/v3"
)

type ActionType string

const (
	actionRun    ActionType = "run"
	actionPass   ActionType = "pass"
	actionFail   ActionType = "fail"
	actionSkip   ActionType = "skip"
	actionOutput ActionType = "output"
)

const (
	ingressEndpointFlag    = "ingress-endpoint"
	ingressAPIKeyFlag      = "ingress-api-key"
	inputFileFlag          = "input-file"
	sessionIDFlag          = "session-id"
	sessionDescriptionFlag = "session-description"
	sessionLabelsFlag      = "session-labels"
	sessionBaggageFlag     = "session-baggage"
	verboseFlag            = "verbose"
)

const (
	statusPass  = "pass"
	statusFail  = "fail"
	statusSkip  = "skip"
	statusError = "error"
)

type Event struct {
	Time    string     `json:"Time"`
	Action  ActionType `json:"Action"`
	Package string     `json:"Package"`
	Test    string     `json:"Test,omitempty"`
	Output  string     `json:"Output,omitempty"`
	Elapsed float64    `json:"Elapsed,omitempty"`
}

type Label struct {
	Key   string `json:"key"`
	Value string `json:"value,omitempty"`
}

type SessionRequest struct {
	Id          string         `json:"id,omitempty"`
	Description string         `json:"description,omitempty"`
	Labels      []Label        `json:"labels,omitempty"`
	Baggage     map[string]any `json:"baggage,omitempty"`
}

type SessionResponse struct {
	Id string `json:"id"`
}

type TestcaseRequest struct {
	SessionId         string         `json:"sessionId"`
	TestcaseName      string         `json:"testcaseName"`
	TestcaseClassname string         `json:"testcaseClassname,omitempty"`
	TestcaseFile      string         `json:"testcaseFile,omitempty"`
	Testsuite         string         `json:"testsuite,omitempty"`
	Status            string         `json:"status"`
	Output            string         `json:"output,omitempty"`
	Baggage           map[string]any `json:"baggage,omitempty"`
}

type TestcasesRequest struct {
	Testcases []TestcaseRequest `json:"testcases"`
}

type TestResultKey struct {
	Package string
	Test    string
}

type TestResult struct {
	Package string
	Test    string
	Status  string
	Output  strings.Builder
}

func main() {
	cmd := &cli.Command{
		Name:  "greener-reporter",
		Usage: "Report Go test results to Greener",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:     ingressEndpointFlag,
				Usage:    "Greener ingress endpoint",
				Sources:  cli.EnvVars("GREENER_INGRESS_ENDPOINT"),
				Required: true,
			},
			&cli.StringFlag{
				Name:     ingressAPIKeyFlag,
				Usage:    "Greener ingress API key",
				Sources:  cli.EnvVars("GREENER_INGRESS_API_KEY"),
				Required: true,
			},
			&cli.StringFlag{
				Name:     inputFileFlag,
				Usage:    "Path to Go test JSON output file (use '-' for stdin)",
				Aliases:  []string{"f"},
				Required: true,
			},
			&cli.StringFlag{
				Name:    sessionIDFlag,
				Usage:   "Session ID (optional, will be generated if not provided)",
				Sources: cli.EnvVars("GREENER_SESSION_ID"),
			},
			&cli.StringFlag{
				Name:    sessionDescriptionFlag,
				Usage:   "Session description",
				Sources: cli.EnvVars("GREENER_SESSION_DESCRIPTION"),
			},
			&cli.StringFlag{
				Name:    sessionLabelsFlag,
				Usage:   "Session labels (comma-separated, e.g. 'ci,tag=value')",
				Sources: cli.EnvVars("GREENER_SESSION_LABELS"),
			},
			&cli.StringFlag{
				Name:    sessionBaggageFlag,
				Usage:   "Session baggage (JSON object)",
				Sources: cli.EnvVars("GREENER_SESSION_BAGGAGE"),
			},
			&cli.BoolFlag{
				Name:    verboseFlag,
				Usage:   "Enable verbose logging",
				Aliases: []string{"v"},
			},
		},
		Action: run,
	}

	if err := cmd.Run(context.Background(), os.Args); err != nil {
		log.Fatal(err)
	}
}

type Reporter struct {
	endpoint           string
	apiKey             string
	sessionID          string
	sessionDescription string
	sessionLabels      []Label
	sessionBaggage     map[string]any
	verbose            bool
	client             *http.Client
	results            map[TestResultKey]*TestResult
	resultsChan        chan *TestResult
	batcherDone        chan struct{}
}

func NewReporter(endpoint, apiKey, sessionID, sessionDescription string, sessionLabels []Label, sessionBaggage map[string]any, verbose bool) *Reporter {
	return &Reporter{
		endpoint:           strings.TrimSuffix(endpoint, "/"),
		apiKey:             apiKey,
		sessionID:          sessionID,
		sessionDescription: sessionDescription,
		sessionLabels:      sessionLabels,
		sessionBaggage:     sessionBaggage,
		verbose:            verbose,
		client:             &http.Client{},
		results:            make(map[TestResultKey]*TestResult),
		resultsChan:        make(chan *TestResult),
		batcherDone:        make(chan struct{}),
	}
}

func parseLabels(labelsStr string) []Label {
	if labelsStr == "" {
		return nil
	}

	var labels []Label
	for _, labelStr := range strings.Split(labelsStr, ",") {
		labelStr = strings.TrimSpace(labelStr)
		if labelStr == "" {
			continue
		}

		if idx := strings.Index(labelStr, "="); idx != -1 {
			labels = append(labels, Label{
				Key:   labelStr[:idx],
				Value: labelStr[idx+1:],
			})
		} else {
			labels = append(labels, Label{
				Key: labelStr,
			})
		}
	}
	return labels
}

func (r *Reporter) createSession() error {
	req := SessionRequest{
		Id:          r.sessionID,
		Description: r.sessionDescription,
		Labels:      r.sessionLabels,
		Baggage:     r.sessionBaggage,
	}

	body, err := json.Marshal(req)
	if err != nil {
		return fmt.Errorf("marshal session request: %w", err)
	}

	httpReq, err := http.NewRequest(
		"POST", r.endpoint+"/api/v1/ingress/sessions", bytes.NewReader(body),
	)
	if err != nil {
		return fmt.Errorf("create session request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("x-api-key", r.apiKey)

	resp, err := r.client.Do(httpReq)
	if err != nil {
		return fmt.Errorf("send session request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return fmt.Errorf(
			"create session failed: status=%d body=%s",
			resp.StatusCode,
			string(bodyBytes),
		)
	}

	var sessionResp SessionResponse
	if err := json.NewDecoder(resp.Body).Decode(&sessionResp); err != nil {
		return fmt.Errorf("decode session response: %w", err)
	}

	r.sessionID = sessionResp.Id
	log.Printf("Created session: %s\n", r.sessionID)
	return nil
}

func (r *Reporter) submitBatch(testcases []TestcaseRequest) error {
	if len(testcases) == 0 {
		return nil
	}

	req := TestcasesRequest{
		Testcases: testcases,
	}

	body, err := json.Marshal(req)
	if err != nil {
		return fmt.Errorf("marshal testcases request: %w", err)
	}

	httpReq, err := http.NewRequest(
		"POST", r.endpoint+"/api/v1/ingress/testcases", bytes.NewReader(body),
	)
	if err != nil {
		return fmt.Errorf("create testcases request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("x-api-key", r.apiKey)

	resp, err := r.client.Do(httpReq)
	if err != nil {
		return fmt.Errorf("send testcases request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return fmt.Errorf(
			"submit testcases failed: status=%d body=%s",
			resp.StatusCode,
			string(bodyBytes),
		)
	}

	log.Printf("Submitted %d test results\n", len(testcases))
	return nil
}

func (r *Reporter) batch() {
	defer close(r.batcherDone)

	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	var batch []TestcaseRequest

	submit := func() {
		if err := r.submitBatch(batch); err != nil {
			log.Printf("Error submitting batch: %v\n", err)
		}
		batch = nil
	}

	for {
		select {
		case result, ok := <-r.resultsChan:
			if !ok {
				submit()
				return
			}

			batch = append(batch, TestcaseRequest{
				SessionId:         r.sessionID,
				TestcaseName:      result.Test,
				TestcaseClassname: result.Package,
				Testsuite:         result.Package,
				Status:            result.Status,
				Output:            result.Output.String(),
			})

			if len(batch) >= 100 {
				submit()
			}

		case <-ticker.C:
			submit()
		}
	}
}

func (r *Reporter) handleEvent(ev Event, verbose bool) {
	if ev.Test == "" {
		return
	}

	key := TestResultKey{
		Package: ev.Package,
		Test:    ev.Test,
	}

	result, ok := r.results[key]
	if !ok {
		if ev.Action == actionOutput {
			if verbose {
				log.Printf("Skipping output for unknown test result")
			}
			return
		} else {
			result = &TestResult{
				Package: ev.Package,
				Test:    ev.Test,
				Status:  statusError,
			}
			r.results[key] = result
		}
	}

	switch ev.Action {
	case actionRun:
	case actionPass:
		result.Status = statusPass
		r.resultsChan <- result
		delete(r.results, key)
	case actionFail:
		result.Status = statusFail
		r.resultsChan <- result
		delete(r.results, key)
	case actionSkip:
		result.Status = statusSkip
		r.resultsChan <- result
		delete(r.results, key)
	case actionOutput:
		result.Output.WriteString(ev.Output)
	}
}

func run(ctx context.Context, c *cli.Command) error {
	endpoint := c.String(ingressEndpointFlag)
	apiKey := c.String(ingressAPIKeyFlag)
	inputFile := c.String(inputFileFlag)
	sessionID := c.String(sessionIDFlag)
	sessionDescription := c.String(sessionDescriptionFlag)
	sessionLabelsStr := c.String(sessionLabelsFlag)
	sessionBaggageStr := c.String(sessionBaggageFlag)
	verbose := c.Bool(verboseFlag)

	sessionLabels := parseLabels(sessionLabelsStr)

	var sessionBaggage map[string]any
	if sessionBaggageStr != "" {
		if err := json.Unmarshal([]byte(sessionBaggageStr), &sessionBaggage); err != nil {
			return fmt.Errorf("parse session baggage: %w", err)
		}
	}

	reporter := NewReporter(endpoint, apiKey, sessionID, sessionDescription, sessionLabels, sessionBaggage, verbose)

	if err := reporter.createSession(); err != nil {
		return fmt.Errorf("create session: %w", err)
	}

	go reporter.batch()

	var reader io.Reader
	if inputFile == "-" {
		reader = os.Stdin
	} else {
		file, err := os.Open(inputFile)
		if err != nil {
			return fmt.Errorf("open file %s: %w", inputFile, err)
		}
		defer file.Close()
		reader = file
	}

	scanner := bufio.NewScanner(reader)
	for scanner.Scan() {
		line := scanner.Bytes()
		if len(line) == 0 {
			continue
		}

		var ev Event
		if err := json.Unmarshal(line, &ev); err != nil {
			if reporter.verbose {
				log.Printf("Skipping invalid JSON line: %v", err)
			}
			continue
		}

		reporter.handleEvent(ev, reporter.verbose)
	}

	if err := scanner.Err(); err != nil {
		return fmt.Errorf("read input: %w", err)
	}

	close(reporter.resultsChan)
	<-reporter.batcherDone

	return nil
}
