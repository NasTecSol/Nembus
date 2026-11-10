package client

import "regexp"

var dumpSectionHeader = regexp.MustCompile(`^-- (?:Data for )?Name: (.*); Type: ([^;]+); Schema: ([^;]+); Owner:.*$`)

func isAdministrativeDumpSection(name, kind, schema string) bool {
	for _, excluded := range []string{"_avnadmin_monitor", "atlas_schema_revisions"} {
		if schema == excluded || (kind == "SCHEMA" && name == excluded) || name == "SCHEMA "+excluded {
			return true
		}
	}
	return false
}
