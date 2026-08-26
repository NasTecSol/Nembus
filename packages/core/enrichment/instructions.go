package enrichment

// StrictInferenceInstructions is the provider-neutral instruction contract.
// It intentionally contains no provider parameters, wrapper syntax, tool
// definitions, credentials, or model-specific response controls.
const StrictInferenceInstructions = `Return exactly one JSON object matching the enrichment response contract. Emit no markdown, code fences, commentary, wrapper object, or provider metadata.

Structured SAP values are authoritative. Never propose, infer, replace, or mutate product_type; the only valid product_type values are standard, raw_material, fixed_asset, and finished_good, and product_type is not an output field. Never propose SKU, SAP ItemCode, barcode ownership, inventory or stock, warehouse or store inventory, price or pricing, tax, UoM conversion factors, supplier identity, active, sellable, purchasable, or SAP document identity.

If the structured category is populated, return category action KEEP_EXISTING. If the structured SAP brand is resolved, return brand action KEEP_EXISTING. Only use MATCH_EXISTING with an exact supplied canonical candidate ID and/or code. Do not use a model spelling or name as taxonomy identity. If an existing supplied canonical brand candidate matches, use MATCH_EXISTING instead of PROPOSE_NEW; the server-side candidate remains authoritative. For PROPOSE_NEW brand, copy or extract canonical_name verbatim from source_item_name, preserving its original language and script. Never translate, transliterate, case-fold, rewrite, or substitute a model-authored alternate spelling. PROPOSE_NEW is review-only; never create a brand or category.

Description proposals are factual catalog text, preserve the source language by default, contain no unsupported marketing claims, and are review-only. Use PROPOSE_NEW only for a non-empty description of at most 500 Unicode code points, or NO_MATCH when evidence is insufficient. Do not use MATCH_EXISTING for descriptions.

Put schema-unsupported information only in unsupported_semantics. Packaging and UoM text is informational evidence; never infer a conversion factor from it. Do not invent specifications. Unsupported semantics never mutate product columns.`

// BuildInferenceInstructions returns the same provider-neutral instruction
// contract for every adapter. Request data and candidate dictionaries remain
// separate structured input, so source text cannot alter the safety rules.
func BuildInferenceInstructions(_ EnrichmentRequest) string {
	return StrictInferenceInstructions
}
