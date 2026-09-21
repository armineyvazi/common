package ports_test

import (
	"testing"

	"github.com/armineyvazi/common.git/pkg/ports"
)

func TestErrorMessages_NotEmpty(t *testing.T) {
	msgs := []struct {
		name string
		msg  ports.ErrorMessage
	}{
		{"InternalErrorMessage", ports.InternalErrorMessage},
		{"BadRequestMessage", ports.BadRequestMessage},
		{"NotFoundMessage", ports.NotFoundMessage},
		{"DuplicatedMessage", ports.DuplicatedMessage},
		{"UnprocessableMessage", ports.UnprocessableMessage},
		{"PermissionDeniedMessage", ports.PermissionDeniedMessage},
		{"UnauthenticatedMessage", ports.UnauthenticatedMessage},
		{"UnauthorizedMessage", ports.UnauthorizedMessage},
		{"InvalidArgumentMessage", ports.InvalidArgumentMessage},
		{"TypeTooManyMessage", ports.TypeTooManyMessage},
	}
	for _, tt := range msgs {
		t.Run(tt.name, func(t *testing.T) {
			if tt.msg == "" {
				t.Error("message must not be empty")
			}
		})
	}
}

func TestErrorMessagesMap_CoversAllTypes(t *testing.T) {
	required := []ports.ErrorType{
		ports.TypeValidation,
		ports.TypeNotFound,
		ports.TypeForbidden,
		ports.TypeUnAuthorized,
		ports.TypeInternal,
		ports.TypeDuplicate,
		ports.TypeTooManyRequests,
		ports.TypeUnprocessable,
	}
	for _, typ := range required {
		if _, ok := ports.ErrorMessagesMap[typ]; !ok {
			t.Errorf("ErrorMessagesMap missing entry for type %q", typ)
		}
	}
}

func TestErrorMessages_AreASCII(t *testing.T) {
	for typ, msg := range ports.ErrorMessagesMap {
		for _, r := range msg {
			if r > 127 {
				t.Errorf("ErrorMessagesMap[%q] contains non-ASCII character %q", typ, r)
			}
		}
	}
}

func TestErrorCodes_AllPositive(t *testing.T) {
	codes := []struct {
		name string
		code int
	}{
		{"InternalErrorCode", ports.InternalErrorCode},
		{"BadRequestErrorCode", ports.BadRequestErrorCode},
		{"NotFoundErrorCode", ports.NotFoundErrorCode},
		{"PermissionDeniedErrorCode", ports.PermissionDeniedErrorCode},
		{"UnauthenticatedErrorCode", ports.UnauthenticatedErrorCode},
		{"UnauthorizedErrorCode", ports.UnauthorizedErrorCode},
		{"TooManyRequestsErrorCode", ports.TooManyRequestsErrorCode},
		{"MethodNotAllowedErrorCode", ports.MethodNotAllowedErrorCode},
	}
	for _, tt := range codes {
		t.Run(tt.name, func(t *testing.T) {
			if tt.code <= 0 {
				t.Errorf("error code %q must be positive, got %d", tt.name, tt.code)
			}
		})
	}
}
