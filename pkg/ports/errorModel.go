package ports

import "context"

// ErrorMessagesMap Define error messages map
var ErrorMessagesMap = map[ErrorType]ErrorMessage{
	TypeValidation:      InvalidArgumentMessage,
	TypeNotFound:        NotFoundMessage,
	TypeForbidden:       PermissionDeniedMessage,
	TypeUnAuthorized:    UnauthorizedMessage,
	TypeInternal:        InternalErrorMessage,
	TypeDuplicate:       DuplicatedMessage,
	TypeTooManyRequests: TypeTooManyMessage,
	TypeUnprocessable:   UnprocessableMessage,
}

const (
	InternalErrorMessage    ErrorMessage = "خطای سیستم"
	BadRequestMessage       ErrorMessage = "خطای ورودی"
	NotFoundMessage         ErrorMessage = "پیدا نشد"
	DuplicatedMessage       ErrorMessage = "داده تکراری‌ است"
	UnprocessableMessage    ErrorMessage = "قابل پردازش نیست"
	PermissionDeniedMessage ErrorMessage = "دسترسی غیر مجاز"
	UnauthenticatedMessage  ErrorMessage = "کاربر احراز هویت نشده است"
	UnauthorizedMessage     ErrorMessage = "عدم دسترسی مجاز"
	InvalidArgumentMessage  ErrorMessage = "داده معتبر نیست"
	TypeTooManyMessage      ErrorMessage = "تعداد درخواست‌ها بیش از حد مجاز است"
)

type ErrorMessage = string

// ErrorType determine category of Error
type ErrorType string

const (
	TypeValidation       ErrorType = "VALIDATION"
	TypeNotFound         ErrorType = "NOT_FOUND"
	TypeDuplicate        ErrorType = "DUPLICATED"
	TypeUnAuthorized     ErrorType = "UNAUTHORIZED"
	TypeForbidden        ErrorType = "FORBIDDEN"
	TypeInternal         ErrorType = "INTERNAL"
	TypeTooManyRequests  ErrorType = "TOO_MANY_REQUESTS"
	TypeUnprocessable    ErrorType = "UNPROCESSABLE"
	TypeMethodNotAllowed ErrorType = "METHOD_NOT_ALLOWED"
)

// ErrorCode determine more specific info about Error
type ErrorCode = int

