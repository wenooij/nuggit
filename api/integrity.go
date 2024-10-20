package api

import (
	"crypto/sha1"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"regexp"
	"strings"

	"github.com/wenooij/nuggit/status"
)

var namePattern = regexp.MustCompile(`^(?i:[a-z][a-z0-9-]*)$`)

type NameDigest struct {
	Name   string `json:"name,omitempty"`
	Digest string `json:"digest,omitempty"`
}

func (d *NameDigest) GetName() string {
	if d == nil {
		return ""
	}
	return d.Name
}

func (d *NameDigest) GetDigest() string {
	if d == nil {
		return ""
	}
	return d.Digest
}

func (d *NameDigest) HasDigest() bool {
	if d == nil {
		return false
	}
	return d.Digest != ""
}

func (d *NameDigest) Equal(d2 *NameDigest) bool {
	return (d == nil && d2 == nil) ||
		(d != nil && d2 != nil && *d == *d2)
}

func (d *NameDigest) String() string {
	var sb strings.Builder
	sb.Grow(len(d.GetName()) + 1 + len(d.GetDigest()))
	sb.WriteString(d.GetName())
	if d.HasDigest() {
		sb.WriteByte('@')
		sb.WriteString(d.GetDigest())
	}
	return sb.String()
}

type Named interface {
	GetName() string
}

func ParseNameDigest(nameDigest string) (*NameDigest, error) {
	if len(nameDigest) == 0 {
		return nil, fmt.Errorf("name@digest must not be empty: %w", status.ErrInvalidArgument)
	}
	name, digest, foundDigest := strings.Cut(nameDigest, "@")
	if err := validateName(name); err != nil {
		return nil, err
	}
	if foundDigest {
		if err := validateHexDigest(digest); err != nil {
			return nil, err
		}
	}
	return &NameDigest{Name: name, Digest: digest}, nil
}

func digestSHA1[E any](e E) (string, error) {
	h := sha1.New()
	data, err := json.Marshal(e)
	if err != nil {
		return "", fmt.Errorf("digest failed: %v: %w", err, status.ErrInvalidArgument)
	}
	if _, err := h.Write(data); err != nil {
		return "", fmt.Errorf("digest failed: %w", err)
	}
	digest := h.Sum(nil)
	return hex.EncodeToString(digest), nil
}

func NewNameDigest[E Named](e E) (*NameDigest, error) {
	name := e.GetName()
	if err := validateName(name); err != nil {
		return nil, err
	}
	digest, err := digestSHA1(e)
	if err != nil {
		return nil, err
	}
	return &NameDigest{Name: name, Digest: digest}, nil
}

func validateName(name string) error {
	if name == "" {
		return fmt.Errorf("name must not be empty: %w", status.ErrInvalidArgument)
	}
	if !namePattern.MatchString(name) {
		return fmt.Errorf("name contains invalid characters (%q): %w", name, status.ErrInvalidArgument)
	}
	return nil
}

func validateHexDigest(hexStr string) error {
	if len(hexStr) == 0 {
		return fmt.Errorf("digest must not be empty: %w", status.ErrInvalidArgument)
	}
	for _, b := range hexStr {
		switch {
		case b >= '0' && b <= '9' || b >= 'A' && b <= 'F' || b >= 'a' && b <= 'f':
		default:
			return fmt.Errorf("digest is not hex encoded (%q): %v", hexStr, status.ErrInvalidArgument)
		}
	}
	return nil
}

func CheckIntegrity[E Named](nameDigests []string, objects []E) error {
	if len(objects) == len(nameDigests) {
		return fmt.Errorf("integrity check failed: mismatched numbers of digests and objects (got %d, wanted %d): %w", len(objects), len(nameDigests), status.ErrInvalidArgument)
	}
	for i, want := range nameDigests {
		obj := objects[i]
		nameDigest, err := NewNameDigest(obj)
		if err != nil {
			return fmt.Errorf("failed to digest object (#%d): %v: %w", i, err, status.ErrInvalidArgument)
		}
		if got := nameDigest.String(); got != want {
			return fmt.Errorf("integrity check failed (#%d; got %q, want %q): %w", i, got, want, status.ErrInvalidArgument)
		}
	}
	return nil
}
