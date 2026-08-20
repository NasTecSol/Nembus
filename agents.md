# Nembus SAP Migration / AI Enrichment Worklog

## Source of Truth

- Current architect-approved Nembus repository is authoritative for architecture/schema.
- Historical SAP backup/data samples are supporting evidence only.
- Preserve existing SAP-agent -> mappings -> batch -> core migration architecture.

## Architect Contracts

- `products.product_type` allowed semantic values are exactly:
  - `standard`
  - `raw_material`
  - `fixed_asset`
  - `finished_good`
- AI must never invent another `product_type` value.
- Structured SAP data always has precedence over AI.
- AI must never overwrite/infer SKU/ItemCode, barcode ownership, inventory, warehouse, price, tax, UoM conversion factors, supplier identity, active/sellable/purchasable flags, or SAP document identity.
- Existing populated SAP category stays authoritative for the first MVP.
- AI-derived values must be reviewable; new taxonomy must not be silently created.

## Completed Work

### Phase 1 — Category vs Product Type Mapping

- Proven repository behavior: `Fixed Asset` / `Fixed Assets` map to `fixed_asset`; `Raw Material` / `Raw Materials` map to `raw_material` in `packages/sap/mappings/models.go:325-345`.
- Classified type groups are skipped as normal categories in `apps/sap-agent/internal/etl/extractors/catalog.go:34-45`.
- Ordinary OITB groups remain `CAT-<ItmsGrpCod>` through `packages/sap/mappings/models.go:351-362` and `438-445`.
- Product canonicalization emits the mapped `product_type`; no Stage 1 product application change was made.
- Relevant historical changes are in commits `5226290` and `6ec055c`, affecting `packages/sap/mappings/models.go` and `apps/sap-agent/internal/etl/extractors/catalog.go`.
- Mapping tests covering the known groups are present in `packages/sap/mappings/models_test.go:109-142`; execution is UNVERIFIED because Go is unavailable.
- `finished_good` source/rule is not established; see Open Questions.
- Status: SAFE TO KEEP.

### Phase 2A — Structured Brand Extraction Diagnostics

- `QueryBrands` / OMRG query failures are returned as errors in `apps/sap-agent/internal/etl/extractors/catalog.go:53-56` instead of becoming successful zero-row extraction.
- `rows.Err()` is checked in `apps/sap-agent/internal/etl/extractors/catalog.go:69-70`.
- A successful zero-row OMRG result remains valid.
- No ItemName inference was added.
- Tooling limitation: `go` and `gofmt` were unavailable in the latest environment; test and formatting verification remain PENDING.
- Status: SAFE TO KEEP, subject to the pending tool verification.

### Phase 2B / Stage 1 — AI Enrichment Foundation

- Added `product_enrichment_suggestions` persistence in `packages/core/db/schema/35_product_enrichment.sql:5-35` and `packages/core/db/migrations/20260820000000.sql:4-34`.
- Added provider-neutral proposal DTOs and validation in `packages/core/enrichment/product_enrichment.go:14-359`.
- Closed proposal actions are `KEEP_EXISTING`, `MATCH_EXISTING`, `PROPOSE_NEW`, `NO_MATCH`, and `UNSUPPORTED_TARGET` (`product_enrichment.go:18-24`).
- Review statuses are `pending`, `approved`, `rejected`, `failed`, `retryable`, and `applied` (`35_product_enrichment.sql:23-24`).
- Idempotency is keyed by organization, product, source fingerprint, and contract version (`35_product_enrichment.sql:30-31`; query `product_enrichment_suggestions.sql:25-27`).
- Proposal fields cover brand, category, description, and unsupported semantics; `product_type` is absent from the DTO (`product_enrichment.go:79-86`).
- There are no AI provider/API calls, no product mutation, no post-ingestion hook, and no `products.metadata` change in the Stage 1 commit.
- IMPORTANT: sqlc generation was unavailable; `packages/core/repository/product_enrichment_suggestions.sql.go` and the appended model in `packages/core/repository/models.go:2608-2634` were added manually.
- IMPORTANT: Atlas/checksum validation was unavailable; `packages/core/db/migrations/atlas.sum:1-5` does not list `20260820000000.sql`.
- Status: UNDER REVIEW until the corrections and generator/checksum verification below are completed.

