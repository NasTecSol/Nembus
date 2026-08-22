package migrations

import (
	"fmt"
	"os"
	"regexp"
	"strings"
	"testing"
)

const (
	canonicalSchemaPath   = "../schema/95_sap_staging.sql"
	provisioningMigration = "20260822000000_add_sap_staging.sql"
	validationMigration   = "20260822010000_validate_sap_staging_contract.sql"
)

var canonicalStagingTables = []string{
	"sap_migration_batches",
	"sap_stores",
	"sap_products",
	"sap_inventory",
}

func TestSAPStagingMigrationHistoryProvisionsCanonicalObjects(t *testing.T) {
	canonical := readMigrationFixture(t, canonicalSchemaPath)
	provisioning := strings.ToLower(readMigrationFixture(t, provisioningMigration))

	for _, table := range canonicalStagingTables {
		if !strings.Contains(strings.ToLower(canonical), "create table if not exists staging."+table) {
			t.Errorf("canonical schema is missing staging.%s", table)
		}
		if !strings.Contains(provisioning, "create table if not exists staging."+table) {
			t.Errorf("provisioning migration is missing staging.%s", table)
		}
	}
}

func TestSAPStagingValidationCoversCanonicalContract(t *testing.T) {
	canonical := readMigrationFixture(t, canonicalSchemaPath)
	validation := strings.ToLower(readMigrationFixture(t, validationMigration))

	for _, table := range canonicalStagingTables {
		block := canonicalCreateTableBody(t, canonical, table)
		for _, column := range splitTopLevelDefinitions(block) {
			name, typ, notNull, defaultKind := canonicalColumnContract(t, column)
			for _, fragment := range []string{
				fmt.Sprintf(`"name":"%s"`, name),
				fmt.Sprintf(`"type":"%s"`, typ),
				fmt.Sprintf(`"not_null":%t`, notNull),
				fmt.Sprintf(`"default":"%s"`, defaultKind),
			} {
				if !strings.Contains(validation, strings.ToLower(fragment)) {
					t.Errorf("validation migration does not cover staging.%s column %s with %s", table, name, fragment)
				}
			}
		}
	}

	for _, fragment := range []string{
		`pg_attribute`, `format_type`, `attnotnull`, `pg_attrdef`, `pg_get_expr`,
		`pg_constraint`, `conkey`, `confkey`, `confrelid`, `confdeltype`, `confupdtype`,
		`c.contype = 'p'`, `c.contype = 'u'`, `c.contype = 'f'`,
		`referenced_schema.nspname = 'public'`, `referenced_table.relname = 'organizations'`,
		`c.confdeltype = 'c'`, `c.confupdtype = 'a'`,
	} {
		if !strings.Contains(validation, strings.ToLower(fragment)) {
			t.Errorf("validation migration is missing required contract check %q", fragment)
		}
	}

	destructive := regexp.MustCompile(`(?im)^\s*(drop|alter|update|delete|truncate|create\s+(table|schema))\b`)
	if destructive.MatchString(validation) {
		t.Fatal("validation migration must not contain destructive or mutating SQL")
	}
}

func readMigrationFixture(t *testing.T, path string) string {
	t.Helper()
	contents, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read migration fixture %q: %v", path, err)
	}
	return string(contents)
}

func canonicalCreateTableBody(t *testing.T, schema, table string) string {
	t.Helper()
	start := strings.Index(strings.ToLower(schema), "create table if not exists staging."+table)
	if start < 0 {
		t.Fatalf("canonical schema is missing staging.%s", table)
	}
	open := strings.Index(schema[start:], "(")
	if open < 0 {
		t.Fatalf("canonical staging.%s definition has no opening parenthesis", table)
	}
	open += start
	depth := 0
	inString := false
	for i := open; i < len(schema); i++ {
		switch schema[i] {
		case '\'':
			if inString && i+1 < len(schema) && schema[i+1] == '\'' {
				i++
				continue
			}
			inString = !inString
		case '(':
			if !inString {
				depth++
			}
		case ')':
			if !inString {
				depth--
				if depth == 0 {
					return schema[open+1 : i]
				}
			}
		}
	}
	t.Fatalf("canonical staging.%s definition has no closing parenthesis", table)
	return ""
}

