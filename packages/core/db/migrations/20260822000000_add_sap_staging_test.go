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
	provisioning := readMigrationFixture(t, provisioningMigration)

	canonical = normalizeSQL(canonical)
	provisioning = normalizeSQL(provisioning)
	if !strings.Contains(provisioning, normalizeSQL("CREATE SCHEMA IF NOT EXISTS staging")) {
		t.Fatal("provisioning migration is missing the staging schema")
	}

	for _, table := range canonicalStagingTables {
		canonicalBody := canonicalCreateTableBody(t, canonical, table)
		provisioningBody := canonicalCreateTableBody(t, provisioning, table)
		assertCanonicalTableColumns(t, table, canonicalBody, provisioningBody)
	}

	batch := canonicalCreateTableBody(t, provisioning, "sap_migration_batches")
	assertSQLContains(t, batch, "batch_id VARCHAR(100) UNIQUE NOT NULL", "batch_id unique semantics")
	assertSQLContains(t, batch, "organization_id INTEGER NOT NULL REFERENCES organizations(id) ON DELETE CASCADE", "batch organization foreign key")
}

func TestSAPStagingValidationCoversCanonicalContract(t *testing.T) {
	canonical := readMigrationFixture(t, canonicalSchemaPath)
	validation := normalizeSQL(readMigrationFixture(t, validationMigration))

	for _, table := range canonicalStagingTables {
		block := canonicalCreateTableBody(t, normalizeSQL(canonical), table)
		validationTable := jsonArraySection(t, validation, table)
		for _, column := range splitTopLevelDefinitions(block) {
			name, typ, notNull, defaultKind := canonicalColumnContract(t, column)
			for _, fragment := range []string{
				fmt.Sprintf(`"name":"%s"`, name),
				fmt.Sprintf(`"type":"%s"`, typ),
				fmt.Sprintf(`"not_null":%t`, notNull),
				fmt.Sprintf(`"default":"%s"`, defaultKind),
			} {
				assertSQLContains(t, validationTable, fragment, fmt.Sprintf("validation coverage for staging.%s column %s", table, name))
			}
		}
	}

	for _, table := range canonicalStagingTables {
		if !strings.Contains(validation, normalizeSQL(`"`+table+`": [`)) {
			t.Errorf("validation migration does not reference staging.%s", table)
		}
	}

	batchUnique := validationConstraintSection(t, validation, "sap_migration_batches", "u")
	assertSQLContains(t, batchUnique, "c.conkey = ARRAY[(SELECT a.attnum FROM pg_attribute a WHERE a.attrelid = c.conrelid AND a.attname = 'batch_id')]::SMALLINT[]", "batch_id unique validation")

	batchForeignKey := validationConstraintSection(t, validation, "sap_migration_batches", "f")
	for _, fragment := range []string{
		"c.conkey = ARRAY[(SELECT a.attnum FROM pg_attribute a WHERE a.attrelid = c.conrelid AND a.attname = 'organization_id')]::SMALLINT[]",
		"referenced_schema.nspname = 'public'",
		"referenced_table.relname = 'organizations'",
		"c.confkey = ARRAY[(SELECT a.attnum FROM pg_attribute a WHERE a.attrelid = c.confrelid AND a.attname = 'id')]::SMALLINT[]",
		"c.confdeltype = 'c'",
		"c.confupdtype = 'a'",
	} {
		assertSQLContains(t, batchForeignKey, fragment, "organizations FK validation")
	}

	for _, fragment := range []string{
		`pg_attribute`, `format_type`, `attnotnull`, `pg_attrdef`, `pg_get_expr`,
		`pg_constraint`, `conkey`, `confkey`, `confrelid`, `confdeltype`, `confupdtype`,
	} {
		assertSQLContains(t, validation, fragment, "catalog contract validation")
	}

	destructive := regexp.MustCompile(`(?im)^\s*(drop|alter|update|delete|truncate|create\s+(table|schema))\b`)
	if destructive.MatchString(stripSQLComments(readMigrationFixture(t, validationMigration))) {
		t.Fatal("validation migration must not contain destructive or mutating SQL")
	}
}

