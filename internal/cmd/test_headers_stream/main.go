package main

import "fmt"

// Test program outputting headers (like those produced by bts/reportbug) followed by long-line body text.
func main() {
	headers := "From: anthony.metzidis@gmail.com\n" +
		"To: anthony.metzidis+103@gmail.com\n" +
		"Subject: Bug report with long lines test\n" +
		"X-Debian-PR-Package: restmail\n" +
		"X-Custom-Test-Header: custom-value-preserved\n\n"

	body := "Package: restmail\n" +
		"Version: 1.0.7\n" +
		"Severity: normal\n\n" +
		"The quick brown fox jumps over the lazy dog repeatedly until the total line length exceeds eighty characters easily.\n" +
		"A standardized test string containing sufficient information to guarantee length well over eighty characters in Go.\n" +
		"Software engineering principles often emphasize readability, yet long lines occur in documentation and error definitions.\n" +
		"When transmitting email payloads via MIME, quoted-printable encoding helps manage long lines and arbitrary ASCII bytes.\n" +
		"Another arbitrarily elongated sentence structured to verify wrapping and boundary handling in mail transport agents.\n" +
		"Deterministic test fixtures are essential when verifying edge cases involving boundary conditions and line splitting.\n"

	fmt.Print(headers + body)
}