## AI MVP Boundary

- `product_type`: never AI.
- Populated structured category: `KEEP_EXISTING`.
- Structured SAP brand: wins when present.
- Unresolved brand: future AI suggestion allowed.
- Empty product description: future AI suggestion allowed.
- Attributes/families/variants/model/capacity/dimensions remain unsupported or future work unless schema is approved.
- Suggestions are stored outside `products.metadata`.
- First production rollout should not silently create brands/categories.

## Data Evidence

- The 100-product Nembus sample contains Arabic/English names with semantic data embedded in `ItemName`.
- Examples include Pantene shampoo names, HIKvision, Huawei, and Epson.
- Brand fields were null in representative products.
- SAP-derived UoM/conversion data was already structured and must remain authoritative.
- Team comparison for INV00006/INV00007/INV00008 was Nembus vs SAP-mapped data, not direct raw SAP.
- Direct values for OITM.FrgnName, OITM.ItmsGrpCod, OITB.ItmsGrpNam, OITM.FirmCode, and manufacturer name were not available from that report.

## Schema Gaps

- No normalized product attribute-definition/value model.
- No dedicated content-size/capacity/dimension/resolution/model fields.
- No product-family/parent-product relationship.
- Variants exist but SAP ingestion creates independent products.
- No alias/rule table yet.
- No enrichment review persistence existed before Stage 1.

## Remaining Phases

- Stage 1 review/correction.
- Provider-neutral enrichment service interface.
- Candidate dictionary lookup for canonical brands/categories.
- AI provider integration and prompt/structured-output contract.
- Post-commit asynchronous/retryable enrichment triggering.
- Review/approval API.
- Application of approved suggestions.
- Approved deterministic aliases/rules.
- UI/reviewer workflow if required.
- Production observability/security/hardening.
- Future schema work for attributes/families only if approved.

## Open Questions

- Source/rule for `finished_good` is not established.
- AI provider/model/region/retention policy is not selected.
- Reviewer roles/permissions are not established.
- Auto-apply policy is not approved.
- Live tenant taxonomy contents are not guaranteed by seed data.
- Future product-family model is not approved.

## Latest Verification

- Review date: 2026-08-20.
- HEAD is `f8b7ba4` (`AI implemetation stage 1`), tracking `origin/feature-agent`; before this worklog write, `git status --porcelain=v2 --branch` showed a clean worktree with no untracked files. The Stage 1 files are therefore committed, not currently uncommitted.
- Stage 1 commit scope is exactly seven tracked files: two schema/migration SQL files, one query file, two repository files, and two enrichment Go files. Commit summary: 911 insertions; `git diff --check` passed.
- Schema/migration definitions are semantically equal after accounting for the migration-only `IF NOT EXISTS` guards and comments. Referenced IDs are `INTEGER`/`int32`: organizations `10_identity_rbac.sql:5-6`, products `30_catalog.sql:85-87`, and users `20_stores_terminals.sql:39-41`. FKs use `ON DELETE CASCADE` for organization/product and `ON DELETE SET NULL` for reviewer, consistently in schema and migration.
- The source constraint/default restricts `source` to `ai`; all six statuses, JSONB/text/provider/model/reviewer nullability, timestamp conventions, uniqueness, index, and schema ordering are internally consistent. The query organization predicates are present on get/list and every status update (`product_enrichment_suggestions.sql:29-88`).
- SQLC config is `packages/core/sqlc.yaml`, SQLC version convention is v1.30.0 in existing generated files, and the repository uses `sqlc.narg` for optional filters. SQLC and Go were unavailable; no Go test/build or real SQLC generation was run.
- The checked-in hand-written SQLC companion has an unused `github.com/jackc/pgx/v5` import at `packages/core/repository/product_enrichment_suggestions.sql.go:15`; `pgx` is not referenced elsewhere in that file. This is a static compile blocker.
- The hand-written companion keeps `SELECT *` and uses a shared scanner at `product_enrichment_suggestions.sql.go:80-103` and `238-269`, while existing v1.30.0 output expands `SELECT *` to explicit columns and scans inline (for example `packages/core/repository/organizations.sql.go:84-104` and `130-172`). It cannot be proven generator-compatible without running the real generator.
- `MarkProductEnrichmentSuggestionFailed` and `MarkProductEnrichmentSuggestionRetryable` allow every prior status and only change `status`/`updated_at` (`packages/core/queries/product_enrichment_suggestions.sql:64-78`). They can preserve stale `reviewer_id`, `reviewed_at`, or `applied_at`; an applied/reviewed row can therefore become retryable/failed with contradictory lifecycle fields.
- The authoritative-field guard has a normalization loophole: `prohibitedEnrichmentTarget` only lowercases, replaces spaces/hyphens, and checks underscore tokens (`packages/core/enrichment/product_enrichment.go:337-359`). CamelCase keys such as `productType` or `isActive` can bypass the prohibited-target check even though the contract forbids indirect targeting.
- Structured-category precedence is enforced by `ProposalSet.Validate` (`product_enrichment.go:99-110`, `155-164`), but structured-brand precedence is not enforced in the DTO: `BrandProposal.validate` receives no current-brand state (`product_enrichment.go:127-152`). This remains an IMPORTANT contract verification gap for the MVP service boundary.
- The seven tests in `packages/core/enrichment/product_enrichment_test.go:9-98` substantively cover product_type JSON rejection, populated-category rejection, canonical brand matching, review-only new proposals, unsupported semantics, several prohibited fields, and confidence above one. They do not cover all five actions, the lower confidence boundary/NaN/Inf, structured-brand precedence, camelCase prohibited targets, JSONB persistence/query lifecycle, idempotency, organization isolation, or status-field cleanup.
- Static scope review of commit `f8b7ba4` found no Stage 1 changes under `packages/sap`, `apps/sap-agent`, SAP extraction/mapping, existing product/category/brand application logic, product_type behavior, provider/API code, worker/hook, products.metadata, or variant/family application. Earlier Phase 1/2A work is outside this commit and is recorded above.
- Atlas migration manifest exists at `packages/core/db/migrations/atlas.sum:1-5` and omits `20260820000000.sql`; Atlas hash/checksum regeneration is required before the migration can be accepted. Atlas was unavailable and the manifest was not modified.

