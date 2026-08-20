package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/joho/godotenv"
)

type M2MClient struct {
	ClientID       string   `json:"client_id"`
	ClientName     string   `json:"client_name"`
	TenantSlug     string   `json:"tenant_slug"`
	TenantID       string   `json:"tenant_id,omitempty"`
	OrganizationID int32    `json:"organization_id"`
	Scopes         []string `json:"scopes"`
	IsActive       bool     `json:"is_active"`
	Token          string   `json:"token"`
}

type M2MRegistry struct {
	Clients []M2MClient `json:"clients"`
}

func main() {
	// Define command line flags
	clientID := flag.String("client-id", "", "Unique identifier for the client app (required)")
	clientName := flag.String("client-name", "", "Friendly name for the client app (required)")
	tenantSlug := flag.String("tenant-slug", "", "Tenant slug the client is bound to (required)")
	organizationID := flag.Int("organization-id", 0, "Tenant-local organization ID the client is bound to (required)")
	scopesStr := flag.String("scopes", "", "Comma-separated list of scopes, e.g. products:read,orders:write")
	durationYears := flag.Int("years", 5, "Token validity in years")
	flag.Parse()

	// Validate inputs
	if *clientID == "" || *clientName == "" || *tenantSlug == "" || *organizationID <= 0 {
		fmt.Println("Usage: go run cmd/m2m-gen/main.go -client-id <id> -client-name <name> -tenant-slug <slug> -organization-id <id> [-scopes <scopes>] [-years <duration>]")
		flag.PrintDefaults()
		os.Exit(1)
	}

	// Load environment variables (to get JWT_SECRET)
	// Try environment-specific files first, then fallback to .env
	_ = godotenv.Load(".env.dev")
	_ = godotenv.Load()

	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		log.Fatalf("Error: JWT_SECRET environment variable is not configured. Please define it in your .env or .env.dev file.")
	}

	// Process scopes
	var scopes []string
	if *scopesStr != "" {
		rawScopes := strings.Split(*scopesStr, ",")
		for _, s := range rawScopes {
			trimmed := strings.TrimSpace(s)
			if trimmed != "" {
				scopes = append(scopes, trimmed)
			}
		}
	} else {
		scopes = []string{}
	}

	// Generate expiration time
	expirationTime := time.Now().AddDate(*durationYears, 0, 0)

	// Create JWT token
	claims := jwt.MapClaims{
		"iss":             "nembus-api",
		"sub":             *clientID,
		"client_id":       *clientID,
		"client_name":     *clientName,
		"tenant_slug":     *tenantSlug,
		"organization_id": *organizationID,
		"token_type":      "machine",
		"scopes":          scopes,
		"is_m2m":          true,
		"exp":             expirationTime.Unix(),
		"iat":             time.Now().Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(jwtSecret))
	if err != nil {
		log.Fatalf("Failed to sign token: %v", err)
	}

	// Load existing registry or create new
	registryFile := "config/m2m_clients.json"
	var registry M2MRegistry

	data, err := os.ReadFile(registryFile)
	if err != nil {
		if os.IsNotExist(err) {
			// Initialize with empty slice
			registry.Clients = []M2MClient{}
		} else {
			log.Fatalf("Failed to read registry file: %v", err)
		}
	} else {
		err = json.Unmarshal(data, &registry)
		if err != nil {
			log.Fatalf("Failed to parse registry file: %v", err)
		}
	}

	// Create/Update the M2M client entry
	newClient := M2MClient{
		ClientID:       *clientID,
		ClientName:     *clientName,
		TenantSlug:     *tenantSlug,
		OrganizationID: int32(*organizationID),
		Scopes:         scopes,
		IsActive:       true,
		Token:          tokenString,
	}

	// Check if client ID already exists
	updated := false
	for i, c := range registry.Clients {
		if c.ClientID == *clientID {
			registry.Clients[i] = newClient
			updated = true
			break
		}
	}

	if !updated {
		registry.Clients = append(registry.Clients, newClient)
	}

	// Write registry back to config file
	updatedData, err := json.MarshalIndent(registry, "", "  ")
	if err != nil {
		log.Fatalf("Failed to serialize registry: %v", err)
	}

	// Ensure config dir exists
	err = os.MkdirAll("config", 0755)
	if err != nil {
		log.Fatalf("Failed to create config directory: %v", err)
	}

	err = os.WriteFile(registryFile, updatedData, 0600)
	if err != nil {
		log.Fatalf("Failed to write to registry file: %v", err)
	}

	// Print output
	fmt.Println("======================================================================")
	fmt.Println(" M2M CLIENT GENERATED AND REGISTERED SUCCESSFULLY")
	fmt.Println("======================================================================")
	fmt.Printf("Client ID:    %s\n", *clientID)
	fmt.Printf("Client Name:  %s\n", *clientName)
	fmt.Printf("Tenant Slug:  %s\n", *tenantSlug)
	fmt.Printf("Organization: %d\n", *organizationID)
	fmt.Printf("Scopes:       %s\n", strings.Join(scopes, ", "))
	fmt.Printf("Expires In:   %d years (%s)\n", *durationYears, expirationTime.Format("2006-01-02 15:04:05 MST"))
	fmt.Println("----------------------------------------------------------------------")
	fmt.Println("JWT ACCESS TOKEN (Bearer Token):")
	fmt.Println(tokenString)
	fmt.Println("======================================================================")
	fmt.Println("Save the token securely. It has been whitelisted in config/m2m_clients.json")
}
