package client

import (
	"strings"
	"testing"
)

func TestAdministrativeDumpSections(t *testing.T) {
	tests := []struct {
		name, kind, schema string
		excluded           bool
	}{
		{"_avnadmin_monitor", "SCHEMA", "-", true},
		{"explain_statement(text)", "FUNCTION", "_avnadmin_monitor", true},
		{"SCHEMA _avnadmin_monitor", "ACL", "-", true},
		{"atlas_schema_revisions", "SCHEMA", "-", true},
		{"atlas_schema_revisions", "TABLE DATA", "atlas_schema_revisions", true},
		{"orders", "TABLE", "public", false},
		{"sap_items", "TABLE DATA", "staging", false},
		{"_avnadmin_monitor_notes", "TABLE", "public", false},
		{"plpgsql", "EXTENSION", "-", false},
	}
	for _, tt := range tests {
		if got := isAdministrativeDumpSection(tt.name, tt.kind, tt.schema); got != tt.excluded {
			t.Errorf("section %q / %q: excluded=%v; want %v", tt.name, tt.schema, got, tt.excluded)
		}
	}
	for _, prefix := range []string{"-- Name: ", "-- Data for Name: "} {
		if dumpSectionHeader.FindStringSubmatch(prefix+"atlas_schema_revisions; Type: TABLE DATA; Schema: atlas_schema_revisions; Owner: cloud") == nil {
			t.Fatal("pg_dump section header was not recognized")
		}
	}
}

func TestSplitSQLPreservesDollarQuotesAndUnicode(t *testing.T) {
	body := strings.Repeat("SELECT 'مرحبا; don''t split';\n", 10000)
	function := "CREATE FUNCTION public.test() RETURNS void AS $body$\n" + body + "$body$ LANGUAGE plpgsql"
	statements := splitSQL(function + "; SELECT 'نهاية';")
	if len(statements) != 2 || statements[0] != function || statements[1] != "SELECT 'نهاية'" {
		t.Fatal("SQL splitter changed Unicode or split inside a quoted function")
	}
}
