package service

import (
	"errors"
	"strings"
)

type ContentStatusInput struct {
	Status string `json:"status"`
}

var allowedContentStatusTransitions = map[string]map[string]struct{}{
	"draft": {
		"published": {},
	},
	"published": {
		"draft":    {},
		"archived": {},
	},
	"archived": {
		"draft": {},
	},
}

var allowedLeadStatusTransitions = map[string]map[string]struct{}{
	"new": {
		"contacted": {},
		"invalid":   {},
	},
	"contacted": {
		"following": {},
		"invalid":   {},
	},
	"following": {
		"contacted": {},
		"converted": {},
		"invalid":   {},
	},
}

func normalizeCreateContentStatus(status string) (string, error) {
	status = strings.TrimSpace(status)
	if status == "" {
		return "draft", nil
	}

	switch status {
	case "draft", "published":
		return status, nil
	default:
		return "", errors.New("invalid content status")
	}
}

func normalizeUpdateContentStatus(currentStatus string, requestedStatus string) (string, error) {
	requestedStatus = strings.TrimSpace(requestedStatus)
	if requestedStatus == "" {
		return currentStatus, nil
	}

	if !isValidContentStatus(requestedStatus) {
		return "", errors.New("invalid content status")
	}

	if requestedStatus == currentStatus {
		return requestedStatus, nil
	}

	nextStatuses, ok := allowedContentStatusTransitions[currentStatus]
	if !ok {
		return "", errors.New("invalid current content status")
	}

	if _, ok := nextStatuses[requestedStatus]; !ok {
		return "", errors.New("invalid content status transition")
	}

	return requestedStatus, nil
}

func isValidContentStatus(status string) bool {
	switch status {
	case "draft", "published", "archived":
		return true
	default:
		return false
	}
}

func normalizeCreateLeadStatus(status string) (string, error) {
	status = strings.TrimSpace(status)
	if status == "" {
		return "new", nil
	}

	if !isValidLeadStatus(status) {
		return "", errors.New("invalid lead status")
	}

	return status, nil
}

func normalizeUpdateLeadStatus(currentStatus string, requestedStatus string) (string, error) {
	requestedStatus = strings.TrimSpace(requestedStatus)
	if !isValidLeadStatus(requestedStatus) {
		return "", errors.New("invalid lead status")
	}

	if requestedStatus == currentStatus {
		return requestedStatus, nil
	}

	nextStatuses, ok := allowedLeadStatusTransitions[currentStatus]
	if !ok {
		return "", errors.New("invalid current lead status")
	}

	if _, ok := nextStatuses[requestedStatus]; !ok {
		return "", errors.New("invalid lead status transition")
	}

	return requestedStatus, nil
}
