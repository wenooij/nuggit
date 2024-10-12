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

func ToHTTP(err error) int {
	statusErr, ok := err.(Error)
	if !ok {
		return http.StatusInternalServerError
	}
	code, ok := statusHTTP[statusErr.Status]
	if !ok {
		return http.StatusInternalServerError
	}
	return code
}

type Error struct {
	Status error `json:"status,omitempty"`
	Reason error `json:"reason,omitempty"`
}

func (e Error) Error() string     { return fmt.Sprintf("%v: %v", e.Reason, e.Status) }
func (e Error) Is(err error) bool { return e.Status == err || errors.Is(e.Reason, err) }
func (e Error) Unwrap() error     { return e.Reason }

const failedMarshalResource = `{"status":"internal","reason":"failed to marshal resource"}`

const marshalIndent = true

func WriteError(w http.ResponseWriter, err error) {
	if err == nil {
		err = Error{}
	} else if _, ok := err.(Error); !ok {
		err = Error{Status: ErrUnknown, Reason: err}
		log.Printf("Error: %v", err)
	}
	writeJSON(w, err)
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
