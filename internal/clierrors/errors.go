package clierrors

import (
	"encoding/json"
	"fmt"
	"io"
)

const (
	ExitOK           = 0
	ExitGeneric      = 1
	ExitValidation   = 2
	ExitAuth         = 3
	ExitNetwork      = 4
	ExitAPI          = 5
	ExitPolicy       = 6
	ExitRateLimit    = 7
	ExitFileIO       = 8
	ExitConfirmation = 10
)

type CLIError struct {
	Type    string `json:"type"`
	Subtype string `json:"subtype"`
	Message string `json:"message"`
	Hint    string `json:"hint,omitempty"`
	Param   string `json:"param,omitempty"`
	Code    int    `json:"code,omitempty"`
	Exit    int    `json:"-"`
}

func (e *CLIError) Error() string {
	return e.Message
}

func New(exit int, typ, subtype, message, hint string) *CLIError {
	return &CLIError{Exit: exit, Type: typ, Subtype: subtype, Message: message, Hint: hint}
}

func Validation(subtype, message, param, hint string) *CLIError {
	return &CLIError{Exit: ExitValidation, Type: "validation", Subtype: subtype, Message: message, Param: param, Hint: hint}
}

func Auth(message, hint string) *CLIError {
	return New(ExitAuth, "authentication", "api_key_missing", message, hint)
}

func Network(err error) *CLIError {
	return New(ExitNetwork, "network", "transport", fmt.Sprintf("network error: %v", err), "check network connectivity and retry")
}

func API(status int, message string) *CLIError {
	exit := ExitAPI
	typ := "api"
	subtype := "upstream_error"
	hint := "retry later or inspect the Agnes API response"
	switch status {
	case 401, 403:
		exit, typ, subtype, hint = ExitAuth, "authentication", "api_key_invalid", "check AGNES_API_KEY or run: agnes key set <api-key>"
	case 404:
		exit, typ, subtype, hint = ExitAPI, "api", "not_found", "check the task or video id"
	case 429:
		exit, typ, subtype, hint = ExitRateLimit, "api", "rate_limited", "wait and retry later"
	case 400:
		exit, typ, subtype, hint = ExitValidation, "validation", "bad_request", "check command flags or run: agnes schema <command>"
	case 503:
		exit, typ, subtype, hint = ExitAPI, "api", "service_busy", "retry later"
	}
	return &CLIError{Exit: exit, Type: typ, Subtype: subtype, Message: message, Hint: hint, Code: status}
}

func FileIO(err error) *CLIError {
	return New(ExitFileIO, "file_io", "local_file", fmt.Sprintf("file error: %v", err), "check the path and file permissions")
}

func TaskFailed(id, reason string) *CLIError {
	return New(ExitAPI, "api", "task_failed", fmt.Sprintf("video task %s failed: %s", id, reason), "")
}

func Timeout(id string) *CLIError {
	return New(ExitNetwork, "network", "timeout", fmt.Sprintf("video task %s timed out waiting for result", id), fmt.Sprintf("check later with: agnes video status %s", id))
}

type envelope struct {
	OK    bool      `json:"ok"`
	Error *CLIError `json:"error"`
}

func WriteJSON(w io.Writer, err *CLIError) {
	_ = json.NewEncoder(w).Encode(envelope{OK: false, Error: err})
}
