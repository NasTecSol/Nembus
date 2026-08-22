package main

import (
	"fmt"
	"os"

	"golang.org/x/crypto/bcrypt"
)

// This narrow helper exists solely because the established disposable E2E
// authentication setup uses bcrypt passwords. It reads its one-time value from
// process memory and writes only the bcrypt hash to stdout.
func main() {
	password := os.Getenv("NEMBUS_APPLY_DIAGNOSTIC_PASSWORD")
	if password == "" {
		panic("NEMBUS_APPLY_DIAGNOSTIC_PASSWORD is required")
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		panic(err)
	}
	fmt.Print(string(hash))
}