func assertCanonicalTableColumns(t *testing.T, table, canonicalBody, provisioningBody string) {
	t.Helper()
	for _, canonicalColumn := range splitTopLevelDefinitions(canonicalBody) {
		name, wantType, wantNotNull, wantDefault := canonicalColumnContract(t, canonicalColumn)
		provisioningColumn := findColumnDefinition(t, provisioningBody, table, name)
		gotName, gotType, gotNotNull, gotDefault := canonicalColumnContract(t, provisioningColumn)
		if gotName != name || gotType != wantType || gotNotNull != wantNotNull || gotDefault != wantDefault {
			t.Errorf("staging.%s column %s differs from canonical contract: got (%s, %s, %t, %s), want (%s, %s, %t, %s)", table, name, gotName, gotType, gotNotNull, gotDefault, name, wantType, wantNotNull, wantDefault)
		}
	}
}

func findColumnDefinition(t *testing.T, body, table, name string) string {
	t.Helper()
	for _, definition := range splitTopLevelDefinitions(body) {
		fields := strings.Fields(definition)
		if len(fields) > 0 && strings.Trim(fields[0], `"`) == name {
			return definition
		}
	}
	t.Fatalf("staging.%s is missing canonical column %s", table, name)
	return ""
}

func assertSQLContains(t *testing.T, section, fragment, description string) {
	t.Helper()
	if !strings.Contains(normalizeSQL(section), normalizeSQL(fragment)) {
		t.Errorf("missing %s: %q", description, fragment)
	}
}

func normalizeSQL(sql string) string {
	sql = stripSQLComments(sql)
	sql = strings.Join(strings.Fields(strings.ToLower(sql)), " ")
	return regexp.MustCompile(`\s*([(),\[\]])\s*`).ReplaceAllString(sql, "$1")
}

func stripSQLComments(sql string) string {
	sql = regexp.MustCompile(`(?s)/\*.*?\*/`).ReplaceAllString(sql, "")
	return regexp.MustCompile(`(?m)--.*$`).ReplaceAllString(sql, "")
}

func jsonArraySection(t *testing.T, source, table string) string {
	t.Helper()
	marker := normalizeSQL(fmt.Sprintf(`"%s":`, table))
	start := strings.Index(source, marker)
	if start < 0 {
		t.Fatalf("validation migration is missing the staging.%s contract section", table)
	}
	open := strings.Index(source[start+len(marker):], "[")
	if open < 0 {
		t.Fatalf("validation migration has no array for staging.%s", table)
	}
	open += start + len(marker)
	return balancedSQLSection(t, source, open, '[', ']', "staging."+table+" validation array")
}

func validationConstraintSection(t *testing.T, source, table, constraintType string) string {
	t.Helper()
	marker := normalizeSQL(fmt.Sprintf("WHERE c.conrelid = 'staging.%s'::regclass AND c.contype = '%s'", table, constraintType))
	start := strings.Index(source, marker)
	if start < 0 {
		t.Fatalf("validation migration is missing staging.%s constraint type %s", table, constraintType)
	}
	end := strings.Index(source[start:], normalizeSQL(") THEN"))
	if end < 0 {
		t.Fatalf("validation migration has an incomplete staging.%s constraint type %s section", table, constraintType)
	}
	return source[start : start+end]
}

func balancedSQLSection(t *testing.T, source string, open int, opening, closing byte, description string) string {
	t.Helper()
	depth := 0
	inString := false
	for i := open; i < len(source); i++ {
		switch source[i] {
		case '\'':
			if inString && i+1 < len(source) && source[i+1] == '\'' {
				i++
				continue
			}
			inString = !inString
		case opening:
			if !inString {
				depth++
			}
		case closing:
			if !inString {
				depth--
				if depth == 0 {
					return source[open : i+1]
				}
			}
		}
	}
	t.Fatalf("%s has no closing delimiter", description)
	return ""
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
