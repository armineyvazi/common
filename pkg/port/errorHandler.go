package port

type ErrorHandler interface {
	CaptureError(msg string)
	CaptureException(exception error)
}

type ErrorDetails struct {
	HttpStatus  HttpStatus   `json:"-"`
	Status      bool         `json:"status"`
	Message     string       `json:"message,omitempty"`
	Code        int          `json:"code,omitempty"`
	TrackId     int          `json:"track_id,omitempty"`
	Errors      []FieldError `json:"errors,omitempty"`
	NestedError error        `json:"error,omitempty"`
}

func (e ErrorDetails) Error() string {
	return e.Message
}

type FieldError struct {
	FieldName    string      `json:"field_name"`
	CurrentValue interface{} `json:"current_value"`
	Errors       string      `json:"errors"`
}
