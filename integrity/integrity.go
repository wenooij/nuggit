package integrity

import (
	"crypto/sha1"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"regexp"
	"strings"

	"github.com/wenooij/nuggit"
	"github.com/wenooij/nuggit/status"
)

var namePattern = regexp.MustCompile(`^(?i:[a-z][a-z0-9-]*)$`)

type Name interface {
	GetName() string
}

type Digest interface {
	GetDigest() string
}

type NameDigest interface {
	Name
	Digest
}

type Spec interface {
	GetSpec() any
}

func GetDigest[E Spec](e E) (string, error) {
	h := sha1.New()
	if err := json.NewEncoder(h).Encode(e.GetSpec()); err != nil {
		return "", fmt.Errorf("digest failed: %w", err)
	}
	digest := h.Sum(nil)
	return hex.EncodeToString(digest), nil
}

type Digestable interface {
	Spec
	SetDigest(string)
}

func SetDigest[E Digestable](e E) error {
	digest, err := GetDigest(e)
	if err != nil {
		return err
	}
	e.SetDigest(digest)
	return nil
}

type Nameable interface {
	SetName(string)
}

func SetName[E Nameable](e E, name string) { e.SetName(name) }

type NameDigestable interface {
	Nameable
	Digestable
}

func SetNameDigest[E NameDigestable](e E, name string) error {
	SetName(e, name)
	SetDigest(e)
	return nil
}

func HasName(n Name) bool     { return n.GetName() != "" }
func HasDigest(d Digest) bool { return d.GetDigest() != "" }

// FormatString formats the entity as a valid "name@digest".
func FormatString(nd NameDigest) (string, error) {
	if err := validateName(nd.GetName()); err != nil {
		return "", err
	}
	if !HasDigest(nd) {
		return nd.GetName(), nil
	}
	if err := validateHexDigest(nd.GetDigest()); err != nil {
		return "", err
	}
	return fmt.Sprintf("%s@%s", nd.GetName(), nd.GetDigest()), nil
}

func CompareNameDigest(a, b NameDigest) int {
	if cmp := strings.Compare(a.GetDigest(), b.GetDigest()); cmp != 0 {
		return cmp
	}
	return strings.Compare(a.GetName(), b.GetName())
}

func ParseNameDigest(s string) (nameDigest NameDigest, err error) {
	if len(s) == 0 {
		return nil, fmt.Errorf("name@digest must not be empty: %w", status.ErrInvalidArgument)
	}
	name, digest, _ := strings.Cut(s, "@")
	if err := validateName(name); err != nil {
		return nil, err
	}
	if err := validateHexDigest(digest); err != nil {
		return nil, err
	}
	return KeyLit(name, digest), nil
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

type CheckDigestable interface {
	NameDigestable
	Digest
}

func SetCheckDigest[E CheckDigestable](e E, digest string) error {
	if err := SetDigest(e); err != nil {
		return err
	}
	if digest != "" && e.GetDigest() != digest {
		return fmt.Errorf("integrity check failed (%q != %q)", e.GetDigest(), digest)
	}
	return nil
}

func CheckDigest[E Digest](e E, digest string) error {
	if e.GetDigest() != digest {
		return fmt.Errorf("integrity check failed (%q != %q)", e.GetDigest(), digest)
	}
	return nil
}

func GetNameDigestArg(a nuggit.Action) NameDigest {
	return KeyLit(a.GetOrDefaultArg("name"), a.GetOrDefaultArg("digest"))
}

type key struct {
	Name   string
	Digest string
}

func (k key) GetName() string   { return k.Name }
func (k key) GetDigest() string { return k.Digest }
func (k key) String() string {
	s, err := FormatString(k)
	if err != nil {
		// Fallback format when name or digest is invalid.
		return fmt.Sprintf("!(%q, %q)", k.Name, k.Digest)
	}
	return s
}

func Key[E NameDigest](e E) NameDigest      { return key{e.GetName(), e.GetDigest()} }
func KeyLit(name, digest string) NameDigest { return key{name, digest} }
