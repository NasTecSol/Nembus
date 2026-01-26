package main

import (
	"fmt"
	"os"

	"NEMBUS/internal/middleware"

	"github.com/joho/godotenv"
)

// This script generates a dev token for testing
// Usage: go run scripts/generate-dev-token.go
func main() {
	// Load environment variables
	if err := godotenv.Load(); err != nil {
		fmt.Println("Note: .env file not found, using system environment variables")
	}

	// Default dev user credentials
	userID := "00000000-0000-0000-0000-000000000000"
	userLogin := "dev_user"

	// Allow override via environment variables
	if envUserID := os.Getenv("DEV_USER_ID"); envUserID != "" {
		userID = envUserID
	}
	if envUserLogin := os.Getenv("DEV_USER_LOGIN"); envUserLogin != "" {
		userLogin = envUserLogin
	}

	// Generate token
	token, err := middleware.GenerateDevToken(userID, userLogin)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error generating token: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("Dev Token Generated:")
	fmt.Println("===================")
	fmt.Printf("Token: %s\n", token)
	fmt.Printf("User ID: %s\n", userID)
	fmt.Printf("User Login: %s\n", userLogin)
	fmt.Println("\nUsage:")
	fmt.Println("  curl -H \"Authorization: Bearer " + token + "\" -H \"x-tenant-id: <tenant-slug>\" http://localhost:8080/api/employees")
}
