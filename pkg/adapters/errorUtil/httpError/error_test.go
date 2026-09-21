package httpError

import (
	"encoding/json"
	"errors"
	"testing"

	"github.com/armineyvazi/common.git/pkg/adapters/errorUtil/appErr"
	"github.com/armineyvazi/common.git/pkg/ports"
)

func TestMapToHttpErr_UnknownError(t *testing.T) {
	httpErr := MapToHttpErr(errors.New("plain error"))
	if httpErr.GetHttpStatus() != ports.StatusInternalServerError {
		t.Errorf("status: got %d, want %d", httpErr.GetHttpStatus(), ports.StatusInternalServerError)
	}
}

func TestMapToHttpErr_AppError(t *testing.T) {
	tests := []struct {
		name       string
		err        ports.AppError
		wantStatus int
	}{
		{"not found", appErr.NewNotFoundErr(nil), ports.StatusNotFoundError},
		{"unauthorized", appErr.NewUnauthenticatedErr(nil), ports.StatusAuthorizationError},
		{"permission denied", appErr.NewPermissionDeniedErr(nil), ports.StatusPermissionDenied},
		{"internal", appErr.NewInternalErr(nil), ports.StatusInternalServerError},
		{"invalid arg", appErr.NewInvalidArgumentErr(nil), ports.StatusUnprocessableEntity},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			httpErr := MapToHttpErr(tt.err)
			if httpErr.GetHttpStatus() != tt.wantStatus {
				t.Errorf("status: got %d, want %d", httpErr.GetHttpStatus(), tt.wantStatus)
			}
		})
	}
}

func TestNew_HttpStatusSet(t *testing.T) {
	e := New(appErr.NewInternalErr(nil), ports.StatusBadGateway)
	if e.GetHttpStatus() != ports.StatusBadGateway {
		t.Errorf("status: got %d, want %d", e.GetHttpStatus(), ports.StatusBadGateway)
	}
}

func TestMarshalJSON_IncludesCodeMessageTrackID(t *testing.T) {
	ae := appErr.NewNotFoundErr(errors.New("missing")).
		WithMessage("not found").
		WithErrorList("field", "required")

	he := New(ae, ports.StatusNotFoundError)

	b, err := json.Marshal(he)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}

	var out map[string]any
	if err := json.Unmarshal(b, &out); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	if out["code"] == nil {
		t.Error("expected 'code' in JSON output")
	}
	if out["message"] == nil {
		t.Error("expected 'message' in JSON output")
	}
	if out["status"] != false {
		t.Errorf("status: got %v, want false", out["status"])
	}
}

func TestConstructors(t *testing.T) {
	inner := errors.New("cause")
	tests := []struct {
		name       string
		err        ports.HttpError
		wantStatus int
	}{
		{"bad request", NewBadRequestErr(inner), ports.StatusBadRequest},
		{"unauthorized", NewUnauthorizedErr(inner), ports.StatusAuthorizationError},
		{"forbidden", NewForbiddenErr(inner), ports.StatusPermissionDenied},
		{"not found", NewNotFoundErr(inner), ports.StatusNotFoundError},
		{"internal", NewInternalServerErrorErr(inner), ports.StatusInternalServerError},
		{"service unavailable", NewServiceUnavailableErr(inner), ports.StatusServiceUnavailable},
		{"method not allowed", NewMethodNotAllowedErr(inner), ports.StatusMethodNotAllowed},
		{"request timeout", NewRequestTimeoutErr(inner), ports.StatusRequestTimeout},
		{"precondition failed", NewPreConditionFailedErr(inner), ports.StatusPreConditionFailed},
		{"bad gateway", NewBadGatewayErr(inner), ports.StatusBadGateway},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.err.GetHttpStatus() != tt.wantStatus {
				t.Errorf("status: got %d, want %d", tt.err.GetHttpStatus(), tt.wantStatus)
			}
		})
	}
}

func TestAddToMap(t *testing.T) {
	// Custom error code -> HTTP status mapping.
	customCode := 99999
	AddToMap(map[int]int{customCode: 418})
	if ports.AppErrorToHttpStatus[customCode] != 418 {
		t.Error("custom mapping was not added")
	}
}
