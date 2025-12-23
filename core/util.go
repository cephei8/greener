package core

import (
	"fmt"

	"github.com/cephei8/greener/core/model/db"
)

func TestcaseStatusFromString(status string) (model_db.TestcaseStatus, error) {
	switch status {
	case "pass":
		return model_db.StatusPass, nil
	case "fail":
		return model_db.StatusFail, nil
	case "error":
		return model_db.StatusError, nil
	case "skip":
		return model_db.StatusSkip, nil
	default:
		return 0, fmt.Errorf("invalid status: %s", status)
	}
}

func TestcaseStatusToString(status model_db.TestcaseStatus) string {
	switch status {
	case model_db.StatusPass:
		return "pass"
	case model_db.StatusFail:
		return "fail"
	case model_db.StatusError:
		return "error"
	case model_db.StatusSkip:
		return "skip"
	default:
		panic(fmt.Sprintf("unknown status: %d", status))
	}
}