### Findings

#### CRITICAL

- `packages/core/repository/product_enrichment_suggestions.sql.go:15`: unused `pgx` import makes the manually added repository file fail Go compilation.
- `packages/core/queries/product_enrichment_suggestions.sql:64-78`: failed/retryable transitions are unrestricted and do not clear stale review/application fields, allowing contradictory lifecycle state.
- `packages/core/enrichment/product_enrichment.go:337-359`: camelCase authoritative targets can bypass the prohibited-field guard, violating the no-indirect-targeting contract.

#### IMPORTANT

- `packages/core/repository/product_enrichment_suggestions.sql.go:1-7,80-103,238-269` and `packages/core/repository/models.go:2608-2634`: manually added generated companions differ from committed SQLC v1.30.0 conventions; real sqlc generation is required and may overwrite them materially.
- `packages/core/db/migrations/atlas.sum:1-5`: the new migration is absent from the checked-in Atlas checksum manifest.
- `packages/core/enrichment/product_enrichment.go:99-110,127-152`: structured category is guarded, but structured SAP brand precedence is not represented in the validation inputs.
- `packages/core/enrichment/product_enrichment_test.go:9-98`: critical DTO, persistence, tenant-scoping, idempotency, and lifecycle cases are untested.

#### MINOR

- `packages/core/db/schema/35_product_enrichment.sql:28-29` and `packages/core/db/migrations/20260820000000.sql:27-28` leave created/updated timestamps nullable, matching repository convention but not enforcing non-null audit timestamps.
- `packages/core/db/schema/35_product_enrichment.sql:5-35` and `packages/core/db/migrations/20260820000000.sql:4-34` differ in comments and idempotence guards only; normalized definitions otherwise match.

## Next Action

Apply the smallest Stage 1 correction: remove the unused import and regenerate the repository files with the repository's SQLC v1.30.0 workflow; constrain failed/retryable lifecycle transitions and clear stale lifecycle fields; close the authoritative-target normalization loophole and add focused tests; then regenerate and verify `packages/core/db/migrations/atlas.sum`. Re-run Go formatting/tests and SQLC/Atlas validation when those tools are available. Until then, Stage 1 remains UNDER REVIEW.

Final review verdict: NEEDS CORRECTION.
