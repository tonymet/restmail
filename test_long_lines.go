package main

import (
	"fmt"
	"strings"
)

// This file contains lines strictly longer than 80 characters for testing quoted-printable and line folding behavior.
func main() {
	longLineOne := "The quick brown fox jumps over the lazy dog repeatedly until the total line length exceeds eighty characters easily."
	longLineTwo := "A standardized test string containing sufficient information to guarantee length well over eighty characters in Go."
	longLineThree := "Software engineering principles often emphasize readability, yet long lines occur in documentation and error definitions."
	longLineFour := "When transmitting email payloads via MIME, quoted-printable encoding helps manage long lines and arbitrary ASCII bytes."
	longLineFive := "Another arbitrarily elongated sentence structured to verify wrapping and boundary handling in mail transport agents."
	longLineSix := "Deterministic test fixtures are essential when verifying edge cases involving boundary conditions and line splitting."
	longLineSeven := "RFC 2045 specifies the MIME format including Quoted-Printable Content-Transfer-Encoding and its line length limits."
	longLineEight := "Lines of text in internet mail messages should not exceed 78 characters according to RFC 5322 recommendations."
	combined := strings.Join([]string{longLineOne, longLineTwo, longLineThree, longLineFour, longLineFive, longLineSix, longLineSeven, longLineEight}, "\n")
	fmt.Printf("Generated test payload containing %d bytes of long-line text across multiple lines for manual verification\n", len(combined))
	fmt.Println(combined)
}
