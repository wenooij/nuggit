package v1alpha

import "time"

type Time struct {
	Op      TimeOp       `json:"op,omitempty"`
	Sink    *Sink        `json:"sink,omitempty"`
	Layout  string       `json:"layout,omitempty"`
	Year    int          `json:"year,omitempty"`
	Month   time.Month   `json:"month,omitempty"`
	Day     int          `json:"day,omitempty"`
	Hour    int          `json:"hour,omitempty"`
	Min     int          `json:"min,omitempty"`
	Sec     int          `json:"sec,omitempty"`
	Nsec    int          `json:"nsec,omitempty"`
	Loc     string       `json:"loc,omitempty"`
	Weekday time.Weekday `json:"weekday,omitempty"`
	Time    *time.Time   `json:"time_value,omitempty"`
}

type TimeOp string

const (
	TimeUndefined TimeOp = ""
	TimeNow       TimeOp = "now"
)
