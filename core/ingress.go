package core

import (
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/cephei8/greener/core/model/db"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/uptrace/bun"
)

type IngressHandler struct {
	db *bun.DB
}

func NewIngressHandler(db *bun.DB) *IngressHandler {
	return &IngressHandler{db: db}
}

type LabelRequest struct {
	Key   string  `json:"key"`
	Value *string `json:"value,omitempty"`
}

type SessionRequest struct {
	ID          *string        `json:"id,omitempty"`
	Description *string        `json:"description,omitempty"`
	Baggage     map[string]any `json:"baggage,omitempty"`
	Labels      []LabelRequest `json:"labels,omitempty"`
}

type SessionResponse struct {
	ID string `json:"id"`
}

type TestcaseRequest struct {
	SessionID         string         `json:"sessionId"`
	TestcaseName      string         `json:"testcaseName"`
	TestcaseClassname *string        `json:"testcaseClassname,omitempty"`
	TestcaseFile      *string        `json:"testcaseFile,omitempty"`
	Testsuite         *string        `json:"testsuite,omitempty"`
	Status            string         `json:"status"`
	Output            *string        `json:"output,omitempty"`
	Baggage           map[string]any `json:"baggage,omitempty"`
}

type TestcasesRequest struct {
	Testcases []TestcaseRequest `json:"testcases"`
}

func (h *IngressHandler) CreateSession(c echo.Context) error {
	userID := GetUserId(c)

	var req SessionRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid request body")
	}

	var sessionID uuid.UUID
	var err error
	if req.ID != nil && *req.ID != "" {
		sessionID, err = uuid.Parse(*req.ID)
		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, "Cannot parse session ID")
		}
	} else {
		sessionID = uuid.New()
	}

	var baggageJSON []byte
	if req.Baggage != nil {
		baggageJSON, err = json.Marshal(req.Baggage)
		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, "Invalid baggage format")
		}
	}

	ctx := c.Request().Context()
	now := time.Now()

	session := &model_db.Session{
		ID:          model_db.BinaryUUID(sessionID),
		Description: req.Description,
		Baggage:     baggageJSON,
		CreatedAt:   now,
		UpdatedAt:   now,
		UserID:      userID,
	}

	_, err = h.db.NewInsert().Model(session).Exec(ctx)
	if err != nil {
		if strings.Contains(err.Error(), "UNIQUE") || strings.Contains(err.Error(), "duplicate") {
			return echo.NewHTTPError(http.StatusBadRequest, "Session with this ID already exists")
		}
		c.Logger().Errorf("Failed to insert session: %v", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to create session")
	}

	if len(req.Labels) > 0 {
		for _, labelReq := range req.Labels {
			label := &model_db.Label{
				SessionID: model_db.BinaryUUID(sessionID),
				Key:       labelReq.Key,
				Value:     labelReq.Value,
				UserID:    userID,
				CreatedAt: now,
				UpdatedAt: now,
			}
			_, err = h.db.NewInsert().Model(label).Exec(ctx)
			if err != nil {
				c.Logger().Errorf("Failed to insert label: %v", err)
				return echo.NewHTTPError(http.StatusInternalServerError, "Failed to create label")
			}
		}
	}

	return c.JSON(http.StatusCreated, SessionResponse{ID: sessionID.String()})
}

func (h *IngressHandler) CreateTestcases(c echo.Context) error {
	userID := GetUserId(c)

	var req TestcasesRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid request body")
	}

	if len(req.Testcases) == 0 {
		return c.NoContent(http.StatusCreated)
	}

	ctx := c.Request().Context()
	now := time.Now()

	for _, tc := range req.Testcases {
		sessionID, err := uuid.Parse(tc.SessionID)
		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, "Cannot parse session ID")
		}

		var session model_db.Session
		err = h.db.NewSelect().
			Model(&session).
			Where("id = ?", model_db.BinaryUUID(sessionID)).
			Scan(ctx)
		if err != nil {
			if err.Error() == "sql: no rows in result set" {
				return echo.NewHTTPError(http.StatusBadRequest, "Unknown session ID")
			}
			c.Logger().Errorf("Failed to find session: %v", err)
			return echo.NewHTTPError(http.StatusInternalServerError, "Failed to verify session")
		}

		if session.UserID != userID {
			return echo.NewHTTPError(http.StatusBadRequest, "Session not found")
		}

		status, err := TestcaseStatusFromString(tc.Status)
		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, err.Error())
		}

		var baggageJSON []byte
		if tc.Baggage != nil {
			baggageJSON, err = json.Marshal(tc.Baggage)
			if err != nil {
				return echo.NewHTTPError(http.StatusBadRequest, "Invalid baggage format")
			}
		}

		testcase := &model_db.Testcase{
			ID:        model_db.BinaryUUID(uuid.New()),
			SessionID: model_db.BinaryUUID(sessionID),
			Name:      tc.TestcaseName,
			Classname: tc.TestcaseClassname,
			File:      tc.TestcaseFile,
			Testsuite: tc.Testsuite,
			Status:    status,
			Output:    tc.Output,
			Baggage:   baggageJSON,
			CreatedAt: now,
			UpdatedAt: now,
			UserID:    userID,
		}

		_, err = h.db.NewInsert().Model(testcase).Exec(ctx)
		if err != nil {
			c.Logger().Errorf("Failed to insert testcase: %v", err)
			return echo.NewHTTPError(http.StatusInternalServerError, "Failed to create testcase")
		}
	}

	return c.NoContent(http.StatusCreated)
}