func splitTopLevelDefinitions(body string) []string {
	body = regexp.MustCompile(`(?m)--.*$`).ReplaceAllString(body, "")
	var definitions []string
	start := 0
	depth := 0
	inString := false
	for i := 0; i < len(body); i++ {
		switch body[i] {
		case '\'':
			if inString && i+1 < len(body) && body[i+1] == '\'' {
				i++
				continue
			}
			inString = !inString
		case '(':
			if !inString {
				depth++
			}
		case ')':
			if !inString {
				depth--
			}
		case ',':
			if !inString && depth == 0 {
				definitions = append(definitions, strings.TrimSpace(body[start:i]))
				start = i + 1
			}
		}
	}
	if tail := strings.TrimSpace(body[start:]); tail != "" {
		definitions = append(definitions, tail)
	}
	return definitions
}

func canonicalColumnContract(t *testing.T, definition string) (string, string, bool, string) {
	t.Helper()
	fields := strings.Fields(definition)
	if len(fields) < 2 {
		t.Fatalf("invalid canonical column definition %q", definition)
	}

	name := strings.Trim(fields[0], `"`)
	upper := strings.ToUpper(definition)
	notNull := strings.Contains(upper, " NOT NULL")

	typ := ""
	switch {
	case strings.HasPrefix(upper, "ID SERIAL"):
		typ = "integer"
	case strings.HasPrefix(upper, "BATCH_ID VARCHAR(100)"), strings.HasPrefix(upper, "RUN_ID VARCHAR(100)"), strings.HasPrefix(upper, "SKU VARCHAR(100)"), strings.HasPrefix(upper, "PRODUCT_SKU VARCHAR(100)"), strings.HasPrefix(upper, "PRIMARY_BARCODE VARCHAR(100)"):
		typ = "character varying(100)"
	case strings.HasPrefix(upper, "CODE VARCHAR(50)"), strings.HasPrefix(upper, "STORE_TYPE VARCHAR(50)"), strings.HasPrefix(upper, "CATEGORY_CODE VARCHAR(50)"), strings.HasPrefix(upper, "BRAND_CODE VARCHAR(50)"), strings.HasPrefix(upper, "PRODUCT_TYPE VARCHAR(50)"), strings.HasPrefix(upper, "DOMAIN VARCHAR(50)"), strings.HasPrefix(upper, "STORE_CODE VARCHAR(50)"), strings.HasPrefix(upper, "STATUS VARCHAR(30)"):
		typ = "character varying(50)"
	case strings.HasPrefix(upper, "NAME VARCHAR(255)"):
		typ = "character varying(255)"
	case strings.HasPrefix(upper, "UOM_CODE VARCHAR(20)"):
		typ = "character varying(20)"
	case strings.HasPrefix(upper, "RECORD_COUNT INTEGER"), strings.HasPrefix(upper, "ORGANIZATION_ID INTEGER"):
		typ = "integer"
	case strings.HasPrefix(upper, "DESCRIPTION TEXT"), strings.HasPrefix(upper, "ERROR_MESSAGE TEXT"):
		typ = "text"
	case strings.HasPrefix(upper, "IS_") || strings.HasPrefix(upper, "TRACK_INVENTORY BOOLEAN"):
		typ = "boolean"
	case strings.HasPrefix(upper, "METADATA JSONB"):
		typ = "jsonb"
	case strings.HasPrefix(upper, "CREATED_AT TIMESTAMP"):
		typ = "timestamp without time zone"
	case strings.HasPrefix(upper, "QUANTITY_") || strings.HasPrefix(upper, "REORDER_LEVEL DECIMAL") || strings.HasPrefix(upper, "MAX_STOCK_LEVEL DECIMAL"):
		typ = "numeric(15,3)"
	}
	if typ == "" {
		t.Fatalf("unsupported canonical column definition %q", definition)
	}

	defaultKind := "none"
	switch {
	case strings.Contains(upper, " SERIAL"):
		defaultKind = "serial"
	case strings.Contains(upper, "DEFAULT 0"):
		defaultKind = "zero"
	case strings.Contains(upper, "DEFAULT 'STAGED'"):
		defaultKind = "staged"
	case strings.Contains(upper, "DEFAULT 'STANDARD'"):
		defaultKind = "standard"
	case strings.Contains(upper, "DEFAULT TRUE"):
		defaultKind = "true"
	case strings.Contains(upper, "DEFAULT FALSE"):
		defaultKind = "false"
	case strings.Contains(upper, "DEFAULT '{}'"):
		defaultKind = "jsonb_empty"
	case strings.Contains(upper, "DEFAULT CURRENT_TIMESTAMP"):
		defaultKind = "current_timestamp"
	}
	return name, typ, notNull, defaultKind
}
