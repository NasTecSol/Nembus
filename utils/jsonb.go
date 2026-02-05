package utils

import "encoding/json"

// BytesToJSONRawMessage converts JSONB []byte from the DB into json.RawMessage
// so it marshals as embedded JSON (object/array) instead of base64.
//
// Use this in usecases before returning entities that have jsonb columns
// (e.g. settings, metadata). Nil or empty input becomes {}.
//
// Pattern for other usecases: define an output struct with jsonb fields as
// json.RawMessage, copy repo entity into it, and set e.g. Metadata:
// utils.BytesToJSONRawMessage(row.Metadata). Return the output struct in
// the response Data so API responses show proper JSON instead of bytes.
//
// Usecases already using this pattern: tenant (settings), role, menu, submenu,
// module, organization, store, user (metadata). Other repo entities with
// jsonb (e.g. Brand, Product, TaxCategory, POS rows) should use the same
// approach when returned in API responses.
func BytesToJSONRawMessage(data []byte) json.RawMessage {
	if len(data) == 0 {
		return json.RawMessage([]byte("{}"))
	}
	return json.RawMessage(data)
}