const (
	// InternalErrorCode Internal errors.  This means that some invariants expected by the
	// underlying system have been broken.  This error code is reserved
	// for serious errors.
	InternalErrorCode ErrorCode = 10001
	// BadRequestErrorCode The client specified an invalid argument.  `INVALID_ARGUMENT` indicates
	// arguments that are problematic regardless of the state of the system
	// (e.g., a malformed file name).
	BadRequestErrorCode ErrorCode = 10002
	// NotFoundErrorCode Some requested entity (e.g., file or directory) was not found.
	// Note to server developers: if a request is denied for an entire class
	// of users, such as gradual feature rollout or undocumented whitelist,
	// `NOT_FOUND` may be used. If a request is denied for some users within
	// a class of users, such as user-based access control, `PERMISSION_DENIED`
	// must be used.	NotFoundErrorCode          ErrorCode = 10003
	NotFoundErrorCode ErrorCode = 10003
	// RetriableInternalErrorCode The service is currently unavailable.  This is a most likely a
	// transient condition and may be corrected by retrying with
	// a backoff.
	RetriableInternalErrorCode ErrorCode = 10004
	// CancelledErrorCode The operation was cancelled, typically by the caller.
	CancelledErrorCode ErrorCode = 10005
	// UnknownErrorCode Unknown error.  For example, this error may be returned when
	// a `Status` value received from another address space belongs to
	// an error space that is not known in this address space.  Also
	// errors raised by APIs that do not return enough error information
	// may be converted to this error.
	UnknownErrorCode ErrorCode = 10006
	// InvalidArgumentErrorCode The client specified an invalid argument.  Note that this differs
	// from `FailedPrecondition`.  `InvalidArgument` indicates arguments
	// that are problematic regardless of the state of the system
	// (e.g., a malformed file name).
	InvalidArgumentErrorCode ErrorCode = 10007
	// DeadlineExceededErrorCode The deadline expired before the operation could complete. For operations
	// that change the state of the system, this error may be returned
	// even if the operation has completed successfully. For example, a
	// successful response from a server could have been delayed long
	// enough for the deadline to expire.
	DeadlineExceededErrorCode ErrorCode = 10008
	// AlreadyExistsErrorCode The entity that a client attempted to create (e.g., file or directory)
	// already exists.
	AlreadyExistsErrorCode ErrorCode = 10010
	// PermissionDeniedErrorCode The caller does not have permission to execute the specified operation.
	// `PERMISSION_DENIED` must not be used for rejections caused by exhausting
	// some resource (use `RESOURCE_EXHAUSTED` instead for those errors).
	// `PERMISSION_DENIED` must not be used if the caller can not be identified
	// (use `UNAUTHENTICATED` instead for those errors).
	PermissionDeniedErrorCode ErrorCode = 10011
	// ResourceExhaustedErrorCode Some resource has been exhausted, perhaps a per-user quota, or
	// perhaps the entire file system is out of space.
	ResourceExhaustedErrorCode ErrorCode = 10012
	// FailedPreconditionErrorCode The operation was rejected because the system is not in a state
	// required for the operation's execution.  For example, the directory
	// to be deleted is non-empty, an rmdir operation is applied to
	// a non-directory, etc.
	//
	// Service implementors can use the following guidelines to decide
	// between `FAILED_PRECONDITION`, `ABORTED`, and `UNAVAILABLE`:
	//  (a) Use `UNAVAILABLE` if the client can retry just the failing call.
	//  (b) Use `ABORTED` if the client should retry at a higher level
	//      (e.g., when a client-specified test-and-set fails, indicating the
	//      client should restart a read-modify-write sequence).
	//  (c) Use `FAILED_PRECONDITION` if the client should not retry until
	//      the system state has been explicitly fixed.  E.g., if an "rmdir"
	//      fails because the directory is non-empty, `FAILED_PRECONDITION`
	//      should be returned since the client should not retry unless
	//      the files are deleted from the directory.
	FailedPreconditionErrorCode ErrorCode = 10013
	// AbortedErrorCode The operation was aborted, typically due to a concurrency issue such as
	// a sequencer check failure or transaction abort.
	//
	// See the guidelines above for deciding between `FAILED_PRECONDITION`,
	// `ABORTED`, and `UNAVAILABLE`.
	AbortedErrorCode ErrorCode = 10014
	// OutOfRangeErrorCode The operation was attempted past the valid range.  E.g., seeking or
	// reading past end-of-file.
	//
	// Unlike `INVALID_ARGUMENT`, this error indicates a problem that may
	// be fixed if the system state changes. For example, a 32-bit file
	// system will generate `INVALID_ARGUMENT` if asked to read at an
	// offset that is not in the range [0,2^32-1], but it will generate
	// `OUT_OF_RANGE` if asked to read from an offset past the current
	// file size.
	//
	// There is a fair bit of overlap between `FAILED_PRECONDITION`,
	// `OUT_OF_RANGE`, and `INVALID_ARGUMENT`.  We recommend using
	// `OUT_OF_RANGE` (the most specific) when it applies so that callers
	// who are iterating through a space can easily look for an
	// `OUT_OF_RANGE` error to detect when they are done.
	OutOfRangeErrorCode ErrorCode = 10015
	// UnimplementedErrorCode The operation is not implemented or is not supported/enabled in this service.
	UnimplementedErrorCode ErrorCode = 10016
	// UnavailableErrorCode The service is currently unavailable.  This is a most likely a
	// transient condition and may be corrected by retrying with
	// a backoff.
	//
	// See the guidelines above for deciding between `FAILED_PRECONDITION`,
	// `ABORTED`, and `UNAVAILABLE`.
	UnavailableErrorCode ErrorCode = 10018
	// DataLossErrorCode Unrecoverable data loss or corruption.
	DataLossErrorCode ErrorCode = 10019
	// UnauthenticatedErrorCode The request does not have valid authentication credentials for the
	// operation.
	UnauthenticatedErrorCode ErrorCode = 10020
	// UnauthorizedErrorCode The request does not have valid authentication credentials for the
	// operation.
	UnauthorizedErrorCode ErrorCode = 10021
	// CanceledErrorCode The operation was cancelled (typically by the caller).
	CanceledErrorCode ErrorCode = 10022

	// TooManyRequestsErrorCode sent too many requests in a given amount of time
	// or one API Request
	TooManyRequestsErrorCode ErrorCode = 10023

	// FailedDependencyErrorCode could not be performed on the resource because the requested
	// action depended on another action, and that action failed
	FailedDependencyErrorCode ErrorCode = 10024

	MethodNotAllowedErrorCode ErrorCode = 10025
)

type ErrorFieldName = string
type ErrorList []string
type AppError interface {
	error
	GetCode() int
	GetError() error
	GetMessage() ErrorMessage
	GetTrackId() string
	GetType() ErrorType
	GetDetail() string
	GetExtraInfo() map[string]interface{}
	GetNestedError() error
	GetErrorLists() map[ErrorFieldName]ErrorList
	IsDebugDisabled() bool

	WithError(e error) AppError
	WithMessage(message string) AppError
	WithDetail(message string) AppError
	WithType(errorType ErrorType) AppError
	WithCode(code int) AppError
	WithTrackId(ctx context.Context) AppError
	WithExtraInfo(map[string]interface{}) AppError
	WithNestedError(err error) AppError
	WithDisableDebug() AppError
	WithErrorList(fieldName, errorMessage string) AppError

	Is(errType ErrorType) bool
}
