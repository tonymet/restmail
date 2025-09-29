package main

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
)

// decode strings from golang http debug
func main() {
	reader := bufio.NewReader(os.Stdin)
	fmt.Println("Enter the hex-escaped string (e.g., \\x02\\x48\\x65\\x6c\\x6c\\x6f):")
	escapedInput, err := reader.ReadString('\n')
	if err != nil && err.Error() != "EOF" {
		fmt.Fprintf(os.Stderr, "Error reading input: %v\n", err)
		return
	}
	cleanedInput := strings.TrimSpace(escapedInput)
	if cleanedInput == "" {
		fmt.Println("No input provided. Exiting.")
		return
	}
	quotedLiteral := fmt.Sprintf("\"%s\"", cleanedInput)
	decodedString, err := strconv.Unquote(quotedLiteral)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error decoding string (Input might be invalid): %v\n", err)
		return
	}
	fmt.Println("---")
	fmt.Printf("Original Escaped Input (String Literal): %s\n", cleanedInput)
	fmt.Printf("Decoded Output (Normal String): %s\n", decodedString)
	fmt.Printf("Decoded Output (Hex Bytes): ")
	for _, b := range []byte(decodedString) {
		fmt.Printf("\\x%02x", b)
	}
	fmt.Println()
	fmt.Println("---")
}
