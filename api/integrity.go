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

func compareNameDigestPtr(a, b *NameDigest) int {
	if a == b {
		return 0
	}
	if a == nil {
		return -1
	}
	if b == nil {
		return +1
	}
	return compareNameDigest(*a, *b)
}

func compareNameDigest(a, b NameDigest) int {
	if cmp := strings.Compare(a.Digest, b.Digest); cmp != 0 {
		return cmp
	}
	return strings.Compare(a.Name, b.Name)
}

func ParseNameDigest(s string) (NameDigest, error) {
	if len(s) == 0 {
		return NameDigest{}, fmt.Errorf("name@digest must not be empty: %w", status.ErrInvalidArgument)
	}
	name, digest, _ := strings.Cut(s, "@")
	nameDigest := NameDigest{name, digest}
	if err := ValidateNameDigest(nameDigest); err != nil {
		return NameDigest{}, err
	}
	return nameDigest, nil
}

func ValidateNameDigest(nameDigest NameDigest) error {
	if err := validateName(nameDigest.Name); err != nil {
		return err
	}
	if nameDigest.HasDigest() {
		if err := validateHexDigest(nameDigest.Digest); err != nil {
			return err
		}
	}
	return nil
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

func NewNameDigest[E interface{ GetName() string }](e E) (NameDigest, error) {
	name := e.GetName()
	if err := validateName(name); err != nil {
		return NameDigest{}, err
	}
	digest, err := digestSHA1(e)
	if err != nil {
		return NameDigest{}, err
	}
	return NameDigest{Name: name, Digest: digest}, nil
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

func CheckIntegrity[E interface{ GetName() string }](nameDigests []NameDigest, objects []E) error {
	if len(objects) != len(nameDigests) {
		return fmt.Errorf("integrity check failed: mismatched numbers of digests and objects (got %d, wanted %d): %w", len(objects), len(nameDigests), status.ErrInvalidArgument)
	}
	for i, want := range nameDigests {
		obj := objects[i]
		nameDigest, err := NewNameDigest(obj)
		if err != nil {
			return fmt.Errorf("failed to digest object (#%d): %v: %w", i, err, status.ErrInvalidArgument)
		}
		if got := nameDigest; got != want {
			return fmt.Errorf("integrity check failed (#%d; got %q, want %q): %w", i, got, want, status.ErrInvalidArgument)
		}
	}
	return nil
}

func CheckIntegrityObject[E interface{ GetName() string }](nameDigests map[NameDigest]struct{}, object E) error {
	nameDigest, err := NewNameDigest(object)
	if err != nil {
		return fmt.Errorf("failed to digest object: %v: %w", err, status.ErrInvalidArgument)
	}
	if _, found := nameDigests[nameDigest]; !found {
		return fmt.Errorf("integrity check failed (unexpected digest %q): %w", nameDigest, status.ErrInvalidArgument)
	}
	return nil
}
