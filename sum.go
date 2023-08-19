package nuggit

import (
	"crypto/sha1"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"hash"
	"hash/crc32"
	"strings"

	"golang.org/x/exp/slices"
)

var crc32Tab = crc32.MakeTable(crc32.Castagnoli)

// Sums specifies checksums used for integrity checks
// on resources fetched from remote sources.
// The polynomial used for CRC32 is Castagnoli's polynomial.
type Sums struct {
	// CRC32 hex encoded checksum.
	CRC32 string `json:"crc32,omitempty"`
	// SHA1 hex encoded hash.
	SHA1 string `json:"sha1,omitempty"`
	// SHA2 hex encoded hash.
	SHA2 string `json:"sha2,omitempty"`
}

// Checksum creates checksums for the data.
func Checksum(data []byte) *Sums {
	var sums Sums
	var hashes [3]string
	for i, h := range []hash.Hash{
		crc32.New(crc32Tab),
		sha1.New(),
		sha256.New(),
	} {
		h.Write(data)
		hashes[i] = hex.EncodeToString(h.Sum(nil))
	}
	sums.CRC32 = hashes[0]
	sums.SHA1 = hashes[1]
	sums.SHA2 = hashes[2]
	return &sums
}

type SumTest struct {
	Sum      string
	Expected string
	Actual   string
}

func (t SumTest) Fail() bool { return t.Expected != "" && t.Expected != t.Actual }

func (t SumTest) Format(verbose bool) string {
	if !verbose {
		if t.Fail() {
			return fmt.Sprintf("%-6s FAIL", t.Sum)
		}
		return fmt.Sprintf("%-6s OK", t.Sum)
	}
	if t.Fail() {
		return fmt.Sprintf("%-6s %s != %s", t.Sum, t.Expected, t.Actual)
	}
	return fmt.Sprintf("%-6s %s", t.Sum, t.Actual)
}

func (s SumTest) String() string {
	return s.Format(false)
}

type SumTests []SumTest

func (s SumTests) Fail() bool {
	for _, t := range s {
		if t.Fail() {
			return true
		}
	}
	return false
}

// Format the response for the `nuggit sum` subcommand.
func (s SumTests) Format(verbose bool) string {
	if s == nil {
		return "PASS"
	}
	var sb strings.Builder
	tests := slices.Clone(s)
	slices.SortFunc(tests, func(a, b SumTest) int { return strings.Compare(a.Sum, b.Sum) })
	for _, test := range tests {
		fmt.Fprintln(&sb, test.Format(verbose))
	}
	return sb.String()
}

func (s SumTests) FormatError() (string, bool) {
	var failed []string
	var failDetails []string

	tests := slices.Clone(s)
	slices.SortFunc(tests, func(a, b SumTest) int { return strings.Compare(a.Sum, b.Sum) })
	for _, t := range tests {
		if t.Fail() {
			failed = append(failed, t.Sum)
			failDetails = append(failDetails, fmt.Sprintf("%s != %s", t.Expected, t.Actual))
		}
	}
	if len(failed) == 0 {
		return "", false
	}
	return fmt.Sprintf("unexpected sums for (%s): %s",
		strings.Join(failed, ", "),
		strings.Join(failDetails, ", ")), true
}

// Test the checksums sums against the other sums.
// The returned error is always of type *SumError.
func (s *Sums) Test(other *Sums) SumTests {
	if s == nil {
		return nil
	}
	var s2 Sums
	if other != nil {
		s2 = *other
	}
	if *s == s2 { // Fast check for equality.
		return nil
	}
	// Slow check all fields.
	// Only validate if Sum is nonempty in s.
	var tt SumTests

	for _, t := range []struct {
		name string
		want string
		got  string
	}{{
		name: "crc32",
		want: s.CRC32,
		got:  other.CRC32,
	}, {
		name: "sha1",
		want: s.SHA1,
		got:  other.SHA1,
	}, {
		name: "sha2",
		want: s.SHA2,
		got:  other.SHA2,
	}} {
		tt = append(tt, SumTest{
			Sum:      t.name,
			Expected: t.want,
			Actual:   t.got,
		})
	}
	return tt
}

// TestBytes tests the bytes against the checksums.
// The returned error is always of type *SumError.
func (s *Sums) TestBytes(data []byte) SumTests {
	if s == nil {
		return nil
	}
	return s.Test(Checksum(data))
}
