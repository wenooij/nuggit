package status

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
)

var (
	ErrCanceled           = errors.New("canceled")
	ErrUnknown            = errors.New("unknown")
	ErrInvalidArgument    = errors.New("invalid argument")
	ErrDeadlineExceeded   = errors.New("deadline exceeded")
	ErrNotFound           = errors.New("not found")
	ErrAlreadyExists      = errors.New("already exists")
	ErrPermissionDenied   = errors.New("permission denied")
	ErrResourceExhausted  = errors.New("resource exhausted")
	ErrFailedPrecondition = errors.New("failed precondition")
	ErrAborted            = errors.New("aborted")
	ErrOutOfRange         = errors.New("out of range")
	ErrUnimplemented      = errors.New("unimplemented")
	ErrInternal           = errors.New("internal")
	ErrUnavailable        = errors.New("unavailable")
	ErrDataLoss           = errors.New("data loss")
	ErrUnauthenticated    = errors.New("unauthenticated")
)

var statusStr = map[string]error{
	"canceled":            ErrCanceled,
	"unknown":             ErrUnknown,
	"invalid argument":    ErrInvalidArgument,
	"deadline exceeded":   ErrDeadlineExceeded,
	"not found":           ErrNotFound,
	"already exists":      ErrAlreadyExists,
	"permission denied":   ErrPermissionDenied,
	"resource exhausted":  ErrResourceExhausted,
	"failed precondition": ErrFailedPrecondition,
	"aborted":             ErrAborted,
	"out of range":        ErrOutOfRange,
	"unimplemented":       ErrUnimplemented,
	"internal":            ErrInternal,
	"unavailable":         ErrUnavailable,
	"data loss":           ErrDataLoss,
	"unauthenticated":     ErrUnauthenticated,
}

var statusHTTP = map[error]int{
	nil:                   http.StatusOK,
	ErrCanceled:           499, // Client Closed Request
	ErrUnknown:            http.StatusInternalServerError,
	ErrInvalidArgument:    http.StatusBadRequest,
	ErrDeadlineExceeded:   http.StatusGatewayTimeout,
	ErrNotFound:           http.StatusNotFound,
	ErrAlreadyExists:      http.StatusConflict,
	ErrPermissionDenied:   http.StatusForbidden,
	ErrResourceExhausted:  http.StatusInsufficientStorage,
	ErrFailedPrecondition: http.StatusPreconditionFailed,
	ErrAborted:            http.StatusConflict,
	ErrOutOfRange:         http.StatusRequestedRangeNotSatisfiable,
	ErrUnimplemented:      http.StatusNotImplemented,
	ErrInternal:           http.StatusInternalServerError,
	ErrUnavailable:        http.StatusServiceUnavailable,
	ErrDataLoss:           http.StatusInternalServerError,
	ErrUnauthenticated:    http.StatusUnauthorized,
}

type apiError struct {
	Status error `json:"status,omitempty"`
	Reason error `json:"reason,omitempty"`
}

func (e *apiError) MarshalJSON() ([]byte, error) {
	apiErr := new(struct {
		Status string `json:"status,omitempty"`
		Reason string `json:"reason,omitempty"`
	})
	if e != nil {
		*apiErr = struct {
			Status string `json:"status,omitempty"`
			Reason string `json:"reason,omitempty"`
		}{}
		if e.Status != nil {
			apiErr.Status = e.Status.Error()
		}
		if e.Reason != nil {
			apiErr.Reason = e.Reason.Error()
		}
	}
	return json.Marshal(apiErr)
}

func (e *apiError) UnmarshalJSON(data []byte) error {
	apiErr := new(struct {
		Status string `json:"status,omitempty"`
		Reason string `json:"reason,omitempty"`
	})
	if err := json.Unmarshal(data, &apiErr); err != nil {
		return err
	}
	if apiErr == nil {
		return nil
	}
	status := statusStr[apiErr.Status]
	if status == nil {
		status = ErrUnknown
	}
	e.Status = status
	e.Reason = fmt.Errorf(apiErr.Reason)
	return nil
}

func makeAPIError(err error) *apiError {
	if err == nil {
		return nil
	}
	if apiErr, ok := err.(*apiError); ok {
		return apiErr
	}
	for status := range statusHTTP {
		if errors.Is(err, status) {
			return &apiError{Status: status, Reason: errors.Unwrap(err)}
		}
	}
	return &apiError{Status: ErrUnknown, Reason: err}
}

func (e apiError) Error() string     { return fmt.Sprintf("%v: %v", e.Reason, e.Status) }
func (e apiError) Is(err error) bool { return e.Status == err || errors.Is(e.Reason, err) }
func (e apiError) Unwrap() error     { return e.Reason }

const failedMarshalResource = `{"status":"internal","reason":"failed to marshal resource"}`

const marshalIndent = true

func WriteError(w http.ResponseWriter, err error) {
	apiErr := makeAPIError(err)
	if apiErr != nil {
		w.WriteHeader(statusHTTP[apiErr.Status])
	}
	writeJSON(w, apiErr)
}

func WriteResponse[E any](w http.ResponseWriter, e E, err error) {
	if err != nil {
		WriteError(w, err)
		return
	}
	writeJSON(w, e)
}

func writeJSON[E any](w http.ResponseWriter, e E) {
	var bs []byte
	var err error
	if marshalIndent {
		bs, err = json.MarshalIndent(e, "", "  ")
	} else {
		bs, err = json.Marshal(e)
	}
	if err != nil {
		log.Printf("Failed to marshal resource: [%T]", any(e))
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(failedMarshalResource))
		return
	}
	w.Write(bs)
}
