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

## Pre-correction Verification (historical)

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

## Pre-correction Next Action

Apply the smallest Stage 1 correction: remove the unused import and regenerate the repository files with the repository's SQLC v1.30.0 workflow; constrain failed/retryable lifecycle transitions and clear stale lifecycle fields; close the authoritative-target normalization loophole and add focused tests; then regenerate and verify `packages/core/db/migrations/atlas.sum`. Re-run Go formatting/tests and SQLC/Atlas validation when those tools are available. Until then, Stage 1 remains UNDER REVIEW.

Pre-correction review verdict: NEEDS CORRECTION.

## Stage 1 Corrective Patch — 2026-08-20

The strict-review corrections were implemented without beginning Stage 2 or changing the completed SAP work. The pre-correction findings above remain historical review evidence; the current acceptance state is recorded here.

### Corrective Findings and Files Changed

- Removed the unused `pgx` import from `packages/core/repository/product_enrichment_suggestions.sql.go`. The file remains provisionally hand-maintained because sqlc was unavailable; generator compatibility is not proven.
- Corrected failed/retryable lifecycle SQL in `packages/core/queries/product_enrichment_suggestions.sql` and its manual repository companion. Failed transitions now allow only `pending`/`retryable`; retryable transitions allow only `pending`/`failed`; both set `status`, clear `reviewer_id`, `reviewed_at`, and `applied_at`, and refresh `updated_at` with `CURRENT_TIMESTAMP`. Approval/rejection also clear `applied_at` so legacy retryable rows cannot retain stale application state.
- Added deterministic case/separator normalization and expanded the authoritative target dictionary in `packages/core/enrichment/product_enrichment.go`. It protects product type, SKU/ItemCode variants, barcode ownership, inventory/stock/warehouse, pricing, tax, UoM/conversion factors, supplier, and active/sellable/purchasable variants while preserving unsupported evidence such as anti-dandruff, capacity, dimensions, resolution, size text, and model number.
- Added structured brand precedence validation in `packages/core/enrichment/product_enrichment.go`. A resolved `brand_id` or canonical `brand_code` permits no replacement action other than `KEEP_EXISTING`; unresolved brands still allow canonical matches, review-only new proposals, and `NO_MATCH`.
- Added focused tests in `packages/core/enrichment/product_enrichment_test.go` for normalization variants, protected SAP-authoritative fields, allowed unsupported evidence, structured brand precedence, unresolved brand proposals, and confidence boundaries including NaN/Inf.

### Current Verification and Acceptance State

- Changed files: `agents.md`, `packages/core/enrichment/product_enrichment.go`, `packages/core/enrichment/product_enrichment_test.go`, `packages/core/queries/product_enrichment_suggestions.sql`, and `packages/core/repository/product_enrichment_suggestions.sql.go`. No commit or push was performed.
- `git diff --check`: passed.
- `go`, `gofmt`, `sqlc`, `atlas`, and `psql`: unavailable. Go tests and gofmt could not run; no tools were installed.
- sqlc generation: not performed. The source query and manual companion were synchronized by inspection. The repository's established command remains `cd packages/core && sqlc generate`; real generator verification is still required.
- Atlas checksum regeneration: not performed. `packages/core/db/migrations/atlas.sum` was left unchanged; the established command remains `atlas migrate hash --dir "file://packages/core/db/migrations"`, and checksum validation is still required.
- Lifecycle review: create/upsert preserves existing lifecycle fields and defaults new rows; get/list are read-only; approve/reject set reviewer state and clear application state; failed/retryable clear all review/application state; applied is restricted to approved and sets `applied_at`.

### Current Stage 1 Status

- Status: NEEDS GENERATOR VERIFICATION.
- Remaining IMPORTANT acceptance blockers: real sqlc v1.30.0 generation/diff verification and Atlas checksum regeneration/validation. Go formatting/tests also remain unverified because the required tools are unavailable.
- No remaining CRITICAL source-correction finding is known from static review. The prior manual-SQLC compatibility concern remains an acceptance blocker until generation is run.

## Next Action

When the repository's existing tools are available, run `gofmt` on the modified Go files, the focused `packages/core/enrichment` tests, `cd packages/core && sqlc generate`, and the established Atlas hash/validate workflow. Keep Stage 1 at NEEDS GENERATOR VERIFICATION until both sqlc generation and Atlas checksum verification succeed; do not mark Stage 2 ready before then.

## Stage 1 Toolchain/Generator Verification — 2026-08-20

### Current Task

Stage 1 generator/toolchain verification

This verification was limited to the existing repository toolchain. No Stage 2 work, new enrichment behavior, SAP changes, source-query changes, provider/API work, application mutation, installation, commit, or push was performed.

### Repository Tooling Discovered

- `packages/core/sqlc.yaml` is the authoritative SQLC configuration. The repository-documented generation command is `cd packages/core && sqlc generate`, also exposed through the root `Makefile` target `generate-core`/`sqlc`.
- `packages/core/atlas.hcl` and the root `atlas.hcl` define Atlas environments. The repository-documented checksum command is `cd packages/core && atlas migrate hash --dir "file://db/migrations"`; validation is `atlas migrate validate --dir "file://db/migrations"`, also exposed through the root `Makefile` targets `db-hash`/`db-validate`.
- `.github/workflows/db-validate.yml` is the only complete automated verification path found. It installs Go 1.25.x, Atlas, and sqlc v1.30.0 on a GitHub-hosted runner, then runs Atlas validation/application, `cd packages/core && sqlc generate`, `git diff --exit-code repository/`, and core/server tests. Its installation steps are not an available local toolchain and were not reproduced.
- `SETUP_GUIDE.md` documents direct host prerequisites and optional SQLC installation; it does not define a local verification container.
- The root `Dockerfile` has a Go build stage and downloads Atlas into the production image, but does not install SQLC or define a verification target. No devcontainer or repository compose file was found. Docker Compose is installed, but the Docker daemon is not running, so the Dockerfile cannot provide a usable execution path here.
- Host availability: `docker compose version` succeeded with Docker Compose v5.3.1. `go`, `gofmt`, `sqlc`, `atlas`, and `psql` were not found. `docker images` and `docker info` could not connect to the Docker Desktop Linux engine.

### Exact Verification Commands and Results

- `cd packages/core && sqlc generate` — not executable: `sqlc` was not recognized.
- `cd packages/core && atlas migrate hash --dir "file://db/migrations"` — not executable: `atlas` was not recognized.
- `cd packages/core && atlas migrate validate --dir "file://db/migrations"` — not executable: `atlas` was not recognized.
- `cd packages/core && go test ./enrichment` — not executable: `go` was not recognized.
- `cd packages/core && gofmt -w enrichment/product_enrichment.go enrichment/product_enrichment_test.go` — not executable: `gofmt` was not recognized; no formatter changes were made.
- `git status --short`, `git diff --check`, and the pre-update diff/stat checks passed. The corrective implementation files were clean at HEAD before this worklog update; this verification update is the only intended current worktree change.
- No database was started or contacted, and no migration was applied.

### SQLC Generation Outcome

Real SQLC generation did not run because neither the SQLC CLI nor a usable repository-defined container/script containing SQLC is available. The Dockerfile does not contain SQLC, and the Docker daemon is unavailable. Therefore `packages/core/repository/product_enrichment_suggestions.sql.go` and `packages/core/repository/models.go` were not regenerated, no generated diff exists, and compatibility with the real SQLC v1.30.0 generator remains unverified. The source query remains authoritative and was not edited.

The current source/query review confirms the intended operations are represented for create/upsert-get, organization-scoped get, organization/status list, approve, reject, failed, retryable, and applied. Generated parameter types, JSONB/null/timestamp types, IDs, and scan order cannot be accepted as generator-verified without running SQLC.

### Atlas Checksum Outcome

Real Atlas hashing and validation did not run because Atlas is unavailable and no usable local repository-defined execution path exists. `packages/core/db/migrations/20260820000000.sql` exists but is not represented in the current `packages/core/db/migrations/atlas.sum`. `atlas.sum` was not manually edited. Pre-existing migration-integrity status cannot be determined until the real Atlas hash/validate workflow runs.

### Go / Formatting / Test Outcome

Go, gofmt, and the focused enrichment test/compile command were unavailable through every existing repository-defined path. No compilation success is claimed. SQLC generated type/signature mismatches, duplicate types, unused imports, and DTO/test compilation remain unverified.

Static lifecycle review of `packages/core/queries/product_enrichment_suggestions.sql` found the corrective semantics present: failed/retryable clear reviewer, reviewed, and applied timestamps, refresh `updated_at`, and restrict prior states; approve/reject record reviewer/reviewed state and clear `applied_at`; applied is restricted to approved and sets `applied_at`; the query does not mutate product data. This is source inspection only, not an execution result.

### Latest Verification

- Toolchain discovery and documented-command probes completed on 2026-08-20.
- `git diff --check`: passed before this update; it must be rerun after this worklog update.
- No implementation or generated-file changes were made by the verification.
- No remaining CRITICAL source-correction finding is known from this static pass.

### Stage 1 Status

UNDER REVIEW — NEEDS GENERATOR VERIFICATION

### Remaining Blockers

#### IMPORTANT

- Real SQLC v1.30.0 generation and inspection of all generated diffs, especially `product_enrichment_suggestions.sql.go` and `models.go`, are still required.
- Real Atlas hash and validation are still required; `atlas.sum` currently omits `20260820000000.sql`.
- Go formatting, focused enrichment tests, and the smallest practical core compile/test scope remain unverified.

#### MINOR

- The nullable `created_at`/`updated_at` audit columns remain a documented schema convention gap; this verification did not redesign the schema.

### Next Action

Provide an existing usable environment containing the repository-compatible Go, SQLC v1.30.0, and Atlas tools, then run `gofmt` on the two human-authored enrichment files, `cd packages/core && sqlc generate`, inspect generator diffs, `cd packages/core && atlas migrate hash --dir "file://db/migrations"`, `atlas migrate validate --dir "file://db/migrations"`, and the focused/core Go tests. Keep Stage 1 UNDER REVIEW and do not begin Stage 2 until those gates pass.

Final verdict: NEEDS GENERATOR VERIFICATION

## Current Phase

Stage 2 provider-neutral design (read-only design completed; implementation is not authorized).

## Stage 2 Design

### Architectural location and repository patterns

- Stage 2 belongs on the server/core side in `packages/core/enrichment`, with orchestration wired from the core usecase/server layer. It must not change the SAP agent or the deterministic SAP mappings.
- The existing core pattern is `handler -> usecase -> repository`: see `packages/core/usecase/README.md`, `packages/core/usecase/brand_usecase.go` (`NewBrandUseCase`, `SetRepository`, repository-backed methods), and `packages/core/repository/db.go` (`DBTX`, `New`, `WithTx`). The provider-neutral domain contract should follow this separation and keep provider/database types out of the provider interface.
- Server bootstrap and dependency injection are explicit in `apps/cloud-server/main.go`: `setupDatabase`, `setupRouter`, `main`, and the `New...UseCase` calls. Future enrichment construction, configuration, and worker startup should imitate this pattern.
- SAP ingestion is a special master-pool path in `packages/core/usecase/sap_migration.go`: `SAPMigrationUseCase`, `NewSAPMigrationUseCase`, `IngestBatch`, `pgx.Tx`, row savepoints, and the final `tx.Commit`. The `/api/v1/migration` route is registered by `packages/core/routing/sap_migration.go` and handled by `packages/core/handler/sap_migration.go`; it uses `x-tenant-id`/organization ID and does not use the normal tenant repository middleware.
- Normal tenant scoping is implemented by `packages/core/middleware/tenant.go` (`TenantMiddleware`, `MasterRepositoryMiddleware`) and the per-request `repository.New(pool)` convention. The post-SAP enrichment path must use the committed `organization_id` and product queries scoped by organization; it must not infer tenant identity from model input.
- Existing outbound HTTP/retry style is in `apps/sap-agent/internal/transport/client.go` (`CloudClient`, `http.Client` timeout, context-aware requests, exponential retry for 429/5xx). Existing server-side asynchronous work is only domain-specific ZATCA reporting in `apps/cloud-server/internal/zatca/service.go` (`StartReportingWorker`, `processPendingReports`), plus the in-memory backup jobs in `packages/core/usecase/backup_usecase.go`. No generic durable queue or enrichment worker exists.
- Current logging is standard-library logging plus `packages/core/middleware/logger.go`; `SAPMigrationHandler.IngestBatch` logs domain/batch errors. The repository has `audit_logs` in `packages/core/db/schema/80_promotions_loyalty.sql`, but no reusable audit service/query was found. SAP batch staging is the existing migration audit record; Stage 2 must not claim a generic audit framework that is not present.
- Configuration is environment-driven by `packages/core/config/config.go` (`Config`, `LoadConfig`, `getEnv`) and is currently unaware of AI providers. No provider credential belongs in SAP-agent JSON config, product metadata, or suggestion JSON.

### Post-commit boundary

- The safest future hook is immediately after the successful `tx.Commit(ctx)` in `packages/core/usecase/sap_migration.go`, after the deterministic product/UoM/category/brand writes are durable and before `IngestBatch` returns its response.
- The hook should only synchronously create an idempotent, pending enrichment request/suggestion for each eligible committed product. It must not call a model, perform network I/O to a provider, or hold the SAP transaction open.
- A future `ProductEnrichmentCoordinator` may receive the committed organization/product identifiers, reconstruct the snapshot from the database, and call the repository upsert. If this post-commit enqueue fails, the migration response remains successful because SAP synchronization is independent of AI availability; the enqueue failure is logged/observed for reconciliation and must not be converted into a transaction failure.
- Because no generic durable worker exists, Stage 2 also requires a future core/server worker mechanism to claim pending work, call the provider outside a transaction, validate the result, and persist the result. The existing ZATCA worker is not an appropriate shared queue abstraction.

### First-MVP eligibility

The deterministic eligibility function takes a committed product snapshot, source identity, current structured identities, existing suggestion state for the effective fingerprint, and an enrichment-enabled policy. It returns `Eligible`, a skip reason, and the missing-field capabilities (`brand`, `description`, `category`).

- Source must be SAP, `source_item_code` must be non-empty, the product must exist under the supplied organization, and `source_item_name`/`products.name` must be non-empty. Empty names are not useful model evidence.
- The product must have a valid architect-approved `product_type`: `standard`, `raw_material`, `fixed_asset`, or `finished_good`. Invalid/missing type is a deterministic data-quality condition, not an AI repair opportunity. A valid `fixed_asset` remains eligible for missing brand or description.
- Brand is eligible only when the structured product brand is unresolved (`brand_id` is null). A resolved structured SAP brand makes brand enrichment ineligible and forces `KEEP_EXISTING` if a brand proposal is represented.
- Description is eligible only when `description` is null/empty after trimming. A populated description is authoritative for the MVP.
- Category is eligible only when `category_id` is null. A populated structured category is never sent for AI refinement; its proposal must be `KEEP_EXISTING`.
- The request is not eligible merely to extract unsupported semantics. At least one of missing brand, empty description, or absent category must be present.
- The recommended first rollout policy is active products only (`is_active = true`); inactive products are skipped unless a later explicit backfill mode is introduced. `is_sellable`, `is_purchasable`, and `track_inventory` are never AI decisions and are not used to change eligibility semantics.
- An existing row for the same effective input fingerprint and contract is not duplicated when it is `pending`, `processing`, `retryable`, `in_review`, `approved`, `rejected`, or `applied`. A `retryable` row is retried only by the worker policy; permanent `failed` rows require explicit operational retry or a changed input/contract.
- A deterministic approved alias/rule that resolves a field before eligibility makes that field ineligible. Approved aliases are a later phase and are not invented by the model.

### Source snapshot

The coordinator constructs a typed, immutable `EnrichmentRequest` from committed core data. Internal organization/product IDs are correlation-only and are not provider content. The provider-safe snapshot contains:

- source system `SAP`, `source_item_code` as an echo/correlation value, and the exact source item name;
- current description, when present, and immutable `product_type`;
- current category ID/code/name plus hierarchy path when resolved;
- current brand ID/code/name when resolved;
- base UoM ID/code/name and the relevant structured purchase/sales/base UoM names;
- relevant product conversion pairs and factors only as immutable context, with an explicit instruction that the model cannot recalculate or propose them;
- a whitelisted, non-operational SAP provenance subset only when already stored and verified. The full `products.metadata` object is not sent.

The snapshot does not send inventory quantities, warehouse/store/location data, prices, tax rates/categories, supplier identity, barcodes, status flags, variant/family structures, serial/batch data, raw credentials, or unrelated metadata. `product_type`, identifiers, UoM/conversion data, and all structured SAP values are context and validation constraints, never proposal targets.

### Candidate dictionaries

- `brands` and `product_categories` are global in the current schema: neither table has `organization_id`; both use globally unique codes. Products are organization-scoped, but candidate taxonomy is not. Therefore the design must describe candidates as globally active canonical records and must not claim tenant-specific filtering that the schema cannot provide.
- Existing brand access patterns are `ListActiveBrands`, `SearchActiveBrands`, and their generated methods from `packages/core/queries/brands.sql` / `packages/core/repository/brands.sql.go`. Future enrichment queries should return only `id`, `code`, and `name`, with `is_active = true`.
- Existing category hierarchy access is `GetCategoryHierarchy` from `packages/core/queries/products.sql` / `packages/core/repository/products.sql.go`; existing category helpers also appear in `packages/core/queries/product_categories.sql`. Future candidate output should include active category `id`, `code`, `name`, `parent_category_id`, level, and canonical full path.
- The model must receive canonical IDs/codes/names and select from supplied candidates. It must not invent an ID/code. `MATCH_EXISTING` is accepted only after server-side validation that the target ID/code exists, is active, and belongs to the supplied dictionary.
- Candidate retrieval should be bounded. For small taxonomies, provide the complete active dictionary. For large taxonomies, deterministically pre-filter by exact/case-folded token, code, prefix, and safe name search, then cap the supplied list. No fuzzy or transliteration alias table exists today; Arabic/English/transliterated matching is model-assisted only when the canonical candidate is actually supplied. If no safe candidate is supplied, the result can be `NO_MATCH` or review-only `PROPOSE_NEW`, never automatic creation.
- A per-request candidate set hash should be calculated from sorted `(id, code, name, parent/path)` records. Do not use a global taxonomy version that would re-enrich every product after an unrelated record is added.

### Provider-neutral interface

The conceptual interfaces belong in `packages/core/enrichment` and expose only standard Go/domain types:

```go
type ProductEnrichmentProvider interface {
    Enrich(ctx context.Context, request EnrichmentRequest) (ProviderResult, error)
}

type ProductEnrichmentService interface {
    Enrich(ctx context.Context, request EnrichmentRequest) (EnrichmentResult, error)
}

type ProductEnrichmentCoordinator interface {
    EnqueueCommittedProduct(ctx context.Context, organizationID, productID int32) error
    ProcessPending(ctx context.Context, limit int32) error
}
```

`EnrichmentRequest` contains correlation identity, `SourceSnapshot`, candidate dictionaries, contract version, effective fingerprint, and field capabilities. `ProviderResult` contains the strict `ProposalSet` plus provider/model metadata. `EnrichmentResult` distinguishes accepted no-match/proposals from validation failure. Provider implementations must not receive `repository.Queries`, write databases, mutate products, or expose OpenAI/Anthropic SDK types.

Provider errors are classified as retryable (`context deadline`, network, 429, 5xx, temporary provider outage) or permanent (`invalid request`, unsupported configuration, malformed/unknown-field output, unknown candidate target, prohibited field, invalid action/confidence, or schema validation failure). Context cancellation and timeout must propagate through the interface; the adapter owns provider-specific timeout and retry policy.

### Strict first-MVP model output contract

The accepted top-level object contains only the correlation echo and `brand`, `category`, `description`, and `unsupported_semantics`. JSON decoding must reject unknown fields and size-limit the response; prohibited fields are rejected, never silently ignored. `source_item_code` must exactly echo the request value.

- Brand and category proposals use only `KEEP_EXISTING`, `MATCH_EXISTING`, `PROPOSE_NEW`, or `NO_MATCH`; each has confidence `[0,1]`, evidence, and a short explanation. `MATCH_EXISTING` requires a supplied canonical `target_id` or `target_code` and server-side candidate validation. `PROPOSE_NEW` has canonical name/value for review but no existing target identity.
- A populated structured brand or category permits only `KEEP_EXISTING`. An absent category may use `MATCH_EXISTING`; an absent brand may use `MATCH_EXISTING`, `PROPOSE_NEW`, or `NO_MATCH`.
- Description uses `KEEP_EXISTING`, `PROPOSE_NEW`, or `NO_MATCH`; `MATCH_EXISTING` is invalid for free text. `PROPOSE_NEW` requires proposed text, confidence, evidence, and explanation.
- `unsupported_semantics` is an array of `semantic_type`, `key`, JSON value, confidence, evidence, and explanation. It is evidence only and is not applied to any product column or `products.metadata`.
- The output schema contains no `product_type`, SKU/ItemCode proposal, barcode, inventory, warehouse, price, tax, UoM/conversion proposal, supplier, status flag, variant, family, or normalized-attribute field. A response containing any such field, including nested/camelCase variants, is a permanent invalid-output/validation error.
- The current Stage 1 `UNSUPPORTED_TARGET` constant is not an accepted model field action in this MVP. A provider response attempting to use it or any prohibited target is rejected as invalid output rather than treated as a valid suggestion.

### Description, brand, category, and unsupported-semantic policies

- Descriptions are factual catalog text, not marketing copy. Maximum proposed length is 500 Unicode characters, with whitespace normalized and no invented specifications, benefits, certifications, compatibility, quantities, or claims. Evidence must be traceable to the supplied ItemName/current description; ambiguous names yield `NO_MATCH` rather than a guess.
- Output should remain in the source language/script. Arabic names remain Arabic; English remains English; mixed names may produce concise mixed-language catalog text. Translation is not an MVP requirement. Mojibake must not be “corrected” into a confident brand/description without reliable evidence.
- `HIKvision DS 2.8mp (20@208.56)` may yield only an evidence-grounded description such as the visible product identity; it must not invent camera technology, resolution semantics beyond the literal text, or model specifications.
- Brand normalization is deterministic at the boundary (trim, Unicode normalization, case/separator normalization, controlled token comparison) and model-assisted for Arabic/English/transliterated strings. `Ø¨Ø§Ù†ØªÙŠÙ†` may match a supplied canonical Pantene candidate, but the persisted proposal records the raw evidence, canonical candidate ID/code/name, and explanation. `PROPOSE_NEW` is review-only; confidence never creates a brand. Approved aliases are a later deterministic phase.
- If a category is populated, category action is always `KEEP_EXISTING`. Semantic words such as Shampoo may be recorded as unsupported evidence, but the model must not refine a populated category into Hair Care. If category is absent, `MATCH_EXISTING` may select only a supplied active hierarchy candidate; `PROPOSE_NEW` is review-only and no category is created automatically.
- Supported evidence examples include `anti_dandruff`, `size_text = 400 ml`, `capacity = 4TB`, `resolution_text = 2.8mp`, `dimensions`, `model_number`, and `family_hint`. Packaging text such as `24*400 ml` may be preserved as evidence, but SAP conversion factors remain authoritative and are never recalculated.

### Fingerprint and idempotency

- Canonicalize input as deterministic JSON with sorted object keys, explicit null/empty representation, UTF-8 Unicode normalization, normalized line endings, trimmed scalar whitespace, and stable sorted conversion/candidate arrays. Do not lowercase or transliterate source evidence in the source hash; source text changes must remain visible.
- The source component contains `source_system`, `source_item_code`, `source_item_name`, current description, `product_type`, resolved category identity, resolved brand identity, base/purchase/sales UoM identities, and immutable conversion context. It excludes inventory, warehouse, price, tax, supplier, flags, timestamps, unrelated metadata, and volatile operational quantities.
- Compute a SHA-256 hex digest for the canonical source component. The contract version is a separate explicit value for the output/prompt/schema contract.
- Compute a second candidate-set digest from the exact bounded brand/category records supplied to this request. Since the current Stage 1 uniqueness key has only `source_data_fingerprint`, the Stage 2 implementation must either first add a candidate fingerprint column or define/document an effective input fingerprint that combines the source digest and this per-request candidate digest while recording both components in the structured snapshot. A global taxonomy hash must not be used.
- The uniqueness operation remains organization + product + effective fingerprint + contract. Same relevant ItemName/context and same candidate set is a no-op; changing ItemName, description, structured brand/category identity, product type, or relevant UoM/conversion context creates a new effective input. Inventory/price-only changes do not.

### Error, retry, and review lifecycle

The current six statuses are not sufficient. `pending` currently conflates “not yet inferred” with “inference completed and awaiting human review”; the schema has no `in_review` or `processing` state, no claim/lease fields, and the current SQL has no operation to atomically claim work or persist a completed proposal set.

Recommended future lifecycle correction before Stage 2 implementation:

```text
pending -> processing -> in_review -> approved -> applied
             |              |
             +-> retryable   +-> rejected
             +-> failed
retryable -> processing
```

- `pending` means queued/not yet inferred; `processing` is a worker claim/execution state (or an equivalent durable lease); `in_review` means validated inference completed, including a successful `NO_MATCH`; `approved`/`rejected` are human review states; `applied` is reserved for a later explicit application operation; `failed` is permanent inference/validation failure; `retryable` is transient failure awaiting retry.
- Approval/rejection must accept only `in_review` (not an un-inferred `pending` row). Applying must accept only `approved`. Retry transitions must clear reviewer/application fields and be organization-scoped. No Stage 2 path mutates products, brands, categories, or `products.metadata`.
- The minimum semantic schema correction is a distinct `in_review` state. A production-safe asynchronous worker additionally needs `processing` or a durable claim/lease mechanism plus attempt/backoff information. This correction is a prerequisite, not Stage 2 implementation work.

### Human review recommendation

For the first production rollout, require review for every AI result, including `MATCH_EXISTING` brand proposals. Do not auto-apply `PROPOSE_NEW`; do not auto-create brands/categories; require review for descriptions; and allow category changes only when the structured category is absent. This is a rollout recommendation, not an architect-approved auto-apply policy. Stage 2 itself performs no product mutation.

### Security and privacy boundary

- Future provider enablement, provider/model, timeout, retry limit, and server-side credential must be configured in `packages/core/config/config.go` from server environment variables. Suggested future knobs are `ENRICHMENT_ENABLED`, `ENRICHMENT_PROVIDER`, `ENRICHMENT_MODEL`, `ENRICHMENT_TIMEOUT_SECONDS`, and `ENRICHMENT_MAX_RETRIES`; region/base URL is optional and only justified for a selected provider.
- Provider credentials are server secrets only. They must never be present in SAP-agent `agent_config.json`, UI configuration, product metadata, suggestion JSON, audit payloads, or logs.
- Send only the whitelisted snapshot needed for catalog enrichment. Do not send prices, inventory, tax, supplier, warehouse, barcode, status, or unrelated metadata. Log only organization/product IDs, suggestion ID, status, provider/model names, latency, attempt, and redacted error class. Never log ItemName, description, evidence, candidate dumps, provider request/response bodies, authorization headers, or secrets.
- Existing `audit_logs` may later record human review actions with organization/reviewer identity, but no generic audit service was found and no such implementation is assumed by this design.

### Exact future test plan

- Unit tests for eligibility reasons/capabilities, active/inactive policy, fixed assets, structured brand/category precedence, snapshot allowlist/redaction, candidate matching and canonical-target validation, strict action/output validation, prohibited nested/camelCase fields, unsupported semantics, description length/language/safety rules, and source/candidate fingerprint stability/change cases.
- Provider contract tests for timeout/network/rate-limit classification, malformed JSON, unknown fields/actions, confidence outside `[0,1]`, unknown canonical IDs/codes, prohibited fields, oversized output, and successful `NO_MATCH`; no live model calls in normal tests.
- Usecase/coordinator tests for post-commit-only behavior, provider unavailable without SAP migration failure, idempotent same-fingerprint enqueue, changed relevant ItemName creating a new request, inventory/price-only changes not re-enriching, and structured brand preventing AI replacement.
- Repository/integration tests for organization-scoped product reads, global active candidate dictionaries, atomic claim/transition behavior, status-field cleanup, uniqueness, and review isolation. These require the real SQLC/Atlas toolchain before acceptance.

### Exact future patch surface

- First, correct and verify Stage 1 schema/query lifecycle in `packages/core/db/schema/35_product_enrichment.sql`, `packages/core/db/migrations/20260820000000.sql`, `packages/core/queries/product_enrichment_suggestions.sql`, and regenerated SQLC files under `packages/core/repository/`; regenerate/validate Atlas and SQLC before Stage 2 implementation.
- Add focused provider-neutral modules under `packages/core/enrichment` for request/snapshot, eligibility, candidate types, fingerprinting, service/coordinator interfaces, validation, and tests. Keep provider-specific adapters outside the core contract (likely under `apps/cloud-server/internal/enrichment/` once a provider is selected).
- Add focused SQLC source queries for product snapshot/conversion lookup, bounded active brand candidates, active category hierarchy candidates, and durable work claim/completion; generated repository output must come from the repository's SQLC workflow.
- Make only small wiring changes to `packages/core/usecase/sap_migration.go` (inject coordinator and call it after `tx.Commit`), `apps/cloud-server/main.go` (construct/configure coordinator and future worker), and `packages/core/config/config.go` (server-only feature/provider settings). Review endpoints and application of approved proposals are later work.
- Do not change `packages/sap/**` or `apps/sap-agent/**`; repository evidence does not make those changes necessary.

## Stage 1 Verification Blockers

- Real SQLC v1.30.0 generation is unavailable; the checked-in repository companions remain provisional and are not generator-verified.
- Atlas checksum generation/validation is unavailable; `packages/core/db/migrations/atlas.sum` still omits `20260820000000.sql`.
- Go, `gofmt`, and focused/core tests are unavailable in the current environment.
- Stage 1 remains UNDER REVIEW / NEEDS GENERATOR VERIFICATION and is not SAFE TO KEEP. No Stage 2 implementation or provider selection has been performed.

## Open Questions

- The Stage 1 lifecycle correction must be architect-approved: at minimum add a distinct inference-review state; decide whether the worker uses a durable `processing` status or lease columns, and how candidate fingerprint components are persisted without global re-enrichment.
- The source/rule for `finished_good` remains unestablished.
- AI provider/model, region, retention policy, server deployment boundary, reviewer roles/permissions, and final review API are not selected.
- Active-only enrichment versus an explicit inactive-product backfill remains a rollout choice.
- Live tenant taxonomies are not guaranteed by seed data; current brand/category tables are global, not organization-scoped.
- Automatic application policy and future attribute/family schema remain unapproved.

## Next Action

Because Stage 2 design exposes a substantive Stage 1 lifecycle/status defect, correct and architect-review the Stage 1 status/claim/completion model before any Stage 2 implementation. Then complete the existing toolchain gates: `gofmt`, focused/core Go tests, `cd packages/core && sqlc generate` with generated-diff inspection, `atlas migrate hash --dir "file://db/migrations"`, and `atlas migrate validate --dir "file://db/migrations"`. Do not begin provider integration or model calls before those gates pass.

## Stage 1 Lifecycle Correction and Host Verification - 2026-08-20

### Tool version compatibility

- Repository Go expectation is Go `1.25.0` in every module, with `SETUP_GUIDE.md` accepting `1.25.0` or higher; CI uses Go `1.25.x`.
- Repository SQLC is explicitly pinned to v1.30.0 in `.github/workflows/db-validate.yml`; checked-in generated repository files also identify SQLC v1.30.0. The host has SQLC v1.31.1, so generation was intentionally not run because output compatibility is not established across the pinned version boundary.
- Repository CI installs Atlas through `ariga/setup-atlas@v0` with `version: latest`; the Dockerfile also downloads the latest Atlas binary. No fixed Atlas version was found. The host Atlas is `v1.3.2-4bf5fb9-canary`.
- Host Go used was `go1.26.7 windows/amd64`. This is newer than, and compatible with, the module `go 1.25.0` directives. No dependency or tool version files were changed.

### Lifecycle correction

The old six-state lifecycle was `pending -> approved/rejected`, with `pending` also serving as the unclaimed and reviewable state; failed/retryable transitions were not aligned with active provider processing. The corrected eight-state lifecycle is:

```text
pending -> processing -> in_review -> approved -> applied
                         |             \\-> rejected
                         +-> retryable
                         +-> failed
retryable -> processing
```

- `pending -> processing` and `retryable -> processing` are atomic, organization-scoped claims whose `WHERE` clause prevents two concurrent claims from succeeding.
- `processing -> in_review` persists validated provider output and clears reviewer/reviewed/applied fields.
- `processing -> retryable` and `processing -> failed` clear reviewer/reviewed/applied fields. Permanent failure is reached from active processing only; retryable work must be claimed again before a terminal failure.
- `in_review -> approved` and `in_review -> rejected` record reviewer identity/time and leave `applied_at` NULL.
- `approved -> applied` records `applied_at` and preserves reviewer/reviewed fields.
- No transition is allowed from `failed`, `rejected`, or `applied`; review cannot bypass `in_review`.

### Files changed by this correction

- `packages/core/db/schema/35_product_enrichment.sql`: status constraint now contains all eight statuses.
- `packages/core/db/migrations/20260820000000.sql`: semantically matching eight-status constraint.
- `packages/core/queries/product_enrichment_suggestions.sql`: added processing claim and in-review completion operations; constrained approve/reject/failed/retryable transitions to the corrected lifecycle; preserved idempotent upsert behavior and organization scoping.
- `packages/core/enrichment/product_enrichment.go`: added the eight `SuggestionStatus` constants, validity check, and provider-neutral transition guard; existing proposal validation and SAP-authoritative-field rules remain intact.
- `packages/core/enrichment/product_enrichment_test.go`: added all-eight-status, invalid-status, valid-transition, and invalid-transition coverage; existing brand/category precedence, prohibited-target normalization, unsupported-evidence, product_type, and confidence tests remain.
- `packages/core/db/migrations/atlas.sum`: regenerated by Atlas.
- `agents.md`: this handoff entry.
- `packages/core/repository/product_enrichment_suggestions.sql.go` and `packages/core/repository/models.go` were not manually edited or regenerated because the repository requires SQLC v1.30.0 while only v1.31.1 is installed.

### SQLC result

- Repository command: `cd packages/core && sqlc generate`.
- Result: NOT RUN. The CI workflow explicitly installs v1.30.0, while the only installed binary is v1.31.1. Running it would violate the repository generator-version gate and could create unrelated generated drift.
- Consequently, no material generated-output comparison is claimed. The prior hand-maintained `product_enrichment_suggestions.sql.go` remains provisional and does not yet contain generated methods for the new processing/in_review source queries. Compatible SQLC v1.30.0 generation and generated-diff inspection remain required.

### Atlas result

- Commands run with the discovered host binary `C:\Users\AnnsMustafa\atlas\atlas.exe`:
  - `atlas migrate hash --dir "file://packages/core/db/migrations"`
  - `atlas migrate validate --dir "file://packages/core/db/migrations"`
- Both succeeded without applying migrations or contacting an external database.
- Atlas regenerated `packages/core/db/migrations/atlas.sum`; `20260820000000.sql` is included. The installed canary rewrote the existing checksum lines as well as adding the new migration, and validation passed against the regenerated manifest.

### Go and static verification

- `C:\Program Files\Go\bin\gofmt.exe -w packages/core/enrichment/product_enrichment.go packages/core/enrichment/product_enrichment_test.go` succeeded; no files were reported by the subsequent `gofmt -l` check.
- `cd packages/core && go test ./enrichment` passed.
- `cd packages/core && go test ./repository` passed compilation; the package has no tests.
- `cd packages/core && go test ./...` passed for all core packages.
- Static schema/query assertions passed for matching eight-state constraints, every required transition predicate, and organization scoping. `git diff --check` passed.

### Scope and acceptance state

- No files under `packages/sap/**` or `apps/sap-agent/**` changed.
- No AI provider, prompt, API route, worker loop, post-ingestion hook, UI, product mutation, or `products.metadata` enrichment write was added. `product_type` remains architect-controlled and structured SAP brand/category precedence remains intact.
- Candidate/taxonomy fingerprinting remains OPEN for Stage 2. The existing Stage 2 read-only design remains documentation only; Stage 2 implementation did not begin.
- Remaining IMPORTANT blocker: compatible SQLC v1.30.0 generation and inspection of `product_enrichment_suggestions.sql.go` and `models.go`. Atlas, formatting, compilation, focused tests, full core tests, and diff checks passed.
- Current Stage 1 status: NEEDS GENERATOR VERIFICATION. It cannot be marked SAFE TO KEEP until the repository-pinned SQLC generation succeeds and the resulting generated repository code is inspected.

## Final Stage 1 Generator Verification and Review - 2026-08-20

### SQLC v1.30.0 generation

- The installed binary was `C:\Users\AnnsMustafa\go\bin\sqlc.exe`; `sqlc version` reported exactly `v1.30.0`, matching the repository CI pin.
- The exact repository command was run from `packages/core`: `sqlc generate`.
- Generation succeeded. No generated file was manually edited after generation.
- Stage 1 generated output is authoritative in `packages/core/repository/product_enrichment_suggestions.sql.go` and `packages/core/repository/models.go`.
- The real output includes create/upsert-get, scoped get, scoped list, processing claim, in-review completion, approve, reject, retryable, failed, and applied methods. It expands `SELECT *`/inline scans into explicit ordered columns, uses `int32` IDs, `json.RawMessage` JSONB fields, `pgtype.Text` for nullable provider/model/model_version, and `pgtype.Int4`/`pgtype.Timestamp` for nullable reviewer/timestamps. No unused imports or duplicate `ProductEnrichmentSuggestion` type remain.
- Compared with the provisional hand-maintained companion, the generated file adds the processing/in-review methods, uses the actual v1.30.0 header, expands all returning/scanned columns explicitly, removes the shared scanner, and follows the repository's generated formatting/conventions.
- SQLC also exposed unrelated pre-existing generated drift: `models.go` added existing `CustomerAddress` and SAP staging model types and regenerated the `User` fields; `users.sql.go` added existing `must_reset_password`/`sap_imported` columns to its query returns/scans. These changes are normal output from the pinned generator against existing source schema/queries, are outside Stage 1 behavior, and were retained rather than suppressed. They should be reviewed separately as repository-wide generated drift.
- The generated `users.sql.go` initially preserved trailing spaces from the existing `packages/core/queries/users.sql`; those two source spaces and the missing final newline were cleaned in the source query, then SQLC v1.30.0 was rerun. This was the smallest unrelated hygiene correction required for `git diff --check`; generated output was not manually edited.

### Go verification

- `C:\Program Files\Go\bin\gofmt.exe -w packages/core/enrichment/product_enrichment.go packages/core/enrichment/product_enrichment_test.go` succeeded.
- `gofmt -l` on both human-authored files returned no files.
- `cd packages/core && go test ./enrichment` passed.
- `cd packages/core && go test ./repository` passed; the package has no test files.
- `cd packages/core && go test ./...` passed for all core packages.
- After the source-query hygiene correction and second SQLC generation, `./enrichment`, `./repository`, and `./...` were rerun and passed again.
- Go used was `go1.26.7 windows/amd64`, compatible with the repository's Go 1.25 module requirements.

### Atlas checksum investigation

- Atlas used was `v1.3.2-4bf5fb9-canary` at `C:\Users\AnnsMustafa\atlas\atlas.exe`.
- The exact commands were run from `packages/core`: `atlas migrate hash --dir "file://db/migrations"` and `atlas migrate validate --dir "file://db/migrations"`. Both passed; no migration was applied and no database was contacted.
- `20260820000000.sql` is present in `atlas.sum` with checksum `h1:KPjere8EtihFYVQ9+OZUiDilOp9DwW+lR9SKUfJVE0k=`.
- The canary rewrote the root checksum and all four historical file entries, not only the new migration entry. The four historical migration blobs are byte-identical to HEAD, contain LF-only line endings, and `.gitattributes` enforces LF for SQL and sum files. Therefore the churn is not caused by changed historical migration content or Windows line endings.
- Repository CI uses `ariga/setup-atlas@v0` without a fixed version, and the Dockerfile also downloads Atlas latest. The available repository evidence therefore identifies the old-entry rewrite as Atlas checksum implementation/version behavior relative to the previously recorded manifest; the exact prior Atlas build is not recorded. The current canary-generated manifest validates successfully and is compatible with the project's unpinned latest-Atlas CI convention.
- Repository history confirms Atlas checksum manifests have previously been regenerated when migration bytes changed for LF/BOM normalization. That historical explanation does not apply to the unchanged files in this run.

### Lifecycle and schema verification

- Final lifecycle: `pending -> processing`, `retryable -> processing`, `processing -> in_review | retryable | failed`, `in_review -> approved | rejected`, and `approved -> applied`.
- Processing claims are organization-scoped and atomically guarded by `status IN ('pending', 'retryable')`; concurrent workers cannot both update the same row successfully under normal PostgreSQL update semantics.
- Create/upsert conflict identity is organization + product + source fingerprint + contract version. Its conflict update only preserves the existing row's lifecycle state and timestamps; it does not reset processing, in-review, approved, rejected, or applied state.
- In-review, retryable, and failed transitions clear reviewer/reviewed/application fields. Approval/rejection record reviewer and review time and leave `applied_at` NULL. Applying is restricted to approved, preserves reviewer fields, and sets `applied_at`.
- The schema and migration are semantically equal after removing migration-only `IF NOT EXISTS` guards/comments. Both contain the same eight statuses; no new schema fields were added.

### Domain and scope verification

- Focused enrichment tests pass. `product_type` and case/separator variants such as `productType` and `product-type` remain prohibited; ItemCode/source-item-code/SAP-item-code, UoM/conversion, operational status, supplier, pricing, tax, inventory, warehouse, and barcode ownership remain protected.
- Structured category precedence and resolved-brand `KEEP_EXISTING` precedence remain enforced. Unresolved brands support review-only canonical matching/new proposals. Confidence accepts only `[0,1]`, including finite boundary values; unsupported evidence such as anti-dandruff and capacity remains evidence-only.
- Changed paths are limited to `agents.md`, Stage 1 SQL/schema/migration/query/domain/test files, the whitespace-only hygiene correction in `packages/core/queries/users.sql`, and SQLC-generated repository files. No `packages/sap/**`, `apps/sap-agent/**`, SAP behavior, provider, prompt, route, worker, post-commit hook, UI, product mutation, `products.metadata`, variant/family, or candidate/taxonomy fingerprint implementation changed.
- The existing Stage 2 provider-neutral design remains documentation only. Stage 2 implementation did not begin.

### Final Stage 1 acceptance

- `git diff --check` passed after generation and formatting.
- All requested generator, Go, lifecycle, migration parity, Atlas hash/validate, domain safety, and scope gates passed.
- Remaining non-blocking review notes: Atlas is not version-pinned by repository CI, so checksum regeneration is tool-version-sensitive; SQLC revealed separate repository-wide generated drift in `models.go`/`users.sql.go` that should be reviewed independently.
- Current Stage 1 status: SAFE TO KEEP.
- Next Action: Stage 2 provider-neutral enrichment implementation planning, starting from the existing read-only design. Do not begin provider implementation without a separately authorized Stage 2 task.

## Stage 2A Pre-implementation Trace - 2026-08-20

- Scope is limited to the deterministic product-enrichment enqueue foundation. The repository trace is complete: the hook belongs after `tx.Commit(ctx)` in `packages/core/usecase/sap_migration.go`; source context and idempotent persistence reuse core repository queries and the existing Stage 1 suggestion table.
- No schema/migration, SAP-agent, deterministic mapping, product mutation, provider SDK, prompt, credential, worker, or UI work is authorized or planned in this stage.
- The resulting implementation decisions, exact files, and verification outcome are recorded in the completed Stage 2A section below.

## Stage 2A — Enrichment Enqueue Foundation

### Architecture and contracts

- Stage 2A is implemented on the server/core side. `packages/sap/**` and `apps/sap-agent/**` remain deterministic and unchanged.
- `packages/core/enrichment/enqueue.go` adds `ProductEnrichmentProvider` as a contract-only interface, `EnrichmentRequest`/`EnrichmentResult`, `EnrichmentSourceSnapshot`, typed product types/gaps/reasons, candidate dictionary types, `EnrichmentStore`, `ProductEnrichmentCoordinator`, and deterministic eligibility, structured-current, and fingerprint functions.
- `packages/core/repository/product_enrichment_store.go` adapts existing generated product, brand, category, UoM, conversion, and Stage 1 suggestion queries. No duplicate SQL, schema, migration, or SQLC-generated file was added.
- `packages/core/usecase/sap_migration.go` calls the optional coordinator only after the existing successful `tx.Commit(ctx)`. `apps/cloud-server/main.go` wires the repository adapter and coordinator without provider configuration.

### Eligibility contract

- Eligibility is restricted to the SAP migration path, a resolved organization/product, non-empty SAP ItemCode/source item code, non-empty ItemName/source item name, and one of exactly `standard`, `raw_material`, `fixed_asset`, or `finished_good`.
- Deterministic gaps are typed as `missing_brand`, `missing_description`, and `missing_category`. A resolved structured brand or category is not a gap; non-empty/trimmed description is not a gap.
- Invalid product types and missing fundamental source values are rejected/skipped with structured reasons. `fixed_asset` and `raw_material` remain eligible when an MVP gap exists. Active/sellable/purchasable flags are not eligibility inputs.

### Safe source snapshot and exclusions

- The allowlisted snapshot contains SAP source identity/item code/item name, current description, immutable product type, structured brand/category identity and category path, base UoM identity, and existing UoM conversion identities/factors as immutable context.
- It excludes inventory/stock, warehouse/store quantities, prices, tax, suppliers, barcodes, credentials, arbitrary `products.metadata`, audit/user/customer data, operational status flags, variants/families, and all proposal targets for authoritative fields.
- `structured_current` is persisted in the Stage 1 JSONB column outside `products.metadata`; no product/category/brand mutation is performed.

### Fingerprint and candidate dictionaries

- Fingerprint contract is `sap-product-enrichment-v1`, SHA-256 hex over canonical JSON containing SAP identity, source item code/name, description, product type, structured brand/category identity, and relevant UoM context. Conversion pairs are sorted before hashing; timestamps, provider/model/version, reviewer state, candidates, inventory, price, tax, supplier, barcode, warehouse, and status flags are excluded.
- Candidate dictionaries remain globally scoped because the current `brands` and `product_categories` schema is global. The adapter reuses `ListActiveBrands` and `GetCategoryHierarchy`, returns only bounded `id`/`code`/`name` plus category path, and uses a default limit of 100. Large-taxonomy retrieval/ranking is deferred to Stage 2B.
- Candidate-set fingerprinting/re-enrichment remains OPEN. Adding an unrelated global candidate does not change this product source fingerprint; a future explicit candidate-version or re-enrichment mechanism is required if candidate changes must trigger work.

### Idempotent enqueue and post-commit isolation

- Eligible products create/get a Stage 1 `source = ai`, `status = pending` suggestion keyed by organization, product, source fingerprint, and contract version. Proposed/provider/reviewer/application fields remain null. Existing rows and lifecycle/proposal state are preserved on conflict.
- Enqueue failures are logged with organization and source-item identifiers after commit and do not change the committed SAP response or attempt rollback. A nil coordinator preserves previous migration behavior.
- No provider/model execution, worker, retry timer, prompt, API key, remote HTTP call, or application of approved suggestions exists in Stage 2A.

### Current lifecycle

The Stage 1 lifecycle remains unchanged:

```text
pending -> processing -> in_review -> approved -> applied
                         |             \\-> rejected
                         +-> retryable
                         +-> failed
retryable -> processing
```

### Files changed and verification

- Stage 2A business/source files: `packages/core/enrichment/enqueue.go`, `packages/core/enrichment/enqueue_test.go`, `packages/core/enrichment/product_enrichment.go` (description validation), `packages/core/repository/product_enrichment_store.go`, `packages/core/usecase/sap_migration.go`, `apps/cloud-server/main.go`, and this worklog.
- SQLC generator output: none. No new SQL query or schema/migration was added, so SQLC generation and Atlas checksum regeneration were intentionally not run. Existing pinned SQLC/Atlas history remains unchanged.
- Pre-existing unrelated generated drift: none changed by Stage 2A; the separate Stage 1 SQLC drift remains historical and outside this patch.
- Passed: Go 1.26.7 `gofmt`; `cd packages/core && go test ./enrichment`; `go test ./repository`; `go test ./usecase`; `go test ./...`; and `cd apps/cloud-server && go test ./...`. `git diff --check` passed. No live provider calls or database migration were run.
- The actual transaction-timing integration test was not added because the current SAP migration usecase has no transaction fake seam; focused coordinator tests cover eligibility, fingerprint stability/change, idempotent preservation, ineligible skip, store failure, and description policy. This is a future integration-test item, not a provider or lifecycle redesign.

## AI MVP Boundary

- `product_type` is never AI; only the four architect-approved values are accepted as immutable context.
- Populated structured category and resolved structured brand remain authoritative. Empty descriptions may receive future review-only proposals.
- No operational SAP field inference or mutation is allowed; proposals remain separate from `products.metadata` and direct product writes.

## Open Questions

- Candidate/taxonomy fingerprinting and explicit re-enrichment strategy remain OPEN.
- The source/rule for `finished_good` remains OPEN unless architect-defined elsewhere.
- Provider/model/region/retention policy is still not selected.
- Reviewer roles/permissions and auto-apply policy remain unresolved/unapproved.
- Future product attribute/family schema is not approved.
- The unrelated pinned-SQLC repository drift recorded in Stage 1 requires separate review if it remains relevant.

## Remaining Phases

- Stage 2B provider adapter and strict structured response parser.
- Stage 2C processing worker/retry execution.
- Stage 2D human review APIs.
- Stage 2E application of approved suggestions.
- Deterministic approved aliases/rules.
- UI/reviewer workflow if required.
- Production security, observability, and hardening.

## Current Stage 2A Verdict

- Stage 1 remains SAFE TO KEEP.
- Stage 2A is SAFE TO KEEP: it creates only deterministic idempotent pending work after SAP commit, preserves organization scoping and structured precedence, and contains no provider/model execution.
- Next Action: begin Stage 2B only under a separately authorized task, after selecting provider/model/region/retention and keeping the existing Stage 1 lifecycle unchanged.

## Stage 2B — Strict Model Contract

### Scope and architecture

- Stage 1 and Stage 2A remain SAFE TO KEEP. The existing SAP-agent -> deterministic mappings -> batch -> core migration architecture remains authoritative.
- Stage 2B stops at a validated provider-neutral enrichment result. It does not select or integrate a provider, call a model/API, run a worker, update suggestion lifecycle status, mutate products, create taxonomy, add routes, add credentials, or change schema/SQL.
- The provider interface remains `ProductEnrichmentProvider.Enrich(context.Context, EnrichmentRequest) (EnrichmentResult, error)`. A future adapter must extract its structured model content and call the core parser; provider wrapper formats and provider metadata are not accepted by the core parser.

### Provider-neutral request DTO

- `EnrichmentRequest` now carries only internal correlation IDs (excluded from JSON), contract/request versions, `source_item_code`, the allowlisted `EnrichmentSourceSnapshot`, exact deterministic Stage 2A gaps, bounded canonical brand/category candidates, and `EnrichmentRequestPolicy`.
- The snapshot retains only approved source name/description/product type, structured brand/category identity/path, and safe UoM context. It excludes inventory, warehouse/store quantities, prices, tax, suppliers, barcodes, arbitrary metadata, credentials, audit/user/customer data, operational status flags, and SAP document data.
- `NewEnrichmentRequest` and `EnrichmentRequest.Validate` enforce SAP source, one of exactly `standard`, `raw_material`, `fixed_asset`, or `finished_good`, source-code correlation, exact deterministic gaps, the strict policy, bounded candidate dictionaries, and candidate identity validity.

### Strict response DTO and parser

- The model response contains only `source_item_code`, `brand`, `category`, `description`, and `unsupported_semantics`. The established Stage 1 `ProposalSet`, proposal types, and action vocabulary are reused; no provider/model/version metadata is accepted as model-authored truth.
- `ParseEnrichmentResponse` / `ParseEnrichmentResponseString` use the standard-library JSON decoder with `DisallowUnknownFields` at the top level and every defined nested proposal/semantic object. Required fields, one JSON value, malformed JSON, trailing values, action values, finite `[0,1]` confidence, structural proposal rules, and exact correlation are validated.
- Textual fields are trimmed where domain-safe. No malformed model output is repaired. Canonical candidate names/codes are replaced with the server-owned candidate record after exact ID/code reconciliation; model names are never used for database lookup.
- Standard-library decoding does not reject duplicate JSON object keys. This is a documented known parser limitation; no parsing dependency was added solely for duplicate-key detection. Future adapters should emit canonical single-key JSON before invoking this boundary.
- Parser errors are classified as `malformed_response`, `contract_violation`, `candidate_mismatch`, `prohibited_output`, or `correlation_mismatch`. Network/rate-limit/timeout retry classification remains a later provider/worker concern.

### Precedence and field policy

- Resolved structured brand accepts only `KEEP_EXISTING`; `MATCH_EXISTING`, `PROPOSE_NEW`, and `NO_MATCH` are rejected. An unresolved brand may `MATCH_EXISTING` only against an exact supplied candidate, may submit a review-only `PROPOSE_NEW`, or may return `NO_MATCH`.
- Populated structured category accepts only `KEEP_EXISTING`; it cannot be refined or replaced. An absent category may use exact candidate matching, review-only `PROPOSE_NEW`, or `NO_MATCH`.
- Description is factual catalog text, source-language preserving by default, review-only, and at most 500 Unicode code points. A populated description accepts only `KEEP_EXISTING`; a missing description accepts `PROPOSE_NEW` or `NO_MATCH`, never `MATCH_EXISTING`.
- `unsupported_semantics` preserves evidence only for schema-unsupported concepts such as shampoo, anti-dandruff, hair type, size/capacity, dimensions/resolution, model number, packaging, and family hints. It cannot target product type, SKU/ItemCode/source identity, barcodes, inventory, warehouse/store, price, tax, UoM/conversion, suppliers, active/sellable/purchasable flags, or SAP document identity. Case, camelCase/PascalCase, spaces, hyphens, underscores, and documented aliases are normalized fail-closed.
- Unsupported semantic values remain informational JSON and never mutate product columns. Packaging such as `24*400 مل` may remain evidence (`packaging_text`, `size_text`); it cannot become a conversion factor.

### Model instruction contract

- `StrictInferenceInstructions` / `BuildInferenceInstructions` specify one exact JSON object, no markdown/wrappers/provider metadata, structured SAP precedence, immutable product type, operational-field prohibitions, populated-category/resolved-brand `KEEP_EXISTING`, exact candidate-only matching, review-only new taxonomy, factual source-language descriptions, no invented specifications, and evidence-only unsupported semantics.
- The instruction contract contains no provider-specific temperature, response format, tool, credential, SDK, endpoint, region, retention, or wrapper setting.

### Real sample contract tests

- `strict_contract_test.go` covers the Arabic Pantene sample `INV00006`, anti-dandruff and packaging/conversion rejection, HIKvision `fixed_asset`, Huawei resolved-brand precedence, and Epson taxonomy non-invention. It also covers malformed/trailing/unknown JSON, nested unknown fields, action/confidence/correlation failures, candidate ambiguity and canonical reconciliation, all brand/category/description precedence cases, Unicode limits, allowed evidence, and prohibited aliases/values.
- No live model or remote API call exists in the tests.

### Files changed and verification

- Stage 2B files: `packages/core/enrichment/enqueue.go` (request/result contract extension), `packages/core/enrichment/product_enrichment.go` (prohibited-value recursion/aliases), `packages/core/enrichment/strict_contract.go`, `packages/core/enrichment/instructions.go`, `packages/core/enrichment/strict_contract_test.go`, and this worklog.
- No files under `packages/sap/**` or `apps/sap-agent/**` changed. No schema, migration, Atlas checksum, SQL query, SQLC-generated repository code, credentials, provider HTTP client, worker, post-commit behavior, product mutation, review API, UI, alias/rule persistence, inventory, pricing, barcode, or UoM conversion behavior changed.
- Passed with Go 1.26.7: `gofmt` on all human-authored Stage 2B Go files; `cd packages/core && go test ./enrichment`; `go test ./repository`; `go test ./usecase`; `go test ./...`; `cd apps/cloud-server && go test ./...`; and `git diff --check`.
- SQLC remains repository-pinned at v1.30.0. No SQL changed, so `sqlc generate` and Atlas hash/validation were intentionally not run and `atlas.sum` was not modified.
- No provider is selected and no remote model/API call exists.

## AI MVP Boundary

- All architect contracts remain in force: `product_type` is never AI; only `standard`, `raw_material`, `fixed_asset`, and `finished_good` are valid; structured SAP data, populated category, and resolved brand win; operational SAP fields and document identity are never inferred or mutated; suggestions remain reviewable outside `products.metadata` and direct product mutation; new taxonomy is never silently created.
- Stage 2B adds only fail-closed request/response validation. It does not change the Stage 1 lifecycle: `pending -> processing -> in_review -> approved -> applied`, with rejection from `in_review`, retryable/failed from processing, and retryable back to processing.

## Remaining Phases

- Stage 2C concrete provider selection/adapter plus provider execution worker.
- Stage 2D review/approval APIs.
- Stage 2E approved suggestion application.
- Deterministic approved aliases/rules.
- Reviewer UI if required.
- Production observability, security, and hardening.

## Open Questions

- Provider/model/region/retention policy remains OPEN; no provider was selected in Stage 2B.
- Reviewer roles/permissions and auto-apply policy remain unapproved.
- Candidate/taxonomy re-enrichment fingerprint/version strategy remains OPEN.
- The source/rule for `finished_good` remains OPEN.
- Attribute/family schema remains unapproved.
- Large-taxonomy scaling and candidate retrieval strategy remain OPEN.

## Current Stage 2B Verdict

- Stage 1: SAFE TO KEEP.
- Stage 2A: SAFE TO KEEP.
- Stage 2B: SAFE TO KEEP. The provider-neutral request, strict parser, candidate security, structured precedence, description policy, unsupported-semantic protection, and model instruction contract are implemented and verified without provider execution.
- Next Action: make the provider/model/region/retention selection decision before beginning Stage 2C concrete adapter implementation.

## Stage 2C — OpenAI Provider + Worker

### Provider and boundary

- The initial provider is the OpenAI Responses API using the official `github.com/openai/openai-go/v3` SDK pinned at `v3.50.0`. OpenAI is the initial provider, not an architectural dependency. Future provider replacement should be isolated to a new `ProductEnrichmentProvider` adapter and configuration/wiring.
- `ProductEnrichmentProvider` remains provider-neutral. OpenAI SDK request/response types are isolated in `packages/core/enrichment/openaiadapter/provider.go`; generic enrichment request, eligibility, parser, lifecycle, and review contracts do not import the OpenAI SDK.
- The configured model is `OPENAI_ENRICHMENT_MODEL`, defaulting to the recommended `gpt-5.6-terra`; arbitrary future model strings remain supported. `ENRICHMENT_PROVIDER=openai` is validated only when enrichment is enabled.
- The adapter uses Responses API Structured Outputs with an explicit strict JSON Schema. It sends the Stage 2B instruction contract plus an allowlisted request payload. UoM identity may be context; conversion factors and operational data are excluded from the provider payload.
- OpenAI output is extracted with `Response.OutputText`, then passed through `ParseEnrichmentResponseString` and the Stage 2B request/candidate/precedence validation. Provider metadata is attached by the adapter and never accepted from generated JSON.

### Durable worker and lifecycle

- `packages/core/enrichment/worker.go` runs a bounded sequential worker. It lists due `pending`/`retryable` rows, atomically claims each as `processing`, reconstructs the current SAP snapshot and candidate dictionaries, checks the source fingerprint and current eligibility, invokes the provider with a finite timeout, and persists validated proposals atomically with `processing -> in_review`.
- Transient provider/network/timeout/rate-limit/5xx errors become `retryable` with deterministic capped exponential backoff. Authentication, invalid-request/model, malformed/contract/prohibited/candidate/correlation failures become terminal `failed`. Graceful context cancellation does not create a retry loop. One row failure does not stop the batch.
- Source changes are fail-closed as `stale_source` (or `stale_source_no_gap`); no stale result is applied and no product is mutated. The worker never writes products, brands, categories, metadata, inventory, pricing, barcodes, or UoM conversions.
- Durable execution metadata was added in forward migration `20260820010000.sql`: `attempt_count INTEGER NOT NULL DEFAULT 0`, `next_attempt_at TIMESTAMP`, and bounded `last_error_code VARCHAR(100)`. No raw provider response or secret is stored.

### Configuration and security

- Stage 2C configuration is disabled by default. When enabled, API key, provider, model, finite timeout, worker interval, batch size, and max attempts are validated from environment-backed core config. The key is never logged, persisted, sent to SAP, or included in suggestion JSON.
- The worker logs only operational identifiers/classifications and does not log full prompts, raw model output, authorization headers, or API keys.

### Files changed and verification

- Business/domain: `packages/core/enrichment/enqueue.go`, `packages/core/enrichment/execution.go`, `packages/core/enrichment/worker.go`, `packages/core/enrichment/worker_test.go`.
- OpenAI adapter: `packages/core/enrichment/openaiadapter/provider.go`.
- Worker/config/wiring: `packages/core/config/config.go`, `apps/cloud-server/main.go`.
- SQL/schema/migration: `packages/core/db/schema/35_product_enrichment.sql`, `packages/core/db/migrations/20260820010000.sql`, `packages/core/queries/product_enrichment_execution.sql`, `packages/core/repository/product_enrichment_execution_store.go`.
- Generated SQLC: `packages/core/repository/models.go`, `packages/core/repository/product_enrichment_suggestions.sql.go`, and `packages/core/repository/product_enrichment_execution.sql.go` were generated with real SQLC v1.30.0. A second generation was deterministic. `packages/core/db/migrations/atlas.sum` was regenerated with Atlas and validated.
- Module dependency: `packages/core/go.mod` and `packages/core/go.sum` add `github.com/openai/openai-go/v3 v3.50.0`.
- No files under `packages/sap/**` or `apps/sap-agent/**` changed. No review APIs or approved-suggestion application logic was added.
- Go 1.26.7 formatting/tests, SQLC v1.30.0 generation, and Atlas hash/validate passed. No live OpenAI call was made.

### Remaining phases and open questions

- Stage 2D review/approval API.
- Stage 2E approved suggestion application.
- Deterministic approved aliases/rules.
- Reviewer UI if required.
- Production observability/security/hardening.
- Candidate/taxonomy re-enrichment strategy.
- Future attributes/families only with an approved schema.
- Reviewer roles/permissions, auto-apply policy, `finished_good` source rule, candidate/taxonomy re-enrichment policy, normalized attributes/family schema, and production OpenAI region/retention requirements remain open.

### Current Stage 2C Verdict

- Stage 1, Stage 2A, and Stage 2B remain SAFE TO KEEP by static scope review.
- Stage 2C is SAFE TO KEEP. Go/gofmt, focused adapter/worker tests, core/server tests, deterministic SQLC v1.30.0 generation, and Atlas hash/validation passed. No critical or important source finding is known from this implementation pass.
- Next Action: separately authorize Stage 2D review/approval API design/implementation. Keep Stage 2E application, aliases/rules, reviewer UI, and production hardening out of Stage 2C.

## Stage 2D — Review / Approval API Design

This section records a read-only Stage 2D design trace. No Stage 2D implementation, schema change, SQL change, provider call, UI work, commit, or push was performed. Stage 1, Stage 2A, Stage 2B, and Stage 2C remain SAFE TO KEEP.

### 1. Existing HTTP architecture and evidence

- `apps/cloud-server/main.go:setupRouter` creates the Gin engine, applies `middleware.LoggerMiddleware`, protects the normal `/api` group with `middleware.JWTAuthMiddleware()` followed by `middleware.TenantMiddleware(tenantManager)`, constructs handlers/usecases, and registers routes. The direct `/api/v1/migration` group is intentionally separate and is not a review route.
- `packages/core/middleware/auth.go:JWTAuthMiddleware` validates the Bearer JWT, stores JWT claims in `middleware.ClaimsKey`, and stores a standard user's string `user_id` claim under `middleware.UserIDKey`. It returns HTTP 401 directly for missing/invalid authentication. M2M tokens carry `client_id`, `tenant_id`, and scopes but do not carry a Nembus `users.id` reviewer identity.
- `packages/core/middleware/tenant.go:TenantMiddleware` requires `x-tenant-id`, treats it as a tenant slug, asks `middleware/manager.Manager:GetPool` to resolve the tenant database, and places `*repository.Queries` in the request context under `middleware.RepoKey`. It does not resolve or authorize an organization ID.
- Existing handlers such as `packages/core/handler/brand.go:CreateBrand`, `packages/core/handler/product_catalog.go:ListProductsWithVariants`, and `packages/core/handler/permission.go:ListPermissions` retrieve the tenant repository from context, call `SetRepository` on a request-scoped usecase, parse Gin path/query/body input, and return `repository.Response` through `c.JSON`.
- Existing routes use resource groups and action subpaths, for example `packages/core/routing/product_catalog.go:RegisterProductCatalogRoutes`, `packages/core/routing/permission.go:RegisterPermissionRoutes`, and `packages/core/routing/sap_migration.go:RegisterSAPMigrationRoutes`. Pagination is normally bounded `limit`/`offset`, not cursor-based. Existing list queries use `LIMIT` and `OFFSET`.
- Request DTOs are ordinary handler/usecase structs in `packages/core/handler/dto.go` and individual handler files. Validation is primarily Gin `binding` plus explicit `strconv` parsing and usecase validation. There is no shared request-validation package for this feature.
- The standard response is `repository.Response` from `packages/core/repository/response.model.go`, constructed by `utils.NewResponse` in `packages/core/utils/response.go`. Current utility codes are only 200, 201, 400, 404, and 500. There is no existing shared 403 or 409 code.
- `packages/core/repository/db.go:DBTX`, `New`, and `Queries.WithTx` provide the repository transaction boundary. Stage 2A/2C repository adapters keep domain interfaces separate from generated SQLC types. The Stage 2D usecase should follow that pattern.
- No current permission middleware or route guard was found. `packages/core/usecase/permission_usecase.go:CheckUserHasPermission` and repository query `CheckUserHasPermission` exist as callable permission checks, but normal handlers do not invoke them automatically. Existing `/api` routes therefore provide authentication and tenant database selection, not authorization.

### 2. Existing users, roles, permissions, and authorization

Current identity/RBAC schema is in `packages/core/db/schema/10_identity_rbac.sql`:

- `users.id` is the local reviewer-capable identity; each user has `organization_id`, `is_active`, and credentials.
- `roles`, `user_roles`, `permissions`, and `role_permissions` represent role membership and permission grants. `role_permissions.scope` exists, with values seeded as `all`, `store`, or `own` in the sample data, but the runtime `CheckUserHasPermission` query checks only whether a code is present and does not enforce scope.
- `module_permissions`, `menu_permissions`, and `submenu_permissions` support UI/navigation visibility. They do not themselves authorize an HTTP action.
- `packages/core/queries/permissions.sql:CheckUserHasPermission`, `GetUserPermissions`, and `GetUserPermissionsWithScope`, together with `packages/core/usecase/permission_usecase.go`, are the existing permission APIs. `packages/core/routing/permission.go` exposes permission administration/check endpoints, but those endpoints are also only behind JWT and tenant middleware.
- No runtime special treatment for `admin`, `owner`, `manager`, `is_system_role`, or any named role was found in the API authorization path. `roles.is_system_role` is a data attribute used by role CRUD SQL, not an automatic approval privilege.
- The sample seed `apps/cloud-server/scripts/init-Data-Dump.sql` contains broad `products:view`, `products:manage`, and `products:delete` permissions, and `settings:audit_logs` for viewing audit logs. `products:manage` is assigned to broad catalog/operator roles such as `store_manager` and `owner`; it is not a narrowly scoped enrichment-review permission. `settings:audit_logs` is not an approval permission.

Answers to the permission questions:

- A. There is no existing permission whose semantics specifically cover product-enrichment review. `products:manage` is the nearest product permission but is materially broader than approving AI suggestions, while `products:view` cannot authorize approval.
- B. Reusing `products:manage` would couple review access to broad product/catalog mutation access and would not preserve least privilege. It is not recommended as the Stage 2D approval gate. `products:view` may be used only if the product separately decides to make queue reads visible to all product viewers; it must not authorize approve/reject.
- C. The repository supports a narrowly scoped permission cleanly because `permissions.code` is unique and `user_roles -> role_permissions -> permissions` is already the grant path. No schema redesign is required.
- D. Adding a permission requires a permission seed/data change and role-to-permission mapping. It does not require a schema migration if deployed through the existing permission tables. A module/menu/submenu mapping is optional for API-only rollout and would be needed only if a later reviewer UI uses the existing navigation model. Frontend changes are not required for API enforcement, but a UI must eventually hide/show the review surface based on the same permission.
- E. `product_enrichment_suggestions.reviewer_id` must store the authenticated local `users.id` (`int32`), after verifying that user is active and belongs to the current tenant/organization. It must not store JWT `sub`, username, M2M `client_id`, or an arbitrary request-body reviewer ID. M2M review is not supported unless a future trusted mapping to a local user is explicitly designed.

### 3. Recommended reviewer authorization

Recommendation, not architect-approved fact: add the permission code `product_enrichment:review`, following the repository's `resource:action` convention while keeping the resource distinct from broad product mutation. Require it for list, detail, approve, and reject. The usecase should call the existing permission query with the authenticated user ID, but also explicitly validate active-user and current-organization membership because the existing permission query does not do either.

The API must not infer approval rights from a role name. A deployment may map the new permission to a deliberately selected role through `role_permissions`; the architect/product owner must choose that role. The current repository does not prove that `admin`, `owner`, `store_manager`, or `manager` should approve.

There is an additional current authentication limitation: user JWTs generated by `middleware.GenerateJWTToken` contain `user_id` and `user_login`, but not tenant or organization identity. `TenantMiddleware` trusts the request `x-tenant-id` header to select a tenant database. A numeric user ID can theoretically collide across tenant databases. Stage 2D implementation must either establish a trusted tenant binding for user JWTs (for example, include and compare a tenant claim) or obtain an equivalent deployment guarantee before production approval access is enabled. Merely trusting a caller-supplied organization query parameter is not acceptable.

### 4. Organization and tenant scoping

- `product_enrichment_suggestions.organization_id` references `organizations(id)` and `product_id` references `products(id)` with cascade deletes (`packages/core/db/schema/35_product_enrichment.sql`). The Stage 1 query and every Stage 2C mutating query carry `organization_id` predicates.
- Normal `/api` requests select a tenant database from `x-tenant-id`; the current organization is not carried in the JWT and is not a middleware context value. `users.organization_id` is the available authenticated membership relationship inside that tenant database.
- Stage 2D routes must be registered only under the authenticated `/api` group. They must not accept `organization_id` as a query/body override. The handler extracts `middleware.UserIDKey`, parses it as a positive `int32`, loads the user from the tenant repository, requires an active user, and uses that user's `organization_id` as the effective organization for every queue/detail/status/audit query.
- Every suggestion read and transition must include `organization_id = effective_user.organization_id`; every current-product reload must include both organization ID and product ID (or the organization-scoped source SKU plus a product-ID correlation check). A guessed ID from another organization must produce the same not-found behavior as any other missing suggestion.
- The existing `UserUseCase.getOrganizationID` chooses the first active organization and is suitable only for its current user-management behavior; it must not be copied for Stage 2D because review authorization must use the authenticated user's own `users.organization_id`.
- M2M requests do not provide a local reviewer identity. They should receive unauthorized/forbidden review access unless a future explicit service-account-to-user audit mapping is approved.

### 5. List review queue API

Recommended routes, following the existing `/api` resource-group style:

- `GET /api/product-enrichment/suggestions`
- Default `status=in_review`; accept a bounded status filter only from the persisted lifecycle values if history is intentionally exposed. The reviewer queue itself must always support `status=in_review`.
- Parse `limit` and `offset` using the existing style, with a conservative default of `limit=50`, `offset=0`, a hard maximum of 100, and rejection of negative/non-numeric values for this new security-sensitive endpoint. No cursor protocol is present in the repository and none is needed for the first queue.
- Do not add provider filtering or free-text SKU/product search in the first API. The current Stage 1 list query has no such filters and the existing queue index is for organization/status/created time.

The response should remain the standard `repository.Response` envelope with a purpose-built review DTO array in `data`, not generated `repository.ProductEnrichmentSuggestion` rows and not raw JSONB. Each list item should contain:

- `id`, `product_id`, `source_item_code`, and `source_item_name`;
- a sanitized `inference_snapshot` containing only product type, brand/category identity, and description as they existed at inference;
- a `current_product_state` containing product type, current category, current brand, and current description;
- `proposals.brand`, `proposals.category`, `proposals.description`, and `proposals.unsupported_semantics`, preserving action, canonical target/name, value, confidence, evidence, and explanation;
- `provider`, `model`, `model_version`, `created_at`, `updated_at`, `status`, `reviewer_id` when present, and `reviewed_at` when present;
- `stale`, `approval_blocked`, and a bounded `conflict_fields`/reason representation when current state can be compared.

The DTO must not return `structured_current` wholesale because the Stage 2A/2C snapshot can contain UoM/conversion context that is not required for review. It must not return `products.metadata`, API keys, provider request/response payloads, prompts, raw OpenAI response IDs, inventory, prices, tax, supplier, barcode, warehouse, or operational flags.

The existing `ListProductEnrichmentSuggestions` query in `packages/core/queries/product_enrichment_suggestions.sql` already provides organization/status/limit/offset ordering and the index `idx_product_enrichment_suggestions_organization_status` supports the queue pattern. It does not join current product/category/brand state. Stage 2D should add a purpose-built organization-scoped review-list query joining `products` and active/current taxonomy names, or a bounded equivalent, and generate its SQLC companion. It should not alter the Stage 1 worker query or add speculative indexes.

### 6. Get single suggestion API and DTO boundary

Recommended route: `GET /api/product-enrichment/suggestions/:id`.

The detail response uses the same DTO as the list, with the full safe proposal evidence and a clearly separated shape:

```text
current_product_state: authoritative database state
inference_snapshot:   sanitized state captured before model inference
proposals:             provider-neutral AI suggestions
review:                status, reviewer_id, reviewed_at, timestamps
review_safety:         stale, approval_blocked, conflict_fields, blocking_reason
```

The detail handler/usecase must reload the current product state rather than treating `structured_current` as current truth. The existing `ProductEnrichmentStore.LoadSAPProductEnrichmentSnapshot` is a useful repository boundary because it reloads by organization and source item code and returns structured brand/category/description/product type with a product identity check. A dedicated Stage 2D query that loads by `(organization_id, product_id)` and joins current taxonomy is preferable for a review detail response; it must not reuse unscoped `GetProduct` or unscoped `GetProductWithDetails` for a user-controlled ID.

The response should include `inference_snapshot`, `current_product_state`, `stale`, and `conflict_fields`. The exact full fingerprint comparison is mandatory in approve/reject safety code; list/detail may also expose field-level conflict information for usability. No current Stage 1 column needs to be changed for these indicators because the persisted `source_data_fingerprint` and `structured_current` already provide the inference baseline.

### 7. Stale-source safety policy

Approval must be fail-closed if the current authoritative source differs from the inference fingerprint. The approve usecase should:

1. Load the suggestion under the effective organization.
2. Reload the current SAP product snapshot and verify organization/product identity.
3. recompute the existing `enrichment.FingerprintSnapshot` and compare it to `suggestion.source_data_fingerprint`.
4. Revalidate the stored proposal JSON through the existing `enrichment.ProposalSet.Validate` contract and current structured precedence.
5. Refuse approval if the fingerprint differs or a current authoritative field has become populated/resolved.

This covers a brand resolved after inference, a description populated after inference, and a category populated after inference. It also safely fails closed for any other source-context change included by the established fingerprint. The response should identify `stale_source` and the safe conflict fields where known, but the current lifecycle has no `stale` status. Leave the row `in_review`; do not silently reject it, add a new status, rerun enrichment, or call OpenAI. A later explicit re-enqueue/rerun workflow can create a new fingerprinted suggestion.

Because Stage 1 has one suggestion-level status, Stage 2D should not approve only the still-safe fields after a fingerprint conflict. The conservative MVP policy is whole-row approval blocked until a fresh enrichment result exists.

### 8. Whole-suggestion versus field-level review

Stage 1 persists brand, category, description, and unsupported semantics in separate JSONB columns, but it has one `status`, one `reviewer_id`, one `reviewed_at`, and one `applied_at` for the whole suggestion. `enrichment.ProposalSet` has no persisted field-decision or reviewer-override structure. Therefore current persistence cannot represent “approve brand, reject category, approve description.”

Recommendation for the first MVP: whole-suggestion review only. A reviewer may approve only when every applyable proposal in the row is acceptable; if one applyable proposal is bad, reject the whole row. `KEEP_EXISTING`, `NO_MATCH`, and unsupported semantics do not create applyable product changes. Description `PROPOSE_NEW` is an applyable description proposal; brand/category `PROPOSE_NEW` are blocked as described below. This makes approval safe without changing Stage 1 schema, at the cost of rejecting otherwise useful mixed-quality rows.

If product requirements demand mixed decisions, the smallest future correction is field-level decision persistence (for example, per-field decision/action plus reviewer/time and canonical target identity) and a corresponding application contract. That is a design gap for field-level review, not a reason to add an unsafe implicit convention to Stage 2D. No schema change is made in this design pass.

### 9. Approve API semantics

Recommended route: `POST /api/product-enrichment/suggestions/:id/approve` with no reviewer ID and no proposal-edit body.

The usecase must require authenticated local user identity, the recommended review permission, active membership in the effective organization, `status = in_review`, a non-stale source fingerprint, valid persisted proposals, and no blocked brand/category `PROPOSE_NEW` action. It then performs only:

- `in_review -> approved`;
- set `reviewer_id` to the authenticated `users.id`;
- set `reviewed_at = CURRENT_TIMESTAMP`;
- keep `applied_at = NULL`;
- do not mutate products, brands, categories, UoMs, variants, aliases, metadata, or taxonomy.

The status update must retain the existing organization and `status = 'in_review'` SQL predicates. The transition and audit insert should run in one database transaction using `repository.Queries.WithTx`; an audit failure should roll back the approval so an accepted review is never silently unaudited. There is no OpenAI call and no Stage 2E application here.

Approval is not idempotent as a repeated state overwrite: a second request after approval/rejection/applied is a conflict, not a silent success. Concurrent reviewers race on the single guarded SQL update; exactly one receives the returned row. The losing request receives a state conflict and never overwrites `reviewer_id`.

### 10. Reject API semantics

Recommended route: `POST /api/product-enrichment/suggestions/:id/reject`, with no reviewer ID and no editable proposal body.

The same organization, identity, permission, active-user, and `status = in_review` checks apply. The atomic transition is `in_review -> rejected`, with authenticated `reviewer_id`, `reviewed_at = CURRENT_TIMESTAMP`, and `applied_at = NULL`. It must not mutate any product or taxonomy record. Concurrent approve/reject requests are resolved by the same guarded update; only the winner succeeds.

The current `product_enrichment_suggestions` schema has no review reason, rejection reason, reviewer note, or comment field. The repository audit table can preserve the fact of the transition but is not a substitute for a user-facing note. Notes are optional for the conservative MVP; if required, the smallest correction is a nullable bounded `reviewer_note`/`review_reason` field with a migration, DTO validation, SQLC regeneration, and audit coverage. Do not accept and discard a reason in Stage 2D.

### 11. Manual editing of AI proposals

Stage 2D should be approve/reject-only. The current JSONB proposal model has no reviewer-edited value, edit provenance, before/after field audit, or field-level decision. Allowing an API caller to rewrite proposal JSON would also create an unvalidated path around Stage 2B's provider contract and canonical-target checks. A reviewer who disagrees with a proposal should reject the whole suggestion in the MVP. Reviewer editing is a later architect decision requiring explicit provenance and validation design.

### 12. `PROPOSE_NEW` brand/category handling

`enrichment.BrandProposal` and `CategoryProposal` deliberately require no target ID/code for `PROPOSE_NEW`; the action is a review proposal and does not authorize creation. Stage 2D must never create a brand or category.

- `MATCH_EXISTING` with a validated canonical existing target may participate in whole-suggestion approval.
- Brand/category `PROPOSE_NEW` is visible with its canonical name/evidence but blocks approval in the MVP because current persistence has no reviewer-selected canonical replacement and Stage 2E cannot safely map the concept to an ID.
- The reviewer may reject the suggestion. A separate future taxonomy-resolution workflow may resolve the proposal to a canonical ID; that workflow must persist the mapping before an approved suggestion can be applied.
- Description `PROPOSE_NEW` is different: it is the current contract's action for proposing a description value and may be approved as part of an otherwise safe whole suggestion.

The smallest future correction is either a taxonomy-resolution table/field linking `(suggestion_id, field)` to an existing canonical brand/category ID, or a field-level review decision structure that stores that mapping. Automatic taxonomy creation remains out of scope.

### 13. Unsupported semantics

Unsupported semantics such as shampoo, anti-dandruff, 400 ml, model number, dimensions, and family hints are displayed as informational evidence only. They remain inside the proposal DTO for reviewer context, do not create an approval target, do not participate in product mutation, and do not imply that future attribute/family schema work has been approved. Approval means only that the whole current suggestion is accepted for the later Stage 2E contract; unsupported semantics must be ignored by Stage 2E until a separately approved destination exists.

### 14. Audit behavior

`audit_logs` exists in `packages/core/db/schema/80_promotions_loyalty.sql` and the generated `repository.AuditLog` model exists, with `organization_id`, `table_name`, `record_id`, `action`, old/new JSON, changed fields, `performed_by`, IP, user agent, session ID, and timestamp. No reusable audit query/usecase/service or current review-action usage was found. The seed has `settings:audit_logs` for viewing, but no review audit writer or review permission.

Stage 2D approve/reject must add a minimal audit repository query and write one entry in the same transaction as the guarded status update:

- `organization_id`: effective authenticated user's organization;
- `table_name`: `product_enrichment_suggestions`;
- `record_id`: suggestion ID as text;
- `action`: `UPDATE` because the existing check constraint allows only `INSERT`, `UPDATE`, `DELETE`, and `SELECT` (not dotted event names);
- `old_values`: `{ "status": "in_review" }`;
- `new_values`: `{ "status": "approved"|"rejected", "event": "product_enrichment.approved"|"product_enrichment.rejected" }`;
- `changed_fields`: `status`, `reviewer_id`, and `reviewed_at`;
- `performed_by`: authenticated `users.id`; request IP/user-agent/session may be passed through when safely available.

Do not store API keys, prompts, raw OpenAI request/response data, full provider payloads, inventory/prices/tax/supplier/barcode data, or unnecessary product metadata in the audit event.

### 15. Error and concurrency model

The current middleware/usecase conventions support the following Stage 2D mapping; 403/409 are recommendations because `utils/response.go` currently lacks constants for them:

- 401: missing/invalid JWT, emitted by `JWTAuthMiddleware`; M2M without a local reviewer identity is also not an authenticated human reviewer.
- 403: authenticated user lacks the recommended review permission or is not an active authorized member of the effective organization. This must not be inferred from a role name.
- 400: malformed ID, invalid pagination/status, unexpected request body, or unsupported review action.
- 404: suggestion not found under the effective organization. Cross-organization IDs must resolve to the same generic not-found behavior and must not disclose that a row exists elsewhere.
- 409: current state is not `in_review`, the source is stale, taxonomy resolution is required for a brand/category `PROPOSE_NEW`, or a guarded concurrent transition returned no row after the suggestion was known to be reviewable. A second concurrent review must never overwrite the first reviewer.
- 500: repository, transaction, audit, or current-product reload failure. Do not expose SQL details or provider secrets.

For a guarded update that returns no row, a scoped read may distinguish a missing suggestion (404) from a known row whose status changed (409); the update predicate remains the concurrency authority. The usecase must not implement check-then-update without the status predicate.

### 16. Pagination/query requirements

The existing Stage 1 `ListProductEnrichmentSuggestions` query is adequate for organization/status/offset ordering and has an appropriate `(organization_id, status, created_at DESC)` index. It is not sufficient as the final review DTO query because it lacks current product joins and would expose raw generated rows. Stage 2D should add only:

- an organization-scoped review queue query with bounded `status`, `LIMIT`, `OFFSET`, ordered by `created_at DESC, id DESC`;
- an organization-scoped review detail/current-state query or an adapter method that reloads current product state safely by organization/product;
- an audit insert query;
- generated SQLC files regenerated by the repository's pinned workflow.

No provider index, search index, cross-organization list, count endpoint, cursor scheme, or speculative filter is justified by current evidence.

### 17. Future UI contract (without UI implementation)

The DTO should allow a later UI to render two explicit columns/sections:

```text
CURRENT (authoritative)          PROPOSED (AI, reviewable)
Brand: None                      Brand: MATCH_EXISTING -> PANTENE
Category: Personal Care          Category: KEEP_EXISTING
Description: Empty               Description: <bounded text>
Product Type: standard           Product Type: LOCKED / absent from proposals

Evidence: proposal evidence/explanation
Unsupported semantics: informational list only
Review safety: stale/conflict/approval-blocked indicators
```

The API must never make an AI proposal look like current product truth and must never include `product_type` in the proposal target. Provider/model metadata is context only; raw provider material is excluded.

### 18. Stage 2E boundary

Stage 2D ends at exactly:

```text
in_review -> approved
in_review -> rejected
```

Stage 2E is a separate phase: `approved -> revalidate current authoritative state -> apply only allowed/canonical proposals -> applied`. Stage 2E must repeat organization/product identity, fingerprint/stale, structured precedence, canonical-target, action, and no-authoritative-field checks. It must handle taxonomy resolution and unsupported semantics explicitly. Stage 2D must never create `applied`, mutate products, create taxonomy, create aliases, or trigger another model run.

### 19. Exact future Stage 2D patch surface

The smallest conceptual implementation surface is:

- `packages/core/enrichment/review.go` or the existing enrichment package: provider-neutral stale/approval eligibility and sanitized proposal review helpers; no provider dependency.
- `packages/core/usecase/product_enrichment_review_usecase.go`: request-scoped repository boundary, authenticated-user/organization/permission checks, list/detail DTO mapping, stale reload, atomic approve/reject orchestration, and whole-suggestion/`PROPOSE_NEW` policy.
- `packages/core/handler/product_enrichment_review.go`: Gin handlers, DTOs, path/query parsing, identity extraction, and standard response mapping.
- `packages/core/routing/product_enrichment.go`: route registration for the list/detail/approve/reject endpoints.
- `packages/core/queries/product_enrichment_suggestions.sql`: only the new review projection/detail queries if the existing rows cannot supply the safe DTO; the existing worker lifecycle queries must remain unchanged.
- a new `packages/core/queries/audit_logs.sql` (or equivalent query file) and generated repository companions for the audit insert and review projections; do not hand-edit generated SQLC output.
- `packages/core/repository/product_enrichment_review_store.go` and/or an audit adapter: narrow interfaces carrying organization IDs on every read/write, using `Queries.WithTx` for status plus audit.
- `packages/core/utils/response.go`: add explicit 403/409 constants only if the implementation keeps the repository response envelope for those outcomes; otherwise map them in one documented handler boundary. Do not silently reuse 400 for concurrency after choosing 409 semantics.
- `packages/core/middleware/auth.go`: only if the architect accepts the required tenant-binding hardening for user JWTs; current JWTs do not carry tenant/organization identity.
- `apps/cloud-server/main.go:setupRouter` and `main`: construct/wire the review usecase/handler and register routes under the authenticated `/api` group. Do not register them under unauthenticated `/api/v1/migration`.
- `apps/cloud-server/scripts/init-Data-Dump.sql`: seed `product_enrichment:review` and explicit role mappings if that seed remains the deployment source. A UI module/submenu mapping is optional and belongs to a later UI decision.
- focused handler/usecase/repository tests and server route/auth tests. No SAP, OpenAI adapter, worker, configuration, product mutation, schema, or migration files are required for the conservative whole-suggestion MVP.

### 20. Future Stage 2D test plan

- Authorization: missing/invalid JWT rejected; M2M without local reviewer rejected; missing `product_enrichment:review` rejected; active authorized user allowed; role name alone does not authorize; tenant binding and user organization are checked.
- Organization isolation: list only returns the authenticated organization; detail/approve/reject with a guessed foreign ID do not disclose it; all repository SQL includes organization predicates; user-supplied organization query/body values are ignored/rejected.
- List: default `in_review`, explicit valid status filtering, bounded limit/offset, negative/oversized/malformed pagination, deterministic ordering, no raw provider payload/secrets/operational data.
- Detail: sanitized inference snapshot/current product separation, provider/model metadata, proposal confidence/evidence/explanation, reviewer/audit metadata, no UoM conversion/inventory/price/tax/supplier/barcode leakage, and stale/conflict indicators.
- Approve: `in_review -> approved`, authenticated reviewer ID persisted, reviewed time persisted, `applied_at` remains NULL, no product/taxonomy mutation, no provider call, and audit row committed atomically.
- Reject: `in_review -> rejected`, reviewer/time persisted, `applied_at` remains NULL, no product mutation, and audit row committed atomically.
- Invalid/concurrent transitions: wrong status returns the documented conflict; simultaneous approve/reject has one winner; the losing request cannot overwrite reviewer identity; audit failure rolls back the status transition.
- Stale source: brand/category/description changes after inference block approval; the row remains `in_review`; no implicit rejection or rerun occurs.
- `PROPOSE_NEW`: visible, no brand/category creation, no approval while canonical target is absent, and no Stage 2E apply path is implied.
- Whole-suggestion policy: mixed-quality actionable proposals require rejection; no hidden field-level partial approval exists.
- Unsupported semantics: informational only and never applied.
- No live OpenAI call is required for any Stage 2D test.

### 21. Architect decisions still required

Answerable from repository evidence:

- the repository has database-backed roles/permissions and an existing permission check query, but no enforced permission middleware;
- `users.id` is the correct reviewer foreign-key identity;
- tenant selection is the `x-tenant-id` header and organization membership is `users.organization_id` inside the tenant database;
- the Stage 1 lifecycle and fingerprint are sufficient to implement conservative whole-row review without changing Stage 1 schema;
- audit table exists, but a reusable audit writer does not;
- no current field-level decision, reviewer note, taxonomy-resolution, or manual-edit persistence exists.

Requires product/architect decision:

- approve the exact permission code `product_enrichment:review` and which roles receive it; do not infer this from role names;
- require tenant binding in user JWTs before production review access, or explicitly approve an equivalent trusted deployment boundary;
- approve whole-suggestion MVP review, where mixed-quality actionable proposals are rejected as a row, or authorize a field-level schema correction before Stage 2D implementation;
- decide whether reviewer notes/rejection reasons are needed in MVP;
- decide the separate canonical-resolution workflow for brand/category `PROPOSE_NEW`;
- decide whether reviewer editing is ever allowed; recommendation is no for MVP;
- decide whether queue history may list statuses other than `in_review`; recommendation is an `in_review`-first queue with optional read-only history later;
- confirm that unsupported semantics are informational only; this is the current safe recommendation;
- confirm no auto-apply and no taxonomy creation in Stage 2D.

### 22. Recommended MVP review policy

- Use an explicit narrowly scoped review permission and organization-scoped access.
- Permit list/detail/approve/reject only for authenticated active users with that permission.
- Default the queue to `in_review`; use bounded limit/offset and no speculative search.
- Use approve/reject only; no reviewer proposal editing.
- Approve/reject only the whole suggestion; reject mixed-quality actionable rows.
- Allow canonical `MATCH_EXISTING` brand/category and description proposals when the source fingerprint remains current.
- Display brand/category `PROPOSE_NEW`, but block approval until a future canonical taxonomy-resolution workflow supplies an existing ID/code.
- Treat unsupported semantics as evidence-only information.
- Block approval on any authoritative source/fingerprint conflict and require fresh enrichment; keep the row `in_review` because no stale status exists.
- Record the authenticated `users.id` and review timestamp; keep `applied_at` NULL.
- Audit approve/reject atomically with the status change using safe status-only old/new values.
- Do not apply approved suggestions, create brands/categories, create aliases, mutate products, call OpenAI, or rerun enrichment.

## Current Phase

Stage 2D design

## Remaining Phases

- Stage 2D implementation.
- Stage 2E approved suggestion application.
- Deterministic aliases/rules.
- Reviewer UI if required.
- Production observability, security, and hardening.

## Stage 1/2 Impact

NO FOUNDATION CORRECTION REQUIRED

The current Stage 1 persistence is sufficient for a conservative whole-suggestion Stage 2D review API. Field-level approval, reviewer notes, taxonomy resolution, and tenant-bound user JWTs remain explicit future/implementation decisions; none is silently added here.

## Next Action

Obtain architect approval for the recommended `product_enrichment:review` permission, tenant-binding requirement, whole-suggestion MVP, stale-source blocking, and `PROPOSE_NEW` taxonomy policy. Then separately authorize Stage 2D implementation. Do not begin Stage 2E application or UI work as part of that implementation authorization.

## Stage 2D Implementation Attempt — BLOCKED

- No Stage 2D review routes, handlers, usecase, repository SQL, permission seed, generated SQLC, or main wiring were added.
- The requested implementation was stopped by the mandatory tenant-security gate before shipping an unsafe endpoint.
- Actual request path: `JWTAuthMiddleware` validates a user JWT containing `user_id`/`user_login` but no tenant or organization binding; `TenantMiddleware` accepts the arbitrary client `x-tenant-id` header and obtains any active tenant pool through `Manager.GetPool`; the selected tenant repository then loads the user and permissions. `Manager.GetPool` performs no user-membership verification.
- Consequently, a valid user token can be replayed with another active tenant header and may resolve to a same-ID/authorized user in that tenant. The review API cannot safely derive organization identity from the authenticated user until tenant selection is bound to that user by an existing trusted mechanism.
- This is a genuine cross-tenant authorization risk under the Stage 2D contract. The feature remains `NEEDS CORRECTION`; no unsafe review endpoint was shipped.
- Required smallest correction before implementation: architect-approved tenant/user binding for user JWT requests (or an equivalent existing membership check in the tenant-selection path), followed by a fresh Stage 2D implementation review.
- Permission provisioning, reviewer notes, whole-suggestion approval, stale-source checks, PROPOSE_NEW blocking, atomic audit behavior, and DTO policy remain design-only and were not implemented in this attempt.

## Tenant Authentication Security Gate

Review date: 2026-08-20. This is a read-only security design review; no source, schema, migration, generated code, configuration, dependency, database, Git history, or external system was changed. Only this worklog was updated.

### Current tenancy model

- The master database is opened from `MASTER_DB_URL` in `apps/cloud-server/main.go:55-68` by `setupDatabase`; its SQLC repository is passed to `manager.NewManager` at `main.go:271-277`.
- The master `tenants` registry is defined in `packages/core/db/schema/10_identity_rbac.sql:19-28` and contains UUID `id`, unique `slug`, `db_conn_str`, active state, and settings. It contains no user, account, membership, or organization-access relation.
- `x-tenant-id` is not an organization ID on normal `/api` routes. `packages/core/middleware/tenant.go:19-47` passes its raw value to `manager.Manager.GetPool`; `packages/core/middleware/manager/manager.go:22-50` uses it as `tenants.slug`, looks up the active master registry row with `GetTenantBySlug`, opens that row's `db_conn_str`, caches the pool by slug, and injects `repository.New(pool)` into request context.
- The application therefore uses physically separate PostgreSQL databases for the master database and active tenant databases. `apps/cloud-server/cmd/migrate-tenants/main.go:97-153,213-240` applies migrations to the master and then to each active tenant connection string.
- Within a tenant database, the schema is also organization-row scoped. `organizations` is a normal table (`10_identity_rbac.sql:5-17`), and `users.organization_id` is a foreign key (`20_stores_terminals.sql:39-52`). The schema permits more than one organization in one tenant database; it does not prove that production deployments use only one.
- Products, users, roles, permissions, and other business rows are local to the selected database and additionally may carry `organization_id`. Their numeric primary keys are not globally unique across physically separate tenant databases. `users.id` is `SERIAL PRIMARY KEY` and `username`/`email` are individually unique only within each database (`20_stores_terminals.sql:39-52`).
- The same email or human can consequently exist in different tenant databases, but no repository membership/account model establishes whether that is an intended product policy. The repository does not prove that one human can switch among tenants without a separate login.
- No `tenant_users`, `user_tenants`, membership, invitation, global-user, account, or session table was found. The master tenant registry is an infrastructure registry, not an authoritative user-membership registry.
- The `/api/v1/migration` path is a separate inconsistency: `apps/cloud-server/main.go:230-236` registers it without JWT or tenant middleware, and `packages/core/handler/sap_migration.go:45-58` parses `x-tenant-id` as an integer organization ID (default `1`) rather than as a tenant slug. This machine-ingestion boundary requires its own hardening review; it is not evidence of authenticated user membership.

### Current authentication and tenant-selection flow

#### Login

1. `POST /api/auth/login` is registered in `apps/cloud-server/main.go:99-104` under an auth group using only `middleware.TenantMiddleware(tenantManager)`.
2. The client must send `x-tenant-id`. `TenantMiddleware` selects the tenant pool from the master registry before credentials are checked.
3. `packages/core/handler/auth.go:49-78` obtains the selected repository from `middleware.RepoKey`, binds `user_login` and `password`, and calls `AuthUseCase.Login`.
4. `packages/core/usecase/auth_usecase.go:30-67` queries `GetUserByUsername` in the already-selected tenant database, checks active state, compares the bcrypt password, and calls `middleware.GenerateJWTToken` with only the local numeric user ID converted to a string and the login name.
5. `packages/core/middleware/auth.go:275-302` issues a 24-hour HS256 JWT containing `user_id`, `user_login`, `exp`, and `iat`. It does not contain tenant or organization identity. User tokens also do not set `iss` or `aud`.

#### Authenticated `/api` request

The exact group order in `apps/cloud-server/main.go:114-117` is:

```text
HTTP request
-> global CORS middleware
-> LoggerMiddleware
-> JWTAuthMiddleware
-> TenantMiddleware
-> route handler
-> usecase/repository using the selected tenant DB
```

- `JWTAuthMiddleware` (`packages/core/middleware/auth.go:126-253`) parses and verifies the bearer token with `JWT_SECRET`, accepts a valid JWT without requiring a tenant claim, and stores the standard token's `user_id` and claims in Gin context. It does not load the user and does not verify membership.
- For M2M tokens, the same middleware validates `is_m2m`, client ID, active config entry, and optional exact whitelisted token, then overwrites `x-tenant-id` with the registry client's `TenantID` (`auth.go:184-240`). M2M claims already include `tenant_id`, but the current code does not explicitly compare that claim with the matched registry entry.
- After JWT parsing, `TenantMiddleware` trusts the client header for standard user requests and selects the database. There is no equality check between a trusted token tenant and the requested header because standard user tokens have no tenant claim.
- There is no universal authenticated-user lookup in middleware. Individual handlers/usecases use the selected repository; for example, `GET /api/users/:id` is registered without a current-user check (`packages/core/routing/user.go:10-38`) and `UserUseCase.GetUser` calls `GetUser(id)` in whichever database `TenantMiddleware` selected (`packages/core/usecase/user_usecase.go:221-239`).
- `apps/pos-client/main.go:109-127` duplicates the same login order and authenticated route order, so the shared correction must cover both cloud and POS HTTP routers.

#### Other routes and callers

- `/health`, Swagger, and the cloud `/api/dev/token` route are outside the authenticated tenant group. `packages/core/handler/dev.go:24-58` currently has its environment restriction commented out and emits an unbound development JWT. The POS client only exposes its dev route in development, but `apps/pos-client/app.go:645-652` also creates an unbound token.
- `/api/tenants` uses `MasterRepositoryMiddleware` and is a master-registry path, not a tenant-business path (`apps/cloud-server/main.go:106-112`).
- There is no refresh-token endpoint, refresh-token table, session store, or refresh-token parser in the repository. `GenerateJWTToken` is the only standard user issuance path; M2M issuance is separate.
- SAP ingestion is intentionally not an HTTP-user flow in the current implementation: `SAPMigrationUseCase` is constructed with `masterPool` (`apps/cloud-server/main.go:339-343`), `/api/v1/migration` has no JWT middleware, and `IngestBatch` accepts an organization ID from payload/header. The SAP agent sends an API key as a bearer header but this route does not validate it.
- Stage 2A/2C enrichment is also an internal flow. `apps/cloud-server/main.go:341-355` constructs `ProductEnrichmentStore` with `masterRepo` and starts `EnrichmentWorker` without HTTP JWT. `packages/core/enrichment/execution.go:83-96` explicitly describes the worker as scanning a master queue while every mutation carries `organization_id` for SQL scoping. JWT tenant binding must not be imposed on this worker; its trusted internal boundary and organization-scoped SQL remain separate.
- ZATCA reporting is a background worker using `masterRepo` (`apps/cloud-server/internal/zatca/service.go:111-173`). gRPC sync/backup services resolve a tenant slug through `Manager.GetPool`/`GetTenantDSN` without HTTP JWT; their trust and authentication are separate internal protocol concerns. In particular, `BackupServer.validateToken` currently only rejects an empty string (`apps/cloud-server/internal/grpc/backup_server.go:65-70,201-210`) and is not a substitute for HTTP tenant binding.

### Cross-tenant security verdict

`CROSS_TENANT_SWITCH PROVEN POSSIBLE`

Concrete static path:

1. A user logs into Tenant A by sending `x-tenant-id: tenant-a` to `/api/auth/login`; TenantMiddleware selects A before `GetUserByUsername` and bcrypt validation.
2. The successful token contains the A user's local `user_id` and login, but no tenant or organization claim.
3. The same bearer token is sent to an authenticated endpoint with `x-tenant-id: tenant-b`.
4. `JWTAuthMiddleware` accepts the signature and expiry and has no tenant check. It does not query A, B, or a central membership source.
5. `TenantMiddleware` asks the master registry for slug `tenant-b`, opens B's `db_conn_str`, and places B's repository in context.
6. The handler executes against B. If B has the same local numeric user ID, path-based user access such as `GET /api/users/1` resolves B's user; broad tenant-scoped list/read/write handlers likewise operate on B. Role and permission queries, when used, are performed in B and therefore do not prove that the token's A principal belongs to B.

No existing check prevents the switch. Equal numeric IDs or equal emails are not a defense: IDs are database-local, and email uniqueness is database-local. A token subject, issuer, or audience check cannot prevent the switch because standard user tokens have no tenant claim and `user_id` is not globally unique.

### Tenant-bound JWT option

This is the smallest repository-compatible correction because login already authenticates inside a selected tenant. Bind each standard user session to the exact tenant selector used by the current architecture: the active `tenants.slug`. The minimum user claim should be `tenant_slug` containing that slug; it must not contain an organization ID, database DSN, secret, or credential. If the project standardizes on the existing M2M field name `tenant_id`, its value must still be explicitly documented and validated as the tenant slug, not `organizations.id`.

- Login must pass the selected slug from `AuthHandler`/request context into `AuthUseCase.Login` and `GenerateJWTToken`; the value is already required for `TenantMiddleware` but is not currently passed to the usecase.
- A trusted binding middleware must require a well-formed tenant claim and compare it exactly with the requested `x-tenant-id` before any tenant pool is opened. Missing, malformed, or mismatched values fail closed.
- This works for the current tenant-specific login model: one token/session is for one tenant. Switching tenants requires a fresh login and a new tenant-bound token.
- It does not support silent multi-tenant switching. That requires a new trusted membership/account authority or an explicitly issued multi-tenant session containing a server-proven allowed-tenant set; neither exists here.
- The canonical UUID `tenants.id` could be used in a larger design, but the current selector and pool manager use `slug`. Claiming the slug is the smallest safe change; changing a slug should invalidate old bindings naturally and require login again.

### Server-side membership option

Option B is not implementable from current repository evidence. The master `tenants` table can prove that a slug maps to an active database, but it cannot prove that an authenticated user may access it. Tenant-local `users.organization_id`, local user IDs, role names, and repeated emails are not cross-database membership authority. A safe Option B would require new central identity/membership storage and a defined global principal, not simply adding `organization_id` to the current JWT.

### Recommended fix and middleware order

Use a tenant-bound user JWT and enforce it before tenant DB selection for every authenticated HTTP business route:

1. `JWTAuthMiddleware` parses/verifies the bearer token and requires a valid standard-user identity or a separately validated M2M identity.
2. `TenantBindingMiddleware` reads the trusted `tenant_slug` claim. For standard users it requires a non-empty `x-tenant-id` and exact equality. For M2M, it requires the registry tenant and signed claim to agree, then allows the existing registry-driven header override.
3. `TenantMiddleware` resolves the already-bound slug through the master registry and opens the selected pool.
4. The handler/usecase loads the user from that selected database and derives `users.organization_id` server-side.
5. Authorization/permission checks and all suggestion/product queries run in the bound repository with organization predicates. No request organization ID is trusted.

For compatibility, retaining the header while requiring equality is preferable to silently accepting a changed header. Deriving the tenant only from the trusted claim is also safe, but would be a broader client/API behavior change. Do not trust a user-supplied organization ID, header alone, role name, same numeric user ID, or same email.

### Token and refresh design

- Access token: retain the current user identity and 24-hour expiry, add `tenant_slug`, and reject malformed/missing claims. Adding a fixed issuer/audience policy may be done in the same hardening, but it is not the tenant-binding authority; current standard user tokens have no issuer/audience and current M2M tokens use `iss: nembus-api`.
- Old standard JWTs without `tenant_slug` must fail closed on protected routes after deployment. The safe operational default is re-login; do not infer a tenant from the old `user_id`, email, header, or organization row.
- There is no current refresh flow to patch or migrate. If one is added, the refresh credential must carry or server-side-resolve the original tenant context, and refresh must always mint the same tenant-bound access token. It must not accept a requested Tenant B and exchange a Tenant A refresh credential for B.
- Existing M2M tokens are a distinct protocol. Preserve their registry binding, but make the future validation explicitly compare the signed tenant claim with the matched `M2MClient.TenantID` and reject disagreement. Do not treat M2M configuration as human user membership.

### Exact future patch surface

No implementation is authorized by this gate. The smallest future source patch is:

- Auth token issuance: `packages/core/usecase/auth_usecase.go:30-70`, `packages/core/handler/auth.go:49-78`, and `packages/core/middleware/auth.go:275-308` to pass and sign the selected tenant slug; `packages/core/handler/dev.go` and `apps/pos-client/app.go:645-652` must not issue an unbound protected token.
- Auth validation: `packages/core/middleware/auth.go:126-253` to require standard identity, validate tenant claim shape, and make M2M claim/registry agreement explicit; add the binding check in `auth.go` or `tenant.go` without modifying generated code.
- Tenant middleware/router order: `packages/core/middleware/tenant.go:19-47` only if the binding helper is placed there; otherwise keep pool lookup unchanged and add the binding middleware between the existing JWT and tenant middleware in `apps/cloud-server/main.go:114-117` and `apps/pos-client/main.go:124-127`.
- Refresh flow: no current file exists. Any future refresh handler/service, token store, or endpoint must be added only with the same tenant-bound context; no current refresh file should be invented for this gate.
- Tests: extend `packages/core/middleware/auth_test.go`; add focused auth-usecase/handler or router tests where the project test harness permits. Include tenant A/B, missing/malformed claims, M2M agreement, same local ID, same email, permission lookup, old-token rejection, and route ordering.
- Configuration: not required for the minimal claim/equality correction. A token version or issuer/audience rollout setting is optional, not a reason to weaken fail-closed behavior.

### Security test plan

- Login in A emits a signed token bound to A's slug; token contains no DSN, secret, password, or operational database data.
- A token plus A header succeeds; the same token plus B header is rejected before `Manager.GetPool` connects to B. B token plus B header succeeds.
- Missing, non-string, empty, or malformed tenant claim is rejected; an unknown tenant slug is rejected by tenant resolution; old tokens without the claim are rejected according to the migration policy.
- A user ID `5` in A and an unrelated user ID `5` in B cannot cross access with an A token. Equal emails in A/B cannot cross access either.
- A's role/permission cannot be evaluated against B using an A token; after binding, the selected repository and server-derived `users.organization_id` are the only authorization context.
- M2M registry tenant, signed claim, and selected header must agree; rotated/inactive clients remain rejected.
- If refresh is introduced: A refresh yields only an A token, cannot request B, and preserves tenant binding.
- Future Stage 2D list/detail/approve/reject tests use organization-scoped IDs and prove a foreign suggestion ID is inaccessible without disclosure.
- SAP migration, enrichment worker, ZATCA worker, and trusted internal gRPC paths continue without HTTP JWT, while their existing explicit tenant/organization boundaries remain separately tested.

### Open policy questions versus necessary correction

Technically necessary now: fail closed for standard protected requests unless the authenticated token is bound to the selected tenant; enforce the check before tenant pool selection; reject old unbound tokens or provide a separately proven migration authority; secure all authenticated business routes, not only enrichment review.

Open product/architecture policy: whether one human must switch among multiple tenants without re-login; whether a central global identity and membership model should be introduced; whether tenant slugs may change operationally; and whether the unauthenticated `/api/v1/migration`/gRPC boundaries need a separate machine-auth redesign. None of these questions justifies trusting the current header alone.

### Stage 2D

Implementation remains BLOCKED until tenant security is corrected and the above tests pass. After correction, the intended authorization chain remains valid in principle:

```text
authenticated token
-> tenant binding verified
-> bound tenant repository selected
-> selected-database user loaded
-> users.organization_id derived server-side
-> explicit review permission checked
-> organization-scoped suggestion query
```

Stage 2D cannot proceed unchanged operationally until its repository boundary is reconciled with current Stage 2A/2C wiring: `apps/cloud-server/main.go:341-355` currently stores and scans enrichment suggestions through `masterRepo/masterPool`, while authenticated `/api` routes select a tenant repository. The tenant-binding authorization design must be retained, and the review API must explicitly use the authoritative suggestion storage or move/bridge it through an approved scoped store. No request organization ID may select either store.

Next Action: Read-only Stage 2D repository-boundary analysis: determine where `product_enrichment_suggestions` is stored versus tenant-selected HTTP repositories and define the safe access boundary before review API implementation. Preserve Stage 1, Stage 2A, Stage 2B, and Stage 2C SAFE TO KEEP history; do not begin Stage 2D implementation, Stage 2E application, or UI work as part of this gate.

## Tenant-Bound User Authentication

Review date: 2026-08-20. The standard interactive-user cross-tenant token replay vulnerability was corrected without adding central identity/membership storage, tenant schema changes, migration files, SQLC changes, Stage 2D routes, SAP changes, or enrichment behavior.

- Before this patch, a valid user JWT could be replayed with another active `x-tenant-id` header because the token contained no tenant binding and `TenantMiddleware` trusted the header before selecting a tenant pool. The vulnerability was `CROSS_TENANT_SWITCH PROVEN POSSIBLE`.
- Standard user JWTs now contain the exact `tenant_slug` claim. The claim is sourced only from the canonical tenant context placed by `TenantMiddleware` after the login request's tenant slug has been verified through the master tenant registry and the tenant pool has been selected. The login credentials are then checked in that selected tenant database before the token is signed.
- Protected cloud and POS route order is `JWTAuthMiddleware -> TenantBindingMiddleware -> TenantMiddleware -> handler`. `TenantBindingMiddleware` requires a valid signed `tenant_slug`, requires a valid `x-tenant-id`, and requires exact equality before `TenantMiddleware` can call `Manager.GetPool`. `TenantMiddleware` uses the trusted context slug and rejects authenticated requests that arrive without the binding middleware; it cannot use the header to override the signed context.
- Login remains `TenantMiddleware -> login handler -> tenant-local credential validation -> tenant-bound token issuance`; login does not require a JWT first. Existing `x-tenant-id` client compatibility is preserved as a consistency assertion on authenticated requests.
- Old standard user tokens without `tenant_slug`, malformed/empty claims, mismatched headers, and malformed headers fail closed with a generic unauthorized response. No tenant is inferred from user ID, login, email, role, organization ID, request body, query, or another tenant database. Users holding old unbound tokens must log in again after deployment.
- POS authenticated requests continue to send `x-tenant-id`; the POS local development token path now requires configured `DEV_TENANT_SLUG`. Cloud/POS development token scripts require explicit `DEV_TENANT_SLUG`, and the development HTTP token endpoint requires an explicit `x-tenant-id`; no arbitrary production tenant is silently defaulted.
- M2M tokens remain a separate mechanism. Their existing registry-selected tenant behavior is preserved, while the signed `tenant_id` is checked against the active M2M registry entry and an explicitly supplied header is checked before tenant selection. SAP migration, product enrichment workers, ZATCA/background workers, and internal gRPC jobs do not use this HTTP user-JWT binding.
- No refresh-token flow exists. Any future refresh/session mechanism must preserve the signed tenant binding and must not accept a client-requested alternate tenant.
- Focused middleware tests cover same-tenant success, cross-tenant rejection, old-token rejection, malformed/empty claims and headers, trusted context extraction, preserved user claims, empty-tenant token issuance rejection, and rejection before a downstream `TenantMiddleware`/pool-selection path. `go test ./middleware`, `go test ./usecase`, and `go test ./...` from `packages/core` passed; `go test ./...` from `apps/cloud-server` passed; affected POS subpackages passed. The POS root package remains untestable in this checkout because its pre-existing `docs/swagger` package and embedded `frontend/dist` are absent.
- Files changed for this correction: `packages/core/middleware/auth.go`, `packages/core/middleware/tenant.go`, `packages/core/middleware/auth_test.go`, `packages/core/usecase/auth_usecase.go`, `packages/core/handler/auth.go`, `packages/core/handler/dev.go`, `apps/cloud-server/main.go`, `apps/cloud-server/scripts/generate-dev-token.go`, `apps/pos-client/main.go`, `apps/pos-client/app.go`, `apps/pos-client/internal/config/config.go`, and `apps/pos-client/scripts/scripts/generate-dev-token.go`.
- Deployment impact: all ordinary user tokens issued before this correction lack `tenant_slug` and are intentionally invalid on tenant-scoped authenticated routes; users must log in again. M2M/system tokens are not converted into standard user tokens.

## Stage 2D

Status: BLOCKED. Do not implement review routes yet.

Next action: Read-only Stage 2D repository-boundary analysis: determine where `product_enrichment_suggestions` is stored versus tenant-selected HTTP repositories and define the safe access boundary before review API implementation.

Open policy questions remain:

- multi-tenant session switching without re-login;
- possible future global identity/membership;
- tenant slug mutation policy;
- separate `/api/v1/migration`/gRPC authentication hardening, if still applicable;
- reviewer roles/permission provisioning;
- reviewer notes;
- field-level review; and
- `PROPOSE_NEW` canonical resolution.

## Stage 2D Repository Boundary

Review date: 2026-08-20. This is the authoritative read-only repository-boundary gate for Stage 2D. No source code, schema, migration, SQL, generated code, configuration, dependency, database, Git history, or external system was changed for this analysis. Only this worklog was updated.

### Database topology and repository ownership

- The master/control PostgreSQL connection is `MASTER_DB_URL`. `apps/cloud-server/main.go:setupDatabase` creates `masterPool` and `masterRepo = repository.New(masterPool)`. `masterRepo` is the control-plane repository used by the tenant manager and master endpoints; the `tenants` registry is read from this database by `packages/core/middleware/manager/manager.go:GetPool`/`GetTenantDSN`.
- Each tenant has a separate PostgreSQL connection in `tenants.db_conn_str`. `Manager.GetPool` caches one `*pgxpool.Pool` per tenant slug, and `packages/core/middleware/tenant.go:TenantMiddleware` wraps that pool in `repository.New(pool)` and stores the `*repository.Queries` under `middleware.RepoKey` in the request context. The repository type is the same as the master repository; the underlying DBTX handle is different.
- `packages/core/repository/db.go` defines the common `DBTX` interface, `repository.New`, and `Queries.WithTx`. It does not identify a database. Database identity comes from the pool/transaction passed to it.
- The normal authenticated HTTP order is `JWTAuthMiddleware -> TenantBindingMiddleware -> TenantMiddleware -> handler`, wired in `apps/cloud-server/main.go` and duplicated in `apps/pos-client/main.go`. The trusted tenant slug is checked before `Manager.GetPool`; the resulting HTTP repository is tenant-local.
- `MasterRepositoryMiddleware` injects `masterRepo` only for master-registry endpoints such as `/api/tenants`. It is not the repository used by normal tenant business routes.
- SAP migration is a separate global construction: `apps/cloud-server/main.go:341` calls `usecase.NewSAPMigrationUseCase(masterPool)` once at startup. The `/api/v1/migration` route is outside JWT and tenant middleware; `packages/core/handler/sap_migration.go:51-59` parses `x-tenant-id` as an integer organization ID and defaults to `1`. The `/api/migration` route uses the same globally constructed usecase, so the request context repository does not control SAP writes.
- Stage 2A coordinator construction is `repository.NewProductEnrichmentStore(masterRepo)` at `apps/cloud-server/main.go:342`, followed by `enrichment.NewProductEnrichmentCoordinator(enrichmentStore)`. It is therefore master-bound, not request-bound and not transaction-bound to SAP.
- Stage 2C construction is also global and master-bound: `enrichment.NewEnrichmentWorker(enrichmentStore, provider, ...)` at `apps/cloud-server/main.go:350-354`. There is one worker per cloud-server process, one store, and one underlying master repository. It does not iterate tenant slugs or create tenant repositories.
- The OpenAI provider is only called by the worker. `packages/core/enrichment/worker.go` receives no tenant slug or tenant pool; the durable row carries only local `organization_id`, `product_id`, and suggestion ID.
- `packages/core/usecase/sap_migration.go` owns one `*pgxpool.Pool`, begins its transaction from that pool, performs staging and deterministic product writes, commits, and only then calls the coordinator. The coordinator reloads the product and inserts the suggestion through the separately constructed master repository.
- Existing tenant handlers commonly read `middleware.RepoKey` and call `SetRepository` on a usecase, for example `packages/core/handler/auth.go:26-34,49-75` and `packages/core/handler/product_catalog.go:25-86`. These usecases are globally constructed but receive a request-selected repository dynamically. No current Stage 2D handler exists.

### Schema and table ownership matrix

The repository's intended migration path applies the same `packages/core/db/migrations` directory to both the master database and every active tenant database. This is explicit in `apps/cloud-server/cmd/migrate-tenants/main.go:104-153`, `apps/cloud-server/MIGRATIONS.md`, and the root `Makefile` `db-migrate` target. Therefore the physical schema copies are as follows; “both” means separate copies in separate PostgreSQL databases, not shared tables.

| Table | Physical schema target | Current operational data path | Evidence / consequence |
|---|---|---|---|
| `organizations` | Both master and tenant DBs when the core migrations are applied | Master SAP path writes/reads master; HTTP identity/catalog uses tenant copy | `10_identity_rbac.sql:5-17`; baseline and Atlas migration runner apply it to both |
| `users` | Both | SAP user ingestion uses master; authenticated HTTP users are loaded from the selected tenant DB | `20_stores_terminals.sql:39-52`; `AuthUseCase.Login` reads the request-selected repository |
| `roles` | Both | Tenant-local RBAC for HTTP; master copy exists by schema, but is not a cross-tenant identity authority | `10_identity_rbac.sql:165-175` |
| `permissions` | Both | Tenant-local permission queries for HTTP | `10_identity_rbac.sql:132-139`; `CheckUserHasPermission` is local to its `Queries` handle |
| `products` | Both | Current SAP migration and enrichment use master; normal catalog APIs use tenant DB | `30_catalog.sql:85-107`; `SAPMigrationUseCase` has `masterPool` |
| `product_categories` / `brands` | Both | Current SAP and enrichment candidate lookup use master; HTTP catalog uses tenant DB | `30_catalog.sql:5-34`; `ProductEnrichmentStore` uses its bound `Queries` |
| `product_variants` | Both | Separate local catalog copies | `30_catalog.sql:109-118` |
| `product_enrichment_suggestions` | Both only after `20260820000000.sql` and `20260820010000.sql` are applied; current Stage 1/2 records are written to master | Master queue is the actual Stage 2A/2C store; tenant copies may exist but are not populated by the current enrichment path | Migration FKs and indexes in `20260820000000.sql`; master wiring in `apps/cloud-server/main.go:341-354` |
| `audit_logs` | Both | No Stage 2D audit write exists yet; future tenant-local review should write tenant audit rows | `80_promotions_loyalty.sql:111-126` |
| `staging.sap_migration_batches` | Both after core migration | Current SAP migration writes the master copy because its transaction is from `masterPool` | `95_sap_staging.sql:5-18`; `sap_migration.go:50-63` |

The live database migration version of any particular tenant was not queried in this read-only pass. The code and deployment path prove the intended targets, not that every deployed tenant has already applied every migration.

There are legacy/duplicated baseline files under `apps/cloud-server/migrations` and `apps/pos-client/migrations`, but the current multi-database Atlas runner resolves `packages/core/db/migrations` and applies it to master and active tenant DSNs. `apps/cloud-server/main.go` does not run migrations automatically at startup.

### Product enrichment lifecycle and FK proof

For a product such as SAP `INV00006`, the current path is:

```text
/api/v1/migration request
  -> global SAPMigrationUseCase(masterPool)
  -> master PostgreSQL transaction
  -> master products / staging.sap_migration_batches
  -> commit
  -> coordinator(masterRepo)
  -> master product reload by (organization_id, source SKU)
  -> master product_enrichment_suggestions insert
  -> one global worker(masterRepo)
  -> master suggestion claim / product reload / candidate lookup
  -> master suggestion status = in_review
```

- The SAP handler supplies an integer organization ID from the request header or payload; it does not supply or authenticate a tenant slug. The `SAPMigrationUseCase` uses its startup-injected `masterPool`, not `RepoKey`.
- The product upsert is in the master transaction. The product ID is allocated in the master database, and the post-commit coordinator receives the same organization ID and source item code.
- The coordinator's `LoadSAPProductEnrichmentSnapshot` calls `GetProductBySKU`, then brand/category/UoM/conversion queries through the same `masterRepo` object. `CreateOrGetPendingSuggestion` inserts through that same master repository, after the SAP transaction has committed.
- The suggestion row's `organization_id` and `product_id` therefore refer to rows in the same physical master database in the current path. The suggestion insert is not part of the SAP transaction; enqueue failure is logged after commit and does not roll back SAP.
- The Stage 1/2C FKs are local PostgreSQL FKs: `organization_id REFERENCES organizations(id)`, `product_id REFERENCES products(id)`, and `reviewer_id REFERENCES users(id) ON DELETE SET NULL`. PostgreSQL cannot enforce these FKs across an independent tenant database. Their presence proves that each deployed copy of the suggestion table must have local `organizations`, `products`, and `users` tables; it does not make master numeric IDs globally meaningful.
- `packages/core/enrichment/execution.go` and `product_enrichment_execution.sql` pass `organization_id` on claim and every mutation, but `ListDueProductEnrichmentSuggestions` scans the one repository's entire queue without a tenant selector. The worker later reloads the product and candidates through that same master-bound store.

### SAP and worker ownership conclusions

- SAP migration is currently master-only, globally constructed once, and not reconstructed per tenant/request. `organization_id` identifies a row in the master database in this path; it is not a tenant database identity.
- Stage 2A is master-only and post-commit. The coordinator has no pool, tenant slug, or transaction parameter beyond the master-bound store supplied at construction.
- Stage 2C is one worker over one master database. It can function for rows and products that genuinely live in that master database. It cannot see tenant-only products or tenant-only suggestion rows, and it cannot recover the tenant database from a suggestion row because no tenant identity is stored.
- The current worker/reviewer consistency is therefore broken for tenant HTTP review: the worker writes `in_review` in master, while a normal authenticated request naturally receives a tenant repository. A tenant repository query cannot see the master row.

### Identifier collision analysis

`organizations.id`, `products.id`, `users.id`, and `product_enrichment_suggestions.id` are `SERIAL`/integer database-local identifiers. They are not globally unique across tenant databases. `tenants.id` is a UUID in the master registry, and `tenants.slug` is the current trusted database selector, but neither is stored in the enrichment suggestion.

For the concrete case:

```text
tenant-a: organization_id = 1, product_id = 95, suggestion_id = 1
tenant-b: organization_id = 1, product_id = 95, suggestion_id = 1
```

those values are valid and independent when the rows are tenant-local. A master/global suggestion table containing only numeric organization/product IDs cannot distinguish them. Current SAP ingestion also does not carry a trusted tenant slug into the master store, so two tenant feeds can attach to the same master numeric organization or produce an ambiguous master product lineage. The uniqueness key `(organization_id, product_id, source_data_fingerprint, contract_version)` does not repair that ambiguity.

This is a CRITICAL design defect for a global enrichment queue: the current global row has no `tenant_slug`, tenant UUID, tenant database identity, or other globally unique product identity. The defect is not present when the suggestion and referenced records are physically tenant-local and every query uses the selected tenant database plus `organization_id`.

### FK, reviewer, and audit analysis

- In the current master path, the suggestion FK targets `master.organizations`, `master.products`, and `master.users`. A tenant-local authenticated user's numeric `users.id` cannot safely be placed into master `reviewer_id`: the same number can identify another master user or no row at all, and there is no cross-database membership mapping.
- `audit_logs` has local FKs to `organizations` and `users`. Stage 1 has no audit insertion. A future master-side suggestion update plus master-side audit insert could be one transaction only for a master reviewer identity; the current tenant-local JWT user is not such an identity. A cross-database suggestion update plus tenant audit insert cannot be one PostgreSQL transaction.
- If the foundation is corrected to tenant-local storage, suggestion, product, organization, reviewer user, permission/RBAC, and audit rows are co-located. Approve/reject can use one tenant pool transaction and `Queries.WithTx`, with `reviewer_id` set only after loading the active tenant-local user and deriving that user's organization.

### Tenant identity and provisioning implications

- `product_enrichment_suggestions` currently stores no `tenant_slug`, `tenant_id`, or database identity. That is unsafe in the actual master/global deployment and is not a reason to add a tenant column if the storage is moved to tenant-local databases.
- `20260820000000.sql` creates the Stage 1 table and `20260820010000.sql` adds Stage 2C retry metadata. The current `atlas.sum` includes both files.
- `apps/cloud-server/cmd/migrate-tenants/main.go` migrates master first when `-master` is enabled, queries active tenants from the master registry, and applies the same Atlas directory to each non-empty `db_conn_str`. It is an explicit operator command, not startup provisioning.
- `TenantUseCase.CreateTenant` only inserts a row into the master `tenants` registry. It does not create a database, run Atlas, apply the enrichment migrations, seed permissions, or warm a pool. A newly registered tenant can therefore lack the enrichment tables until the operator provisions/migrates it.
- Existing tenants are not automatically migrated by application startup. The deployment must run the tenant migration command and inspect failures; inactive tenants are skipped by the current active-tenant discovery. Stage 2D must not be enabled until every intended review tenant has the Stage 1/2C tables and current Atlas state.

### Physical ownership verdict

`ENRICHMENT_DATA_MASTER_GLOBAL`

This verdict refers to the actual Stage 2A/2C lifecycle, not merely schema presence. The table is defined in the common schema and intended to exist in both database families, but `apps/cloud-server/main.go` binds the coordinator and worker to `masterRepo`, and SAP products are written through `masterPool`. The current enrichment records therefore live in the master database. The prior note calling this “current master-database enrichment storage” was correct; this audit adds the missing qualification that tenant HTTP repositories are separate and cannot safely review those rows.

### Recommended Stage 2D architecture and foundation decision

`ENRICHMENT FOUNDATION CORRECTION REQUIRED`

The smallest safe architecture is a foundation correction to tenant-local enrichment, followed by tenant-local review:

1. Establish a trusted tenant slug for the SAP machine request. The current numeric `x-tenant-id`/default `1` is not a safe database selector.
2. Resolve the tenant pool for that request and execute SAP deterministic writes, staging, post-commit snapshot reload, and suggestion enqueue against that tenant pool/repository. Do not use the global `masterPool` for tenant business data.
3. Run the existing Stage 1 and Stage 2C migrations on every intended tenant. No tenant column is needed once physical database tenancy is the isolation boundary.
4. Replace the single master-bound worker with a tenant-aware deployment model: one worker/store per tenant, or a trusted global worker that enumerates active tenants from the master registry and constructs a tenant repository for each. Each worker iteration must retain the tenant slug/pool context and never fall back to master for product or suggestion data.
5. Use the tenant-selected repository exclusively for future Stage 2D suggestion, product, user, permission, audit, and transaction operations. Every suggestion query still requires the authenticated user's derived `organization_id`; tenant binding alone is not organization authorization.

Existing master suggestion rows cannot be safely migrated by numeric IDs alone. Any backfill requires an authoritative mapping of master product/source SKU to tenant slug and tenant organization/product. Rows without that mapping must be quarantined or left inaccessible; guessing from equal numeric IDs is forbidden.

### Safe future Stage 2D request flow after correction

```text
HTTP request
 -> JWTAuthMiddleware
 -> TenantBindingMiddleware
 -> TenantMiddleware / trusted tenant pool
 -> RepoKey tenant repository
 -> authenticated tenant-local user lookup
 -> active-user + users.organization_id derivation
 -> tenant-local product_enrichment:review permission check
 -> organization-scoped suggestion query
 -> organization/product-scoped current product reload
 -> fingerprint and stale-source validation
 -> tenant-pool transaction
      suggestion approve/reject
      + tenant audit_logs insert
 -> response
```

The handler must not accept organization ID, reviewer ID, tenant slug, or database identity from the request body/query as authority. The reviewer ID is the authenticated tenant-local `users.id`. The approve/reject transaction must use the same tenant database handle as the suggestion and audit rows; `Queries.WithTx` is the existing repository primitive, but a future request-scoped review store must also have access to the selected tenant pool/transaction factory.

### Worker/reviewer consistency and multi-tenant assessment

- Current state: CRITICAL mismatch. Stage 2C writes `status = 'in_review'` in master; a tenant-bound HTTP reviewer reads a tenant copy. They are not the same physical record.
- Corrected target state: the worker and reviewer use the same tenant slug, tenant pool, local suggestion ID, local product ID, and organization predicate. Worker completion and reviewer reads then address the same physical row.
- Current Stage 2C is functionally single-master, not multi-tenant complete. It does not iterate the tenant registry. If the foundation is moved tenant-local, leaving this worker wiring unchanged would make enrichment silently process no tenant rows. This is a Stage 2C deployment/design defect that must be corrected before enabling the tenant-local review API.
- Disabled or removed tenants are rejected by `Manager.GetPool` because tenant lookup requires active status. An unavailable tenant must be skipped/retried as an operational error; neither worker nor review API may fall back to master or another tenant. A worker must not process a tenant's rows after that tenant is disabled unless an explicit operational policy says otherwise.

### Organization and permission ownership

Multiple organizations can exist inside one tenant database: `organizations` is a table with `SERIAL id`, and `users`, `products`, staging, suggestions, and audit rows carry organization references. Tenant binding is therefore insufficient. Stage 2D must load the authenticated tenant-local user, verify it is active, derive `users.organization_id`, and apply that value to every suggestion/product/audit query.

The future `product_enrichment:review` permission belongs in the tenant-local `permissions`, `roles`, `role_permissions`, and `user_roles` tables. The existing permission query is `CheckUserHasPermission` in `packages/core/queries/permissions.sql`; it is local to the selected repository and does not itself verify active user or organization membership. Permission seed/mapping must be applied to existing tenants as deployment data; creating a permission in a source seed file does not update already-running tenant databases automatically.

### Deployment prerequisites

Code prerequisites:

- Complete the tenant-local foundation correction before any Stage 2D route or review store is implemented.
- Define and authenticate the SAP tenant selector; preserve the already SAFE tenant-bound user JWT path for HTTP.
- Make SAP coordinator/store and Stage 2C worker tenant-aware; remove master fallback for product/suggestion operations.
- Decide and provision the `product_enrichment:review` permission and role mappings in each intended tenant.
- Define treatment of existing master rows and require an authoritative tenant mapping before migration/backfill.

Deployment prerequisites:

- Apply and verify `20260820000000.sql` and `20260820010000.sql` in every intended tenant database through the Atlas tenant migration runner; do not rely on startup or tenant creation to do this.
- Verify tenant migration status and repair failed/inactive-tenant provisioning before enabling review.
- Deploy tenant-bound JWT correction and require user re-login for old standard tokens without `tenant_slug`.
- Run a tenant-aware Stage 2C worker for every enabled tenant, or a verified registry-iterating worker.
- Enable OpenAI configuration only for tenants/environments approved for provider use; enrichment must remain disabled by default elsewhere.

### Exact next patch surface

Because the foundation correction is required, do not implement Stage 2D routes yet. The first source correction is limited to these existing boundaries:

- `packages/core/handler/sap_migration.go` — obtain a trusted machine tenant selector and stop treating the tenant header as an integer-only organization/database authority.
- `packages/core/usecase/sap_migration.go` — accept/use the request-selected tenant database handle, keeping the SAP transaction and post-commit enqueue in the same tenant database family.
- `apps/cloud-server/main.go` — remove the globally master-bound SAP/enrichment wiring and construct the tenant-aware coordinator/worker orchestration.
- `packages/core/repository/product_enrichment_store.go` and `packages/core/repository/product_enrichment_execution_store.go` — retain the provider-neutral interfaces but ensure stores are constructed from the selected tenant repository and expose the transaction-capable boundary required by review.
- `packages/core/enrichment/worker.go` / `packages/core/enrichment/execution.go` — carry tenant execution context through the worker orchestration if the chosen design is a registry-iterating worker.
- `apps/cloud-server/cmd/migrate-tenants/main.go` — use the existing runner operationally to apply current enrichment migrations to all intended tenants; no new schema column is justified by this audit.

No safe Stage 2D handler, route, review query, or audit adapter patch surface is authorized until those corrections and the legacy-master-row disposition are resolved.

### Security test plan

- Provision tenant A and tenant B with local `suggestion_id=1`, `product_id=95`, and `organization_id=1`, with different product/source content. Authenticate a tenant-A user and prove only tenant-A data is visible; repeat for tenant B.
- Send tenant-A JWT with tenant-B header and prove rejection before tenant-B pool selection; prove no master/global repository fallback occurs.
- Within one tenant, create organizations A and B with local users/products/suggestions and prove a user from organization A cannot list, detail, approve, reject, or audit organization B rows even when numeric suggestion/product IDs are guessed.
- Prove the worker's `in_review` update and the review API's read use the same tenant pool and physical row, not two copies with equal numeric IDs.
- Approve/reject in a failure-injected transaction and prove suggestion status and audit insert commit or roll back together in the same tenant DB.
- Prove `reviewer_id` references the authenticated tenant-local user, rejects inactive/foreign-organization users, and cannot accept a request-body reviewer ID, JWT subject, or M2M client ID.
- Disable a tenant or make its DB unavailable and prove review returns the existing tenant-not-found/unavailable behavior and worker skips/retries only that tenant without touching another or falling back to master.
- Verify all enrichment queries have the organization predicate, status transitions remain lifecycle-constrained, and no global queue can collide on `(organization_id, product_id)` across tenant databases.

### Current Stage 2D status

Status: BLOCKED. Preserve `Stage 1 SAFE TO KEEP`, `Stage 2A SAFE TO KEEP`, `Stage 2B SAFE TO KEEP`, `Stage 2C SAFE TO KEEP`, the architect `product_type` contract, the OpenAI provider decision, and the tenant-bound JWT `SAFE TO KEEP` correction.

Next Action: Correct the foundation first by moving the SAP post-commit enrichment lifecycle and Stage 2C execution to tenant-local repositories with a trusted SAP tenant boundary, migrate/verify every intended tenant, and resolve or quarantine existing master rows. Only after that gate passes: implement Stage 2D using the tenant-selected repository exclusively for suggestion/product/audit/reviewer operations.

## Multi-Tenant SAP / Enrichment Foundation Correction

Review date: 2026-08-20. This is a read-only design gate. Only this worklog was updated; no source, schema, SQL, migration, generated file, configuration, dependency, database, Git history, or external system was changed. No commit or push was performed.

### 1. CURRENT SAP REQUEST IDENTITY CONTRACT

The SAP agent currently sends `POST {Cloud.BaseURL}/api/v1/migration/batch` from `apps/sap-agent/internal/transport/client.go:113-125` with:

- `Content-Type: application/json`;
- `Content-Encoding: gzip`;
- `X-Request-ID` containing a generated UUID;
- `x-tenant-id` containing `fmt.Sprintf("%d", Cloud.OrganizationID)`;
- `Authorization: Bearer <Cloud.APIKey>` and `x-api-key: <Cloud.APIKey>` only when the editable local `Cloud.APIKey` is non-empty; and
- a payload whose `OrganizationID` is also populated from the same local `Cloud.OrganizationID` (`apps/sap-agent/internal/etl/engine.go:493-500` and the corresponding domain branches).

`CloudConfig` has only `base_url`, `api_key`, `organization_id`, and timeout fields (`apps/sap-agent/internal/config/config.go:22-26`). The embedded agent UI exposes the organization/tenant ID and API key as editable values (`apps/sap-agent/ui/index.html:147-158`, `apps/sap-agent/ui/app.js:114-120`). The default organization is `1` (`config.go:67-71`). There is no SAP-agent tenant slug field, registration/onboarding flow, agent-to-tenant database association, or SAP gRPC/M2M registration path in the repository. The agent stores the configuration locally in `agent_config.json`.

The current request therefore does not carry a proven tenant identity. It carries a client-controlled integer-like value twice, once as a header named `x-tenant-id` and once in the JSON body. The bearer value is called an API key by the agent, but the `/api/v1` route does not currently validate it.

### 2. MIGRATION AUTHENTICATION FLOW

There are two registrations of the same migration handler:

```text
/api/migration/batch
  -> JWTAuthMiddleware
  -> TenantBindingMiddleware
  -> TenantMiddleware(tenantManager)
  -> SAPMigrationHandler
  -> globally constructed SAPMigrationUseCase(masterPool)

/api/v1/migration/batch
  -> global logger/CORS only
  -> SAPMigrationHandler
  -> globally constructed SAPMigrationUseCase(masterPool)
```

Evidence: `apps/cloud-server/main.go:114-117,229-239`; route registration is `packages/core/routing/sap_migration.go:9-13`.

For `/api/v1/migration/batch`, no `JWTAuthMiddleware`, M2M validation, API-key middleware, `TenantBindingMiddleware`, `TenantMiddleware`, or master tenant-registry lookup runs. The agent's `Authorization` and `x-api-key` headers are ignored by this route. The handler reads `x-tenant-id`, attempts `strconv.Atoi`, and silently uses organization `1` if it is absent, non-numeric, or invalid (`packages/core/handler/sap_migration.go:51-59`). The usecase then preserves a positive `payload.OrganizationID`, so the JSON body can override the parsed header (`packages/core/usecase/sap_migration.go:45-50`). There is no existence, active-state, tenant-local organization, or authorization check.

For `/api/migration/batch`, JWT authentication does run. Standard user tokens must contain the signed `tenant_slug` and match the slug header. M2M tokens must contain `is_m2m`, a client identity, and `tenant_id`; the signed tenant is checked against the file-backed `config/m2m_clients.json` entry, the client must be active, and a whitelisted token must match exactly (`packages/core/middleware/auth.go:24-30,105-126,130-267`). `TenantBindingMiddleware` repeats the M2M/header consistency check and `TenantMiddleware` resolves an active tenant pool (`packages/core/middleware/tenant.go:54-136`). However, the SAP handler still parses the slug as an integer, and the SAP usecase still writes through its startup-injected `masterPool`; the selected request repository is not used by the usecase.

The current M2M system is therefore a usable tenant-binding primitive only on protected `/api` routes, not an authentication mechanism for the actual `/api/v1` SAP route. The M2M registry is a local JSON file, not a master-database agent registration table. No code maps an SAP machine credential to an organization.

### 3. X-TENANT-ID SEMANTICS

`X_TENANT_ID_SEMANTICS_CONFLICT_CONFIRMED`

Interactive protected HTTP uses `x-tenant-id` as the exact tenant registry slug (`packages/core/middleware/tenant.go:90-117`; the Postman documentation also calls it a slug). SAP transport uses the same header name for a decimal `organization_id` (`apps/sap-agent/internal/transport/client.go:122`), and the migration handler confirms that interpretation (`packages/core/handler/sap_migration.go:51-57`). On `/api/migration/batch`, the collision is operationally visible: a valid slug fails integer parsing and leads to the fallback path.

Recommendation: use option C as the primary design. Authenticate the migration as M2M, derive the tenant slug from the verified tenant-bound credential/registry entry, and keep organization identity separate. For explicit consistency/debugging use `x-tenant-id` only for the slug and add `x-organization-id` only as a non-authoritative consistency value. Do not preserve one header with two meanings, and do not use the current numeric header or default `1` as a database selector.

### 4. TRUSTED SAP TENANT SOURCE

The strongest currently available source is the existing M2M combination:

1. JWT signature validation under the server `JWT_SECRET`;
2. `is_m2m` and `client_id`/`sub` claims;
3. lookup of the client in `config/m2m_clients.json`;
4. active-client check and optional exact token-whitelist check; and
5. exact equality between signed `tenant_id` and the registry entry's `TenantID`.

This is the existing tenant-bound M2M consistency mechanism and should be reused. It is not currently applied to `/api/v1/migration/batch`. No stronger SAP-specific source exists: there is no agent registration table, no signed registration record in the master DB, no tenant association in the SAP payload, and no trusted local agent configuration. A free client-supplied slug is not sufficient.

The current M2M registry binds only a client to a tenant slug. It does not bind a client to an organization. The correction must add an organization binding to the machine credential/registry contract (and issue a token containing the same signed organization claim, or have the server derive it from the server-side registry entry). A tenant-only claim is insufficient because one tenant database can contain multiple organizations.

### 5. ORGANIZATION RESOLUTION

`organizations` is `SERIAL`/integer and has no tenant foreign key because each physical tenant database is separate (`packages/core/db/schema/10_identity_rbac.sql:5-24`). Multiple organization rows are possible in one tenant database. `users`, `products`, staging rows, and enrichment suggestions reference that local integer.

After trusted tenant selection, the server must resolve the organization from the tenant-bound M2M registration/token and validate that the organization exists and is active in the selected tenant database before beginning the SAP transaction. The request body `OrganizationID` and any organization header can be checked for consistency, but neither can be authoritative. If the credential has no organization binding, reject the request or require an explicitly authorized organization claim; never infer from numeric IDs and never default to `1`.

### 6. AGENT CONFIGURATION IMPACT

No current onboarding supplies a tenant slug or organization binding. The secure rollout needs a new server-issued/re-registered M2M credential bound to exactly one tenant slug and one tenant-local organization, while the agent keeps only the bearer credential and endpoint plus optional non-authoritative display/consistency fields. The server must not expose `MASTER_DB_URL`, tenant DSNs, or registry contents to the agent.

An editable `organization_id` can remain temporarily for UI/display and payload consistency, but the server must compare it to the trusted credential and reject mismatches. A freely editable tenant or organization field must not select a database. Existing agents require a coordinated configuration/token update because the current default and numeric header contract are not safe compatibility behavior.

### 7. RECOMMENDED SAP REQUEST CONTRACT

Preferred future contract:

```text
Authorization: Bearer <tenant-and-organization-bound M2M JWT>
x-tenant-id: <tenant slug>                 # optional consistency header
x-organization-id: <positive integer>      # optional consistency header
```

The M2M JWT must be valid, active, and bound by the server-side registry to `client_id`, `tenant_id=<tenant slug>`, and `organization_id=<local org id>`. The server derives the canonical tenant and organization from the verified credential/registry. If consistency headers or the payload organization value are present, they must exactly match the trusted values; they never select a pool or override the claims. The SAP handler should not accept a numeric `x-tenant-id` as organization authority.

Deprecated and rejected after rollout:

- unauthenticated `/api/v1/migration/batch`;
- `x-api-key` as an unvalidated credential;
- numeric `x-tenant-id` as organization authority;
- body-only `OrganizationID` authority; and
- defaulting a missing/invalid organization to `1`.

The existing M2M token already provides the tenant claim and registry comparison. The missing part is applying it to the migration route and binding/validating organization identity. Existing old tokens without an organization binding cannot be safely accepted as unrestricted migration credentials; they must be reissued or mapped to an explicit server-side organization during a coordinated rollout.

### 8. TENANT-LOCAL SAP MIGRATION FLOW

The target flow is:

```text
M2M authentication
 -> trusted tenant + organization established
 -> master registry lookup by trusted tenant slug
 -> active tenant pool/repository
 -> organization existence/active validation in that tenant DB
 -> request-scoped tenant-bound SAPMigrationUseCase
 -> transaction begins on tenant pool
 -> deterministic SAP staging/product/UoM/etc. writes
 -> commit
 -> post-commit Stage 2A coordinator built from the same tenant repository
```

`manager.NewManager(masterRepo)` and `Manager.GetPool(slug)` are the existing pool-selection primitives (`packages/core/middleware/manager/manager.go:18-52`). The manager's master access is control-plane lookup only. The existing `GetPool` cache is keyed by slug and does not re-check active state or detect a rotated DSN after a cached pool is returned; the correction must add revalidation/invalidation or a bounded refresh policy before relying on it for machine traffic and dynamic tenant changes.

No SAP extraction, mapping, batch contract, or deterministic core write semantics need to change. The selected pool replaces only the current global `masterPool` binding.

### 9. SAPMigrationUseCase LIFETIME

Choose A/C: use a tenant-aware request factory that resolves the trusted tenant, creates a request-scoped `SAPMigrationUseCase` from the selected pool, creates its tenant-local enrichment coordinator, and invokes the existing `IngestBatch` method. This is the smallest safe change because the current usecase already owns a pool (`packages/core/usecase/sap_migration.go:16-30`) and starts its transaction from that pool (`:45-54`).

Passing a pool/store per invocation is also possible, but refactoring the large usecase to become database-neutral is broader and easier to misuse. A global mutable usecase with a request-swapped pool is unsafe under concurrent migration requests. The request factory should reject unresolved/disabled/unavailable tenants before constructing a write-capable usecase and should never substitute `masterPool`.

### 10. STAGE 2A SAME-DB DESIGN

For each migration request, construct:

```text
selected tenant pool
 -> repository.New(pool)
 -> NewProductEnrichmentStore(tenantRepo)
 -> NewProductEnrichmentCoordinator(tenantStore)
 -> tenant-scoped SAPMigrationUseCase.SetProductEnrichmentCoordinator
```

The product upsert and staging transaction use `selected tenant pool`; after commit, the coordinator reloads the product, brand/category/UoM/conversion context, and inserts the suggestion through the tenant repository. The current coordinator/store interfaces already make this composition possible (`packages/core/enrichment/enqueue.go:260-334`; `packages/core/repository/product_enrichment_store.go:19-23`). It is post-commit, not the same transaction, but it must be the same physical tenant database. No global coordinator or master enrichment store may be reachable from this path.

### 11. STAGE 2C MULTI-TENANT WORKER DESIGN

Choose option A: one bounded supervisor, not a permanent goroutine per tenant.

```text
supervisor tick
 -> masterRepo.ListActiveTenants()
 -> for each active tenant, resolve/revalidate pool
 -> repository.New(tenantPool)
 -> tenant-local ProductEnrichmentStore
 -> existing EnrichmentWorker.RunOnce with a bounded batch
 -> next tenant
```

The current worker is sequential and store-driven (`packages/core/enrichment/worker.go:11-92`), so its domain logic can be reused by constructing one worker/store per tenant iteration. The current global construction at `apps/cloud-server/main.go:341-355` must be removed. `ListDueProductEnrichmentSuggestions`, product snapshot reloads, candidate brands/categories, claims, and status transitions must all use the same tenant-local repository. The worker must carry the tenant slug as execution context for logging/metrics, but must not use it as a row-level substitute for local IDs.

One tenant outage is logged and retried independently; it must not stop other tenants. Disabled tenants are skipped. There is no master suggestion fallback. Bounded per-tenant batch size and sequential tenant iteration provide predictable connection/model usage.

### 12. TENANT DYNAMIC CHANGES

The supervisor must enumerate active tenants on every polling cycle (or a short bounded registry refresh), so a tenant added after startup is discovered without restart. A disabled tenant must stop being processed even if a pool was previously cached. A rotated `db_conn_str` must invalidate/replace the cached pool before the next iteration. A temporarily unavailable database is an isolated retryable tenant error. The current manager's never-expiring slug cache is insufficient by itself for all four behaviors and must gain revalidation/invalidation or a refreshable cache as part of the foundation correction.

### 13. MASTER DB ROLE AFTER CORRECTION

The master database remains the control plane: tenant registry (`tenants.slug`, `db_conn_str`, `is_active`, settings), active-tenant enumeration, tenant lifecycle administration, and infrastructure/migration coordination. It may retain legacy/control copies required by existing unrelated functions until separately retired.

It must not be the active business store for tenant SAP products, organizations, staging rows, enrichment suggestions, candidate dictionaries, reviewer users, permissions, or review audit rows after the correction. The common schema currently creates many of these tables in both database families; physical presence is not authority. The tenant-local database selected for a request is the authority for that tenant's business rows.

### 14. EXISTING MASTER SAP PRODUCT DATA

Repository searches found no authoritative tenant slug, tenant UUID, source-tenant field, agent registration identity, or organization-to-tenant association in SAP products, SAP staging rows, products metadata, or enrichment suggestions. SAP staging and products contain local organization/SKU values only (`packages/core/db/schema/95_sap_staging.sql:5-18`; `30_catalog.sql`), while tenant identity appears only in routing/M2M/gRPC control paths.

Therefore existing master products cannot be automatically assigned to tenant databases from equal numeric organization/product IDs, SKU alone, or matching timestamps. Backfill requires an authoritative mapping of master SAP source/batch/agent identity to tenant slug and then a validated tenant-local organization/product. Rows without that evidence must remain quarantined/inert or be manually mapped. Do not move data in this foundation phase.

### 15. EXISTING MASTER SUGGESTIONS

Leave existing master `product_enrichment_suggestions` rows in place and mark them operationally quarantined/inert for tenant review. Do not surface them through tenant-local review APIs, do not merge them by numeric IDs, and do not delete them automatically. Only a later, explicitly approved backfill may migrate a row with authoritative tenant and tenant-local organization/product mapping, preserving source fingerprint and lifecycle evidence. Unmapped rows remain available only for controlled audit/manual disposition.

### 16. TENANT DB MIGRATION READINESS

`apps/cloud-server/cmd/migrate-tenants/main.go` requires `MASTER_DB_URL`, resolves the Atlas directory, connects to master, optionally migrates master (`-master`, default `true`), enumerates `tenants WHERE is_active = true`, and applies Atlas migrations to every non-empty active tenant `db_conn_str`. It logs each tenant, continues after per-tenant failures, summarizes successes/failures, and exits nonzero if any tenant fails. `-status` runs Atlas status instead of apply; `-dir` overrides the migration directory; `-baseline` defaults to `20260101000000`. It requires Atlas via `ATLAS_PATH` or PATH and does not run automatically at server startup.

The eventual prerequisite is to run the repository-compatible command from `apps/cloud-server` (for example, `go run cmd/migrate-tenants/main.go -master=false` for tenant-only application, with the required environment and Atlas binary), and verify both `20260820000000.sql` and `20260820010000.sql` in every intended tenant. The current `apps/cloud-server/migrate-tenants.ps1` advertises a `-Down` flag that the Go runner does not define; it is not the authoritative rollback procedure. Tenant creation also does not create/migrate a database automatically.

### 17. REVIEW PERMISSION PROVISIONING

No `product_enrichment:review` permission or role mapping currently exists in the repository. The permission must be provisioned in each tenant database through an approved tenant-local seed/data migration or explicit administration, using the existing `permissions`, `role_permissions`, `user_roles`, and `CheckUserHasPermission` patterns (`packages/core/db/schema/10_identity_rbac.sql:132-175`; `packages/core/queries/permissions.sql`). It must not be granted globally through the master copy and must not be assumed present merely because the schema table exists. Do not grant it in this gate.

### 18. STAGE 2D FINAL REQUEST FLOW

After foundation correction, the review path is:

```text
browser
 -> tenant-bound user JWT
 -> TenantBindingMiddleware
 -> TenantMiddleware / tenant pool
 -> tenant-local repository
 -> authenticated active tenant-local user
 -> derive users.organization_id server-side
 -> tenant-local product_enrichment:review permission
 -> organization-scoped tenant-local suggestion/product reads
 -> stale fingerprint/source validation
 -> same tenant-pool transaction
      approve/reject suggestion
      + tenant-local audit record
```

No master business repository participates. Request body/query tenant, organization, reviewer, product, or database identifiers are not authority. `reviewer_id` is the authenticated tenant-local `users.id`; organization is the authenticated user's tenant-local `organization_id`. Keep Stage 2D BLOCKED until the SAP/worker foundation, tenant migrations, legacy-row policy, and permission provisioning are complete.

### 19. SECURITY INVARIANTS

Implementation must enforce all of these:

1. One migration request resolves exactly one trusted tenant before product/staging writes.
2. The SAP organization ID exists and is valid in that selected tenant database.
3. SAP staging, product upsert, commit, and Stage 2A suggestion insertion use the same physical tenant database.
4. Stage 2C builds provider requests only from the same tenant-local product, taxonomy, UoM, and suggestion repository.
5. Stage 2D reads and writes the same tenant-local suggestion/product/user/audit database.
6. Master numeric organization/product/user/suggestion IDs are never interpreted as tenant-local identities in another database.
7. Tenant-resolution failure has no fallback to master product/enrichment persistence.
8. Disabled or unavailable tenants produce explicit reject/skip/retry outcomes and never fall through to another tenant or master.

### 20. TEST PLAN FOR FOUNDATION CORRECTION

Required tests include:

- M2M tenant A selects only tenant A; tenant B selects only tenant B; signed/header mismatch, unknown, inactive, missing, invalid, and revoked credentials fail before pool/write selection.
- Organization binding is required and validated inside the selected tenant; invalid organization, body/header mismatch, and no-contract default-to-`1` are rejected.
- Tenant A and B may both use local `organization_id=1`, `product_id=95`, and `suggestion_id=1`; SAP and Stage 2A writes remain isolated, and each worker/reviewer sees only its own physical row.
- Tenant migration never writes master products/staging/suggestions; tenant resolution failure never invokes a master fallback.
- Worker outage isolation, disabled-tenant skip, newly active tenant discovery, DSN rotation refresh, and tenant-local candidate dictionaries are covered.
- The worker's `in_review` row is the exact row later read by the tenant-bound review path; stale source rejection and organization predicates are tested.
- Review approval/rejection uses the tenant-local authenticated user and permission, and suggestion plus audit commit/rollback together in the tenant database.

### 21. BACKWARD COMPATIBILITY

There is no safe transparent compatibility mode for the currently deployed unauthenticated/numeric contract. Old agents can send a bearer token, but `/api/v1` currently ignores it, and old tokens/registry entries do not bind organization identity. Do not preserve default-to-master/default-organization behavior to avoid an agent update.

Operational rollout should issue/re-register a tenant-and-organization-bound M2M credential, deploy the server-side authenticated migration route, update the agent to send slug `x-tenant-id` (if retained as consistency) and optional `x-organization-id`, then enable tenant-local migration. A temporary legacy endpoint may be considered only if it has a separately named contract, mandatory M2M authentication, an authoritative server-side client-to-tenant-and-organization mapping, and no database guessing; the current `/api/v1` behavior must not remain open.

### 22. EXACT IMPLEMENTATION PHASES

Phase F1 — trusted SAP tenant boundary and tenant-local SAP/Stage 2A:

- Extend the M2M client/token contract and registration tooling to bind organization as well as the already bound tenant; apply M2M authentication to the machine migration route.
- Add trusted-claim/header/body consistency validation and remove integer `x-tenant-id`/default `1` authority.
- Add tenant-pool revalidation/invalidation for active-state and DSN changes.
- Make migration handler/factory resolve the tenant pool and construct a request-scoped `SAPMigrationUseCase`; construct the coordinator/store from the same tenant repository.
- Patch surface: `packages/core/middleware/auth.go`, `packages/core/handler/m2m.go` or `apps/cloud-server/cmd/m2m-gen/main.go`, `apps/cloud-server/config/m2m_clients.json.example`, `apps/sap-agent/internal/config/config.go`, `apps/sap-agent/internal/transport/client.go`, `apps/sap-agent/internal/etl/engine.go`/UI, `packages/core/handler/sap_migration.go`, `packages/core/routing/sap_migration.go`, `apps/cloud-server/main.go`, `packages/core/usecase/sap_migration.go`, and `packages/core/middleware/manager/manager.go`, plus focused tests. Do not touch SAP extraction/mapping contracts.

Phase F2 — tenant-aware Stage 2C supervisor:

- Enumerate active tenants from master control data on each bounded cycle; resolve/revalidate each pool; construct a tenant-local store/worker; isolate failures and never fall back to master.
- Patch surface: `apps/cloud-server/main.go`, `packages/core/enrichment/worker.go` or a new supervisor at that boundary, `packages/core/enrichment/execution.go`, `packages/core/repository/product_enrichment_store.go`, `packages/core/repository/product_enrichment_execution_store.go`, manager/cache support, and worker/isolation tests.

Phase F3 — tenant provisioning and legacy-data disposition:

- Run/verify Atlas migrations in every intended tenant; provision the review permission per tenant; record active/inactive/failed tenant outcomes; quarantine master suggestions; require manual authoritative mapping for any approved backfill.
- Operational surface: `apps/cloud-server/cmd/migrate-tenants/main.go`, tenant administration/runbooks, and tenant-local RBAC data. No automatic data movement.

Only after F1–F3 pass should Stage 2D review routes, tenant-local audit transaction, stale validation, and UI be implemented. Stage 2D remains BLOCKED now.

### 23. FOUNDATION SCHEMA VERDICT

`NO_FOUNDATION_SCHEMA_CHANGE_REQUIRED`

Physical tenant isolation already provides the required boundary, and `20260820000000.sql` plus `20260820010000.sql` already define the suggestion/retry schema; both are present in `packages/core/db/migrations/atlas.sum`. Moving SAP/enrichment persistence from master to the selected tenant DB does not require adding `tenant_slug` to tenant-local rows. The M2M organization binding is a credential/control-plane contract change, not a required tenant-row schema change. If the file-backed M2M registry is later replaced by a master control-plane table, that would be a separate control-plane schema decision, not a reason to add tenant identity to every tenant-local product/suggestion row.

### 24. REUSABLE STAGE 1/2 COMPONENTS

- Stage 1 suggestion schema/lifecycle: logically reusable unchanged in each tenant DB after migrations; repository generation/verification remains a separate recorded gate.
- Stage 2A eligibility, fingerprinting, structured-current snapshot, and coordinator interfaces: reusable; build the store from the selected tenant repository.
- Stage 2B/provider-neutral parser and contract validation: reusable; it has no database identity or tenant authority.
- Stage 2C OpenAI adapter: reusable; provider calls do not choose a tenant, and the worker must supply tenant-local request data.
- Stage 2C worker logic: reusable per tenant because its state transitions carry local organization/suggestion IDs; the global master-bound construction and lack of tenant supervisor are the parts to replace.

The conclusion is “reuse domain logic, correct pool/repository wiring,” not a rejection of the architect-approved enrichment contracts.

### 25. agents.md UPDATED

This section records the current SAP identity contract, authentication gap, confirmed `x-tenant-id` collision, M2M trust source, organization resolution, tenant-pool/usecase/coordinator design, tenant-aware worker, master role, master-row quarantine, migration and permission prerequisites, security invariants, compatibility plan, and implementation phases. Stage 2D remains BLOCKED.

Next Action: Phase F1 only after the tenant-and-organization M2M trust contract is explicitly approved and provisioned. No source implementation is authorized by this design gate.

### 26. FINAL VERIFICATION

The final verification for this gate is `git status --short` and `git diff --check`. The intended result is that only `agents.md` is modified; no staging, commit, or push is allowed.

## Foundation F1 â€” Tenant-Local SAP + Stage 2A

Implementation status: F1 implementation completed in the working tree. No commit or push was performed.

The old migration path mounted `/api/v1/migration/batch` without an effective machine-only boundary, used a startup-constructed `SAPMigrationUseCase(masterPool)`, interpreted numeric `x-tenant-id` as an organization ID, and defaulted missing values to organization `1`. The old Stage 2A coordinator also used `masterRepo`.

The corrected path is:

```text
Bearer machine JWT
  -> tenant_slug + organization_id claims
  -> optional x-tenant-id/x-organization-id consistency checks
  -> active tenant lookup in master registry
  -> cached selected tenant pool/repository
  -> tenant-local organization validation
  -> request-scoped SAPMigrationUseCase(selected tenant pool)
  -> commit tenant-local staging/products/domains
  -> coordinator/store from the same tenant repository
  -> tenant-local product_enrichment_suggestions
```

- New machine JWTs contain `is_m2m=true`, `token_type=machine`, `tenant_slug`, and positive `organization_id`. No credentials, DSNs, or provider secrets are claims.
- Existing `tenant_id` registry entries remain readable as legacy tenant-slug values for non-SAP M2M compatibility; the SAP migration route requires the new `tenant_slug` claim and a registry `organization_id` match.
- `x-tenant-id` is standardized as a tenant slug. It is a consistency assertion, never an organization selector. `x-organization-id` is an optional consistency assertion. A present payload `OrganizationID` must match the trusted claim; a missing value is filled from the trusted claim, never from a default.
- `SAPMigrationAuthMiddleware` rejects ordinary user JWTs, missing/malformed claims, invalid consistency values, mismatches, and unbound legacy M2M credentials before tenant pool selection.
- `TenantMiddleware` resolves only the signed tenant slug through the master registry, stores the selected pool and repository in typed request context, and no longer exposes infrastructure/DSN details in errors.
- `SAPMigrationOrganizationMiddleware` validates the trusted organization ID through `GetOrganization` on the selected tenant repository. Master organizations are not consulted.
- The migration handler constructs immutable request-local `SAPMigrationUseCase`, `ProductEnrichmentStore`, and coordinator objects from the same selected tenant pool/repository. No global mutable repository setter is used.
- The entire deterministic transaction, including staging, categories, brands, products, UoM, inventory, pricing, barcodes, suppliers, and other existing domains, remains in the selected tenant pool. Mapping and extraction contracts were not changed.
- The previous master-bound Stage 2C worker is disabled until F2. New tenant-local suggestions remain pending; no worker processes them in the intermediate state. Existing master products/staging/suggestion rows remain in place and quarantined/inert; F1 does not move, delete, or infer mappings for them.
- Agent transport now requires `m2m_token`, `tenant_slug`, and positive `organization_id`; it sends the M2M bearer token, slug `x-tenant-id`, and optional `x-organization-id`. API-key-only migration is rejected. Agent UI/config no longer invents organization `1`.

### Security Invariants

- Trusted machine tenant and organization claims are established before any SAP business write.
- The trusted organization is validated in the selected tenant database.
- SAP staging, deterministic SAP writes, and Stage 2A suggestion persistence use one physical tenant database and one request-scoped repository/pool selection.
- Numeric IDs are never used to select a tenant, and equal IDs in two tenants remain isolated.
- Unknown, disabled, unavailable, or invalid tenants/organizations do not fall back to master business persistence.
- Stage 2D remains BLOCKED.

### F1 Tests and Verification

- Added focused machine-claim and consistency tests in `packages/core/middleware/sap_migration_test.go` and M2M claim-generation coverage in `packages/core/middleware/auth_test.go`.
- Updated M2M handler/token tooling tests and SAP-agent transport tests for bearer, slug, organization, and API-key rejection semantics.
- `git diff --check` passed.
- Go, gofmt, sqlc, Atlas, and psql are unavailable in this environment, so Go compilation/tests and formatter verification remain unexecuted. No SQL source or schema was changed; SQLC and Atlas generation are not required for F1.

### Files Changed

M2M/auth and tenant boundary: `packages/core/middleware/auth.go`, `packages/core/middleware/tenant.go`, `packages/core/middleware/sap_migration.go`, `packages/core/middleware/manager/manager.go`, and their focused tests.

Migration and Stage 2A wiring: `packages/core/handler/sap_migration.go`, `packages/core/usecase/sap_migration.go`, and `apps/cloud-server/main.go`.

Credential issuance/config: `packages/core/handler/m2m.go`, `packages/core/handler/m2m_test.go`, `apps/cloud-server/cmd/m2m-gen/main.go`, `apps/cloud-server/config/m2m_clients.json.example`.

Agent transport/config/UI: `apps/sap-agent/internal/config/config.go`, `apps/sap-agent/internal/transport/client.go`, `apps/sap-agent/internal/transport/client_test.go`, `apps/sap-agent/ui/app.js`, and `apps/sap-agent/ui/index.html`.

### Current Phase

F1 implementation. Next action, if F1 is accepted, is Foundation F2 â€” tenant-aware Stage 2C worker supervision over active tenant databases. Do not implement Stage 2D review APIs or Stage 2E application yet.
## Foundation F2 — Tenant-Aware Enrichment Worker

F1 remains SAFE TO KEEP. This section records the F2 implementation.
The supervisor uses the master repository only to enumerate active tenant
slugs through the existing `ListActiveTenants` control-plane query. It does
not enumerate or process master products, organizations, or
`product_enrichment_suggestions`.

Each supervisor cycle re-enumerates active tenants. Tenant processing is
sequential and bounded: the supervisor resolves the current tenant slug
through the existing tenant pool manager, constructs `repository.New` from
that pool, and creates the execution/context/candidate store and worker from
that repository. The generic `ProductEnrichmentWorker` remains tenant-agnostic
and retains the existing claim, stale-source, OpenAI, strict Stage 2B
validation, retry/backoff, and `in_review`/retryable/failed behavior. AI never
mutates products and no review or application API is included.

One stateless, server-configured OpenAI provider/client is shared by the
tenant-local workers. Tenant database credentials, registry data, and raw
provider payloads are not sent to OpenAI. Worker logs use the tenant slug with
tenant-local numeric IDs and normalized error classes; DSNs and secrets are
not logged.

Tenant setup, database, schema, and worker/provider failures are isolated to
that tenant pass. Disabled or unavailable tenants are skipped/fail locally;
there is no master fallback, no cross-tenant repository reuse, and no dynamic
schema migration. Tenant databases must already have the enrichment
migrations; the operational prerequisite is the established tenant migration
command, for example `go run cmd/migrate-tenants/main.go -master=false` from
`apps/cloud-server` with the required environment and Atlas binary.

Legacy master suggestions remain quarantined and untouched. Any recovery or
backfill requires a separate authoritative migration/quarantine decision and
must not infer tenant identity from colliding numeric IDs.

### Tenant-Local Data Invariant

For active business data:

Product DB = Suggestion DB = Enrichment Worker DB = Future Review DB = Future Application DB = Audit DB.

The master DB remains the control-plane registry only for this worker path.

### F2 Tests and Verification

The supervisor tests cover two active tenants with colliding local
`organization_id=1`, `product_id=95`, and `suggestion_id=1`; tenant-local brand
and category candidates and lifecycle updates remain isolated. They also cover
disabled-tenant skipping, newly active tenant discovery on a later cycle,
empty active-tenants, setup/pool failure isolation, and tenant worker/schema
failure isolation. The tests use fakes and never call live OpenAI.

SQLC and Atlas changes are not required: the existing active-tenant SQLC
query is reused and no schema/query source was changed. Final Go formatting and
test results are recorded in the task handoff after execution.

### F2 Files Changed

- `packages/core/enrichment/supervisor.go`: tenant enumeration and isolated
  sequential supervisor.
- `packages/core/enrichment/worker.go`: normalized safe worker error logging.
- `packages/core/enrichment/supervisor_test.go`: tenant routing, colliding-ID,
  dynamic discovery, disabled, and failure-isolation tests.
- `apps/cloud-server/main.go`: active control-plane registry adapter, tenant
  pool/repository worker factory, shared OpenAI provider, and lifecycle wiring.
- `agents.md`: this worklog and invariant update.

### Current Phase

F2 SAFE TO KEEP. F1 tenant-local SAP routing and tenant-bound JWT security
remain intact. Stage 2D and Stage 2E are not implemented and remain blocked.

### Next Action

If F2 is SAFE TO KEEP: Foundation F3 — tenant deployment/migration
verification, tenant-local `product_enrichment:review` permission provisioning,
and explicit legacy master suggestion quarantine readiness. Do not mark Stage 2D ready before F3.

### F2 Verification Gate - 2026-08-21

- Go toolchain: `C:\Program Files\Go\bin\go.exe`, exact version `go1.26.7 windows/amd64`. `go` and `gofmt` were not on PATH; the absolute installed executables were used. No installation or upgrade was performed.
- gofmt: `C:\Program Files\Go\bin\gofmt.exe` applied formatting only to `packages/core/enrichment/supervisor_test.go`. A subsequent `gofmt -l` check over `supervisor.go`, `worker.go`, `supervisor_test.go`, and `apps/cloud-server/main.go` returned no files.
- F2 focused tests: `cd packages/core && go test ./enrichment` passed, including the colliding-ID, disabled/new-tenant, setup-failure, and worker-failure isolation tests.
- Packages/core tests: `./repository`, `./usecase`, `./middleware`, `./handler`, and `./...` all passed with the absolute Go executable.
- Cloud-server regression: `cd apps/cloud-server && go test ./...` passed; the cloud bootstrap and tenant-aware supervisor wiring compile successfully.
- SAP-agent regression: `cd apps/sap-agent && go test ./...` passed. No unrelated pre-existing failure was observed.
- Static tenant-isolation review: the master repository is used only by `masterEnrichmentTenantRegistry.ListActiveTenants` for control-plane discovery. Each active tenant resolves its own pool through `tenantManager.GetPool`, constructs `repository.New(pool)`, and builds the local enrichment store/worker from that repository. Disabled tenants are skipped; setup/worker failures continue to later tenants; active tenants are re-enumerated on each cycle; there is no master fallback or master suggestion query. The shared OpenAI provider has no mutable tenant identity. The worker only completes suggestions to `in_review` or marks them `retryable`/`failed`; it does not mutate products.
- Supervisor lifecycle review: the first pass runs immediately, later passes use the configured ticker interval, `ticker.Stop()` is deferred, cancellation exits the goroutine, empty tenant lists do not busy-loop, processing is sequential, and no active-tenant list is permanently cached. No lifecycle defect was found.
- Colliding-ID test quality: both fake tenants use `organization_id=1`, `product_id=95`, and `suggestion_id=1`. Their snapshots, candidates, provider requests, and lifecycle completions are distinct per tenant store; both rows reach `in_review`, and no master enrichment fallback is supplied or invoked. The test establishes routing isolation without live OpenAI calls.
- SQLC / Atlas: `git diff` contains no changes to `packages/core/queries/**`, generated SQLC files, `packages/core/db/schema/**`, `packages/core/db/migrations/**`, or `atlas.sum`; SQLC generation and Atlas hashing/validation were not required for F2.
- Files changed during verification: formatting-only change to `packages/core/enrichment/supervisor_test.go`, plus this `agents.md` verification record. No business logic, SQL, schema, migration, generated file, commit, or push was performed.
- Git verification: `git diff --check` passed. The final uncommitted diff contains only the formatting change and this worklog update.

## Foundation F3 - Tenant Review Deployment Readiness

F3 implementation completed on 2026-08-21. No Stage 2D review route, Stage 2E
application path, product mutation, taxonomy creation, SAP change, commit, or
push was performed.

### Tenant migration runner findings

- `apps/cloud-server/cmd/migrate-tenants/main.go` reads `MASTER_DB_URL` from
  the environment after loading the repository's `.env` candidates.
- It resolves the Atlas migration directory from the explicit `-dir` flag or
  the existing relative `/app/migrations` and repository candidates. The
  default repository directory is `packages/core/db/migrations`.
- The master connection is used only to enumerate the registry. The query is
  explicitly `WHERE is_active = true`; disabled tenants are excluded.
- Each selected tenant's `db_conn_str` is passed independently to Atlas. An
  empty DSN is counted as a failed tenant. A tenant migration failure is
  logged, processing continues for other tenants, and the command exits with
  status 1 when any tenant failed.
- `-master=false` skips applying Atlas to the master database; it does not
  skip the master registry read required to discover active tenants. There is
  no tenant-to-master business-data fallback.
- `-status` invokes Atlas status instead of apply and is the non-mutating
  operational inspection mode, subject to Atlas/database connectivity.
- The runner now returns an error when the master tenant-table probe itself
  fails or when the `tenants` table is absent, instead of silently treating
  that condition as an empty successful registry.
- Atlas migration state makes repeated application idempotent through Atlas's
  migration table. A newly registered tenant still requires this runner or an
  approved equivalent before business use.

### Enrichment migration chain

The tenant deployment directory contains the required chain:

- `20260820000000.sql`: `product_enrichment_suggestions` foundation,
  organization/product/reviewer foreign keys, source fingerprint and contract
  uniqueness, structured/proposed JSONB fields, provider/model metadata,
  lifecycle status, review/application timestamps, and organization/status
  indexes.
- `20260820010000.sql`: non-null `attempt_count` defaulting to zero,
  `next_attempt_at`, `last_error_code`, and the pending/retryable due index.
- `20260821000000.sql`: idempotent tenant-local provisioning of the narrow
  `product_enrichment:review` permission.

The canonical schema at `packages/core/db/schema/35_product_enrichment.sql`
contains the Stage 1 and Stage 2C fields, constraints, uniqueness, and indexes
represented by the two enrichment table migrations. No schema redesign was
needed. Atlas hash and validation passed after adding the F3 migration.

### Existing and new tenant deployment state

Repository evidence establishes CODE READY for the migration chain, permission
provisioning, tenant-local F1/F2 wiring, and the migration utility. Actual
tenant database migration versions, permission rows, and reviewer mappings are
DEPLOYMENT STATE UNVERIFIED because no live tenant database was queried or
altered in this task.

`CreateTenant` only inserts the tenant registry row; it does not provision a
database or run Atlas. The current API accepts the requested active flag, so
the deployment process must register a new tenant inactive (or keep it out of
business use), provision its database, run the migration process, verify the
state, configure reviewers, and activate it only after those gates pass.
Enrichment workers never create schema dynamically.

### Review permission provisioning

- Exact permission code: `product_enrichment:review`.
- Name: `Product Enrichment Review`.
- Mechanism: forward data migration `20260821000000.sql`, using the existing
  unique `permissions.code` constraint and `ON CONFLICT (code) DO NOTHING`.
- The permission is deliberately not inserted into `role_permissions`.
- Because the Atlas directory is shared by master and tenant migration runs,
  the default runner may create an inert permission metadata copy in master.
  Tenant review authorization must use only the tenant-local repository/RBAC;
  the master copy is never an authorization source for tenant review.
- `CheckUserHasPermission` resolves by stable permission code through the
  local `permissions -> role_permissions -> user_roles` joins. Numeric IDs are
  not used by the migration or future permission lookup.

Role assignment for `product_enrichment:review` remains an explicit
deployment/architect decision. No admin, manager, owner, superadmin, product
manager, or other production role is automatically granted it.

### Legacy master suggestion quarantine

Legacy master `product_enrichment_suggestions` rows remain QUARANTINED / INERT.
They were not deleted, updated, moved, marked failed, copied, or backfilled.
No recovery mapping is inferred from organization ID, product ID, SKU,
suggestion ID, or user ID.

Static review confirms there is no master `ProductEnrichmentWorker`, no master
`ProductEnrichmentExecutionStore` used for active processing, no Stage 2D route,
no cleanup/application process for master suggestions, and no automatic tenant
suggestion copy. F2 uses the master repository only for active tenant registry
discovery; each worker/store is constructed from the selected tenant pool.
Existing F2 isolation tests confirm legacy master suggestions are not processed
and colliding tenant-local IDs remain isolated.

### F3 verification

- Go 1.26.7 was used from `C:\Program Files\Go\bin\go.exe`; no installation
  or upgrade was performed.
- `gofmt` passed for the modified Go file; the existing F2 modified test and
  enrichment/server files were also clean under `gofmt -l`.
- `cd packages/core && go test ./enrichment`, `./middleware`, `./repository`,
  `./usecase`, and `./...` passed.
- `cd apps/cloud-server && go test ./...` passed, including
  `cmd/migrate-tenants` compilation.
- `cd apps/sap-agent && go test ./...` passed.
- Atlas v1.3.2-4bf5fb9-canary was already available at
  `C:\Users\AnnsMustafa\atlas\atlas.exe`; `atlas migrate hash --dir
  file://db/migrations` and `atlas migrate validate --dir
  file://db/migrations` both passed. No migration was applied.
- SQLC was not required: no query source or generated repository file changed.
- No live OpenAI call, tenant database connection, or production deployment
  was performed.

### F3 files changed

- Permission migration: `packages/core/db/migrations/20260821000000.sql`.
- Atlas manifest: `packages/core/db/migrations/atlas.sum`.
- Migration tooling: `apps/cloud-server/cmd/migrate-tenants/main.go`.
- Tests: no new test file; existing F2 tenant-isolation/supervisor tests were
  rerun and passed.
- Generated files: none; SQLC was not needed.
- Worklog: `agents.md`.

### Deployment checklist before enabling Stage 2D

For each intended tenant, operations must verify:

1. The tenant exists and is active in the master registry only after its DB
   migration gate is complete.
2. The tenant DB connection is valid without exposing its DSN in logs.
3. The established runner has been run with the approved migration directory;
   use `-status` first for non-mutating inspection.
4. `20260820000000.sql`, `20260820010000.sql`, and `20260821000000.sql` are
   present in the tenant's Atlas migration state.
5. The tenant-local `product_enrichment:review` permission exists exactly once.
6. Approved reviewer role/user mappings are configured explicitly; no default
   production role assignment is assumed.
7. F1 tenant-local SAP migration and tenant-bound machine credentials are
   deployed.
8. F2 tenant supervisor and tenant pool/repository wiring are deployed.
9. Tenant-local suggestions can reach `in_review` through F2.
10. Users re-authenticate after tenant-bound JWT deployment where required.
11. OpenAI enrichment is enabled only when intentionally configured for the
   deployment.

## Stage 2D

Status: READY FOR IMPLEMENTATION - DEPLOYMENT PREREQUISITES STILL APPLY.

Next Action: Implement Stage 2D tenant-local human review/approval API.

The approved Stage 2D boundary remains:

`browser -> tenant-bound user JWT -> TenantBindingMiddleware ->
TenantMiddleware -> tenant repository -> tenant-local authenticated user ->
server-derived users.organization_id -> CheckUserHasPermission(
product_enrichment:review) -> tenant-local suggestions/product context ->
stale validation -> tenant-local approve/reject transaction -> tenant-local
audit_logs`.

Stage 2D scope remains whole-suggestion MVP, approve/reject only, no editing,
stale source blocks approval, `brand`/`category` `PROPOSE_NEW` blocks approval,
unsupported semantics informational, approved is not applied, and Stage 2E is
separate. Master DB must not participate in review business-data access.

## Stage 2D - Tenant-Local Human Review API

Stage 2D implementation completed on 2026-08-21. No Stage 2E application,
product mutation, taxonomy creation, OpenAI invocation, role auto-assignment,
commit, or push was performed.

### Routes and security boundary

- `GET /api/product-enrichment/suggestions`
- `GET /api/product-enrichment/suggestions/:id`
- `POST /api/product-enrichment/suggestions/:id/approve`
- `POST /api/product-enrichment/suggestions/:id/reject`

Routes are registered in the existing `/api` group after
`JWTAuthMiddleware -> TenantBindingMiddleware -> TenantMiddleware`. The
handler obtains the request-local repository from `middleware.RepoKey` and the
same tenant pool from `TenantPoolFromContext`; it never constructs a pool from
client input and has no master-repository fallback.

The authenticated JWT user ID is read from the trusted middleware context,
loaded from the selected tenant database, and its `users.organization_id` is
the only organization scope used by review queries. The exact tenant-local
permission `product_enrichment:review` is checked through the existing
database-backed `CheckUserHasPermission` query. Role names, `products:manage`,
headers, query values, and request bodies cannot grant or broaden review
authority. Review mutations accept an empty body only.

### Review behavior

The MVP is whole-suggestion review only: approve or reject, no field-level
decisions, editing, reviewer notes, rejection reasons, or taxonomy creation.
List defaults to `status=in_review` and permits only `in_review`, `approved`,
or `rejected`, with bounded limit/offset pagination. List and detail responses
are explicit reviewer DTOs; they do not serialize repository models, raw
`structured_current`, provider payloads, prompts, secrets, operational
inventory, prices, tax, suppliers, or arbitrary metadata.

Detail reloads the current tenant-local product/enrichment context and
separates source identity, a safe inference snapshot projection, current
authoritative state, proposals, unsupported informational semantics, provider
context, review state, and computed safety analysis. Stage 2A's existing
`FingerprintSnapshot` and `StructuredCurrent` implementation is reused; no
second fingerprint algorithm was introduced. Stale source blocks approval and
leaves the row `in_review`, without rerunning OpenAI or updating the persisted
snapshot/fingerprint. Detail reports stale state without mutation.

Approval performs explicit precedence and safety checks in addition to the
fingerprint comparison: structured SAP brand/category remain authoritative,
`KEEP_EXISTING` requires current support, active canonical brand/category
targets are revalidated by ID/code, description proposals cannot replace a
populated description, product type is never a proposal target, and invalid
or prohibited Stage 2B proposal content blocks approval. Brand/category
`PROPOSE_NEW` block approval. Description `PROPOSE_NEW` is allowed only when
the current description remains empty. Unsupported semantics are returned as
informational evidence only and are never applied.

Approve and reject use guarded `in_review` transitions in the same tenant
transaction as a minimal `audit_logs` update event. The audit records the
tenant organization, suggestion/product/reviewer identifiers, old/new status,
and event name. Audit failure rolls back the lifecycle update. The SQL guard
ensures concurrent reviewers have one winner; the second receives conflict.
Approved rows retain `applied_at = NULL` and products remain unchanged.
Rejected stale or `PROPOSE_NEW` suggestions remain rejectable.

Legacy master suggestions remain quarantined/inert and are not surfaced or
used by Stage 2D.

### Stage 2D files changed

- Enrichment/domain: `packages/core/enrichment/review.go`.
- Usecase: review orchestration is provider-neutral in the enrichment domain;
  no shared mutable `SetRepository` usecase was introduced.
- Handler/routes: `packages/core/handler/product_enrichment_review.go`,
  `packages/core/routing/product_enrichment_review.go`, and
  `apps/cloud-server/main.go`.
- Repository/SQL: `packages/core/repository/product_enrichment_review_store.go`,
  `packages/core/repository/product_enrichment_store.go`, and
  `packages/core/queries/product_enrichment_review.sql`.
- Generated SQLC: `packages/core/repository/product_enrichment_review.sql.go`.
- Tests: `packages/core/enrichment/review_test.go`.
- Worklog: `agents.md`.

### Verification and deployment findings

- SQLC `v1.30.0` generation completed successfully. Only the new review
  companion was generated; no existing generated SQLC files required changes.
- No Atlas/schema migration was added; `atlas.sum` was not modified.
- `gofmt` completed for all human-authored Stage 2D Go files.
- `cd packages/core && go test ./...` passed.
- `cd apps/cloud-server && go test ./...` passed.
- No live tenant database, role mapping, or OpenAI request was used. F3
  deployment prerequisites remain operational: tenant migrations must be
  applied, reviewer role mappings configured, tenant-bound auth/F1/F2
  deployed, and OpenAI enabled only where intended.

### Current Phase

Stage 2D complete - SAFE TO KEEP.

Next Action: Stage 2E DESIGN GATE - approved tenant-local suggestion
application with current-state revalidation, explicit allowed-field
application rules, atomic product/audit updates, and no changes to protected
SAP fields. Do not implement Stage 2E now.

## Stage 2E Design — Approved Suggestion Application

### Design gate and repository evidence

Stage 2E is a design gate only. No application endpoint, product mutation,
background application worker, permission, SQL, schema, generated file, or
configuration change is implemented here. The current Stage 2D worktree is
review-only: its routes are mounted under
`JWTAuthMiddleware -> TenantBindingMiddleware -> TenantMiddleware`, its
repository is obtained from `middleware.RepoKey`, its organization is derived
from the authenticated tenant-local `users.organization_id`, and approval
updates only the suggestion plus a tenant-local audit row. Products remain
unchanged and `applied_at` remains NULL after approval.

The Stage 2E invariant is:

`approved -> same-tenant current-state revalidation -> allowed narrow product
write(s) -> audit -> applied`,

with `approved != applied`. Master DB access is not part of this flow. The
master database may continue to provide control-plane tenant discovery for
other workers, but it must never provide Stage 2E product, suggestion,
taxonomy, user, permission, or audit data.

### Exact approved proposal shapes

Stage 1 persists the following JSONB columns separately:
`proposed_brand`, `proposed_category`, `proposed_description`, and
`unsupported_semantics`. The proposal DTOs are in
`packages/core/enrichment/product_enrichment.go`; there is one whole-row
status/reviewer lifecycle and no field-level approval state.

- Brand `KEEP_EXISTING`: action, confidence, and optional evidence/explanation;
  no target ID/code. Stage 2D requires the current structured brand to be
  resolved. Stage 2E is a no-op.
- Brand `MATCH_EXISTING`: action, confidence, optional evidence/explanation,
  and at least one canonical `target_id` or `target_code`; `canonical_name` is
  descriptive context. Stage 2D revalidates that the target exists, is active,
  and matches every stored ID/code supplied. Stage 2E may set only
  `products.brand_id` when the current brand is still unresolved.
- Brand `PROPOSE_NEW`: action, required `canonical_name`, confidence, and
  optional evidence/explanation, with no existing target. Current Stage 2D
  blocks approval, so it must not reach application. It never creates a brand.
- Brand `NO_MATCH`: action, confidence, and optional evidence/explanation;
  no target ID/code. Stage 2E is a no-op.

Category has the same four action shapes and target rules. A category
`MATCH_EXISTING` target is validated against the active global
`product_categories` dictionary in the selected tenant database; categories
are not organization-owned in the current schema. Category `PROPOSE_NEW` is
blocked by Stage 2D and is never created by Stage 2E.

- Description `KEEP_EXISTING`: action, confidence, and optional
  evidence/explanation; no value. Stage 2E is a no-op.
- Description `PROPOSE_NEW`: action, `value`, confidence, and optional
  evidence/explanation. Stage 2B `NormalizeProposedDescription` trims outer
  whitespace, requires valid UTF-8, and caps the value at 500 Unicode runes.
  Stage 2D permits it only while the current description is empty by the same
  trimmed-whitespace rule. Stage 2E re-runs this validation and may set only
  `products.description`.
- Description `NO_MATCH`: action, confidence, and optional
  evidence/explanation; no value. Stage 2E is a no-op.
- `UNSUPPORTED_TARGET` is accepted by the generic DTO contract for defensive
  representation, but Stage 2D blocks it for brand, category, and description.
  Unsupported semantic records are informational JSON evidence only and never
  imply a product destination.

No proposal contains or authorizes `product_type`; no proposal field may be
interpreted as a product name, SKU, metadata, variant, inventory, price, tax,
UoM, barcode, supplier, or operational-flag update.

### Exact application matrix

| Target | Approved action | Stage 2E result |
|---|---|---|
| brand | `KEEP_EXISTING` | no product mutation; mark the safe suggestion applied |
| brand | `MATCH_EXISTING` | set only `products.brand_id` when current brand is unresolved and the exact active canonical target still validates |
| brand | `PROPOSE_NEW` | defense-in-depth conflict; remain approved; never create taxonomy |
| brand | `NO_MATCH` | no product mutation; mark applied |
| category | `KEEP_EXISTING` | no product mutation; mark applied |
| category | `MATCH_EXISTING` | set only `products.category_id` when current category is missing and the exact active canonical target still validates |
| category | `PROPOSE_NEW` | defense-in-depth conflict; remain approved; never create taxonomy |
| category | `NO_MATCH` | no product mutation; mark applied |
| description | `KEEP_EXISTING` | no product mutation; mark applied |
| description | `PROPOSE_NEW` | set only `products.description` when current description is NULL, empty, or whitespace-only and the value remains valid under the 500-rune contract |
| description | `NO_MATCH` | no product mutation; mark applied |
| any | unsupported semantics | informational only; no product mutation; mark applied if the rest of the persisted proposal is safe |

The matrix is whole-suggestion safe application. A malformed proposal,
unexpected action, prohibited target, invalid product type, stale source, or
canonical-target conflict is not a no-op success: it remains `approved` and
returns a conflict without mutation.

### Category application policy

`CATEGORY_MATCH_EXISTING_APPLICATION_SUPPORTED`

This is supported by the current contracts, not inferred from the SQL schema
alone. Stage 2A eligibility explicitly creates a `missing_category` gap;
`CategoryProposal.validate` permits `MATCH_EXISTING` when no structured
category is present; Stage 2D blocks replacement only when the current
structured category is populated and revalidates an active canonical target.
Therefore a missing-category exact canonical match may be applied. A populated
SAP category remains authoritative forever for this MVP and can never be
replaced or refined by AI.

### Brand policy

Structured SAP FirmCode/brand remains authoritative. An approved brand
`MATCH_EXISTING` is only a fallback while the current product still has no
resolved `brand_id`/canonical brand identity. Stage 2E must reload the current
product and refuse to write if a later SAP sync resolved any brand. It must
never overwrite that later brand with the approved AI target. Brand and
category IDs are nullable foreign-key columns; the current repository treats
NULL/invalid non-positive IDs as missing and uses positive ID or canonical code
identity as resolved state.

### Description policy

The only destination is `products.description`, whose database type is
unbounded `TEXT`; the current product update query uses a nullable text
parameter and does not normalize the value. Stage 2A eligibility, Stage 2D
stale/precedence logic, and Stage 2E must all use
`strings.TrimSpace(description) == ""` for missing. A non-whitespace
description is authoritative and must not be overwritten. Stage 2E must
re-run `NormalizeProposedDescription`, storing its trimmed valid UTF-8 result
and never a value over 500 runes.

The repository has no HTML/Markdown allowlist, sanitizer, or formatter. The
current code therefore does not enforce a separate plain-text syntax policy;
it accepts any valid UTF-8 string within the rune limit. Stage 2E must not
invent HTML/Markdown transformation or rendering semantics and must treat the
value as catalog text. A future strict plain-text requirement would be a
separate contract gate.

### Apply trigger and API contract

`EXPLICIT_APPLY_ENDPOINT_RECOMMENDED`

Keep human review and product execution visibly separate. The smallest safe
MVP is an authenticated tenant-bound endpoint:

`POST /api/product-enrichment/suggestions/:id/apply`

with an empty body only. The server derives suggestion, product,
organization, and applier identity; the client may not supply tenant,
organization, product ID, applier ID, field selection, proposed values,
overrides, `force`, or `ignore_stale`. A successful response should expose a
small result containing suggestion ID, product ID, `status: applied`,
`applied_at`, changed field names, and a safe current product summary. It must
not serialize raw provider payloads, prompts, secrets, or arbitrary product
metadata.

An approval-triggered write would blur the hard `approved != applied`
boundary and make human control/retry semantics less visible. A background
worker is possible later but adds a new durable execution path for a small,
operator-triggered deterministic operation.

### Application permission verdict

`SEPARATE_APPLY_PERMISSION_RECOMMENDED`

The existing `product_enrichment:review` permission is checked through
`CheckUserHasPermission` in the tenant repository and is appropriate for
listing, inspection, approve, and reject. Approval authority does not
automatically imply authority to mutate product master data. The apply
endpoint should require a new narrower `product_enrichment:apply` capability;
role names and `products:manage` must not substitute for it. This requires a
tenant-local RBAC data migration that inserts the permission only, with no
automatic role grants. Deployment must explicitly map it to the intended
role/user. The master copy is inert for Stage 2E business access.

### Tenant boundary and apply-time revalidation

Stage 2D's security boundary can be reused unchanged:

`tenant-bound JWT -> TenantBindingMiddleware -> TenantMiddleware -> RepoKey
tenant repository -> tenant-local authenticated user -> users.organization_id`

Every suggestion, product, taxonomy, user, permission, and audit query must
use the server-derived organization and selected tenant pool. There is no
tenant or organization route/body authority, no master fallback, and no
cross-tenant lookup. Brand/category taxonomy is globally keyed in the current
tenant schema, so target validation is tenant-database-local and exact by ID
and/or code, not organization-filtered or selected by name.

Apply-time validation is mandatory even after approval:

1. Begin one transaction on the selected tenant pool and lock/load the
   organization-scoped suggestion by ID.
2. Require `status = approved`; load its persisted proposal JSON and original
   `source_data_fingerprint`.
3. Lock/load the organization-scoped current product and reconstruct the same
   Stage 2A `EnrichmentSourceSnapshot`.
4. Recompute the existing `FingerprintSnapshot` and require equality with the
   approved row's original fingerprint and contract context.
5. Recheck SAP source/product identity, valid immutable product type, brand and
   category precedence, and description emptiness.
6. Revalidate exact active canonical brand/category targets by stored ID/code;
   never select a replacement by name.
7. Validate the action matrix and description contract.
8. Execute only the eligible narrow product update, guarded by the current
   missing-field predicates.
9. Transition `approved -> applied`, write tenant audit records, and commit.

The current source fingerprint is sufficient for the approved-state design:
Stage 2D can reach `approved` only after the current snapshot equals the
persisted inference fingerprint, and Stage 2E compares the current snapshot
to that same original fingerprint again. A separate review-time fingerprint
column is not required for the invariant “current apply state equals the
approved/inference source state.” This cannot detect an unobservable
F1->F2->F1 history, but no current schema field can provide that history and
it does not change the current-state equality guarantee. No application
foundation schema change is required.

If source changes after approval—brand resolves to Y, category becomes
populated, description becomes non-empty, ItemName changes, product type
changes, or another fingerprinted source/UoM value changes—Stage 2E must not
mutate, must leave the suggestion `approved`, and must return/report a
conflict. It must not silently reject, mark failed, rerun AI, or force apply.
Inventory/price-only changes are intentionally outside the fingerprint and do
not create a false enrichment-staleness conflict; they are never written by
Stage 2E.

### Canonical target validation

For a brand/category `MATCH_EXISTING`, apply-time lookup must confirm the
record still exists, is active under current active-state semantics, and
matches every canonical identity stored in the approved proposal. If the
proposal has both ID and code, both must match. If it has code only, the exact
code lookup supplies the ID; no name-based substitution is allowed. The
current schema has global `brands` and `product_categories` tables, so
“tenant-local” means the selected tenant database's active canonical records,
not an organization column that does not exist. A missing, inactive, changed,
or ambiguous target is a conflict with no product mutation and the suggestion
remaining approved.

### Product update and transaction strategy

Do not reuse the broad `UpdateProduct` query or round-trip a whole `Product`
struct: it exposes name, UoM, product type, tax, flags, and metadata and could
overwrite unrelated concurrent changes. Add a focused application query/store
that updates only `brand_id`, `category_id`, and/or `description`, with the
organization/product predicates and expected missing-state predicates. Use a
single tenant transaction with row locks:

`BEGIN -> SELECT approved suggestion FOR UPDATE -> SELECT product FOR UPDATE
-> load/revalidate current context and exact taxonomy targets -> narrow
conditional product UPDATE -> guarded approved-to-applied UPDATE -> product
and lifecycle audit INSERT(s) -> COMMIT`.

The product lock removes the validation/write TOCTOU window. Conditional
missing-field predicates are defense in depth: if a concurrent SAP/process
write wins the lock first, reload/revalidation reports conflict; if it waits,
the later writer observes the committed result and structured SAP precedence
remains authoritative. Target rows should be read with a lock mode that
prevents an active/code change or deletion between validation and the foreign
key write.

The existing `MarkProductEnrichmentSuggestionApplied` query already guards
`organization_id`, `id`, and `status = 'approved'`, and sets `status = applied`,
`applied_at = CURRENT_TIMESTAMP`, and `updated_at = CURRENT_TIMESTAMP`. It may
be reused inside the same transaction after the product update. It must never
be called outside that transaction. `reviewer_id` and `reviewed_at` remain
unchanged; there is no application-user column.

Only `approved -> applied` is legal. `in_review`, `rejected`, `retryable`, and
`failed` must not apply. A guarded transition returning no row rolls back any
product write. An already-applied row should return the current applied result
idempotently without another product write or reviewer-state change; the
first successful application remains the only mutation.

### Audit and no-op policy

`EXISTING_SCHEMA_SUFFICIENT_FOR_APPLIER_AUDIT`

The tenant-local `audit_logs` table already has organization, table/record,
action, old/new JSONB, changed fields, and `performed_by` referencing the
tenant-local user. No `applied_by` column is required when the authenticated
applier is recorded in the application audit row. Product update, suggestion
transition, and all audit inserts must commit or roll back together.

Use minimal tenant-local audit payloads, preferably two records in the same
transaction: one product `UPDATE` event and one suggestion application/status
event. Include suggestion ID, product ID, organization ID, authenticated
applier ID, old/new status, changed field names, and old/new brand/category
IDs when changed. For description, record only `description_changed` and
bounded presence/length or a non-reversible digest if operational evidence is
needed; do not store full old/new description content by default. Never store
raw AI/provider payloads, prompts, API keys, JWTs, or database credentials.

An approved safe suggestion containing only `KEEP_EXISTING`, `NO_MATCH`, or
unsupported informational semantics has completed its approved work even if
the mutation set is empty. Mark it `applied` with an auditable zero-change
application event. Invalid, stale, prohibited, or target-invalid rows are
conflicts and remain approved.

### Errors, force policy, retry, and stale recovery

For the explicit endpoint, use repository-compatible safe semantics: 401 for
missing/invalid authentication, 403 for missing `product_enrichment:apply`,
404 for an organization-scoped missing suggestion, 409 for non-approved or
already-invalid/stale/target-conflict application (with already-applied
retries treated as idempotent success), and 500 for transaction/audit failure
without exposing raw SQL/provider errors. The request has no force flag.

`NO FORCE APPLY` is mandatory. No client option may bypass stale validation,
structured SAP precedence, canonical-target validation, product type
protection, or the allowed-field matrix.

If an approved row becomes stale, leave it approved and historical. A future
operator/reviewer workflow may enqueue a new suggestion from the new F2
fingerprint; the existing uniqueness key
`(organization_id, product_id, source_data_fingerprint, contract_version)`
supports a distinct row. Do not silently convert the old approved row to
rejected/failed and do not rerun AI from Stage 2E.

### Schema verdict

`NO_APPLICATION_SCHEMA_CHANGE_REQUIRED`

The existing original source fingerprint is sufficient for current-state
apply revalidation; `applied_at`, reviewer fields, and tenant audit
`performed_by` provide lifecycle and applier accountability; and the existing
DBTX/`Queries.WithTx` boundary can support atomic narrow writes. A separate
RBAC permission migration is still required because least privilege recommends
`product_enrichment:apply`; that is deployment data provisioning, not an
application-foundation schema correction.

### Exact future Stage 2E patch surface

Only after this gate is separately authorized, the expected narrow surface is:

- `packages/core/enrichment/application.go` and focused application tests for
  matrix, revalidation, conflicts, no-op, and lifecycle behavior;
- `packages/core/repository/product_enrichment_application_store.go`;
- `packages/core/queries/product_enrichment_application.sql` with locked
  suggestion/product/target reads, field-specific conditional updates, and
  audit inserts;
- SQLC-generated application companion(s), with the existing guarded
  `MarkProductEnrichmentSuggestionApplied` reused or regenerated only as
  required by the source query set;
- `packages/core/handler/product_enrichment_application.go` and
  `packages/core/routing/product_enrichment_application.go` for the empty-body
  authenticated endpoint;
- `apps/cloud-server/main.go` for route wiring only;
- a tenant-local RBAC migration provisioning
  `product_enrichment:apply` if that permission recommendation is accepted;
- application/repository/handler tests and `agents.md`.

Do not change SAP mappings/extraction, F1 routing, F2 supervision, Stage 2A
eligibility/fingerprint semantics, Stage 2B output policy, Stage 2D review
semantics, the provider adapter, product-type classification, or unrelated
product/inventory/taxonomy systems.

### Implementation phases

- E0: provision `product_enrichment:apply` in each intended tenant database;
  map it explicitly to the deployment role; do not auto-grant it.
- E1: implement the tenant-local deterministic application domain/store,
  locked current-state queries, narrow product writes, guarded lifecycle
  transition, and atomic audit support.
- E2: add the explicit apply endpoint, authenticated applier/permission
  boundary, response DTO, conflict mapping, and cloud wiring.
- E3: run SQLC generation, formatting, focused/core/server tests, migration
  validation, concurrency/atomicity tests, and tenant deployment checks before
  enabling the route.

### Stage 2E test plan

- Auth/permission: no auth is 401; missing apply permission is 403; role name
  alone is insufficient; request cannot supply tenant, organization, applier,
  product, values, field selection, or force flags.
- Tenant/org: tenant A cannot see/apply tenant B; organization A cannot apply
  organization B; master suggestion collisions are invisible; every query is
  organization-scoped.
- Lifecycle: approved applies; in_review/rejected/retryable/failed do not;
  concurrent apply has one effective mutation; an already-applied retry is
  idempotent and does not rewrite product/reviewer state.
- Brand: exact canonical ID/code is applied only while missing; later
  structured brand causes conflict/no mutation; deleted/inactive/changed target
  causes conflict; KEEP_EXISTING and NO_MATCH are zero-change applied; a
  defense-in-depth PROPOSE_NEW cannot apply.
- Category: missing-category MATCH_EXISTING follows the supported policy;
  populated category is never replaced; target identity/active checks match
  brand behavior.
- Description: NULL/empty/whitespace-only accepts a valid normalized proposal;
  populated description conflicts; exact trimmed text and 500-rune/UTF-8
  boundaries are enforced; no HTML/Markdown transformation is introduced.
- Protected fields: SKU, name, product type, flags, metadata, UoM,
  inventory, pricing, barcode, supplier, variants, and unrelated columns stay
  unchanged.
- Stale/concurrency: ItemName, product type, brand, category, description,
  UoM, or other fingerprinted source change blocks apply; inventory/price-only
  change does not falsely stale; concurrent SAP update cannot be overwritten.
- Atomicity/audit: product failure leaves suggestion approved; lifecycle
  failure rolls back product; audit failure rolls back both; audit identifies
  authenticated applier and changed fields without raw AI secrets/content.
- No-op: safe zero-change approved suggestions reach applied with an explicit
  zero-change audit event; invalid or stale zero-change-looking rows remain
  approved with conflict.

### Deployment prerequisites and status

Before enabling Stage 2E, all intended tenant databases must have the Stage 1,
Stage 2C, Stage 2D, and any E0 RBAC migrations applied; tenant-bound JWT and
F1/F2 routing/supervision must be deployed; Stage 2D review permission and
role mappings must be configured; the separate apply permission must be
explicitly mapped if provisioned; and at least one approved tenant-local
suggestion must exist. No master business-data fallback is permitted.

Stage 2E status: DESIGN GATE.

Next Action: if the separate apply permission is accepted, provision that
tenant-local permission first; otherwise separately authorize implementation
of the explicit tenant-local deterministic application endpoint. Do not
implement Stage 2E as part of this design gate.

## Stage 2E E0 — Apply Permission

- Exact permission: `product_enrichment:apply`.
- Semantic scope: may apply an already-human-approved product enrichment
  suggestion to the tenant-local product using the deterministic Stage 2E
  application rules. It does not approve/reject suggestions, edit proposal
  values, create taxonomy, or mutate unrelated product domains.
- Provisioning uses `packages/core/db/migrations/20260821010000.sql`, the
  same idempotent `permissions` insert by unique `code` used for
  `product_enrichment:review` in `20260821000000.sql`.
- The permission is tenant-local in actual authorization because the future
  applier will use the physical tenant repository. A shared migration may
  leave an inert permission row in master; master RBAC must never authorize
  tenant Stage 2E.
- No role or user is granted `product_enrichment:apply` by this migration.
  `product_enrichment:review` does not imply it. Deployment administrators
  must explicitly map review and apply authority independently.
- The permission must exist in each intended tenant database before Stage 2E
  is enabled. Until an approved application role/user is explicitly assigned
  it, a future Stage 2E endpoint must return `403`.
- Atlas hash/validation passed with Atlas v1.3.2-4bf5fb9-canary. Regression
  tests passed with `C:\Program Files\Go\bin\go.exe` for `packages/core`,
  `apps/cloud-server`, and `apps/sap-agent`. No SQLC query or generated SQLC
  file is in scope.
- E0 files changed: this migration, `packages/core/db/migrations/atlas.sum`,
  and `agents.md`.

Stage 2E status: E0 complete. The migration hash/validation and all requested
regression checks passed.

Next Action: E1 — implement deterministic tenant-local approved-suggestion
application domain/store/transaction logic. No HTTP endpoint yet.

## Stage 2E E1 - Deterministic Tenant-Local Application Engine

E1 is implemented as a domain/store/transaction layer only. Stage 2E is not
complete; E2 remains the next phase and will add the authenticated HTTP apply
endpoint.

### Application API and tenant boundary

- Domain entry point:
  `enrichment.ProductEnrichmentApplicationService.ApplyApprovedSuggestion`.
- Inputs are explicit trusted domain values: positive `organizationID`,
  `suggestionID`, and authenticated `applierUserID`. The service accepts no
  tenant slug, product ID, proposed values, force flag, field selection, or
  taxonomy override, and performs no HTTP permission check.
- The application store is constructed from the tenant-local `Queries` and
  `pgxpool.Pool` selected by the request. E1 never uses `masterRepo` and never
  queries master users or master business data.

### Transaction and locking

- One tenant-local PostgreSQL transaction contains suggestion lock/load,
  product lock/load, current-state validation, canonical target revalidation,
  narrow product update, guarded lifecycle transition, audit insert, and
  commit.
- Lock order is deterministic: approved suggestion `FOR UPDATE` first, then
  product `FOR UPDATE`; canonical target rows are locked only after the
  product. Any error rolls back the transaction.
- The product row lock and conditional update prevent an E1 application from
  overwriting a structured SAP field that changed before the final write. A
  later SAP write waits for the lock and remains authoritative under the
  existing deterministic SAP synchronization behavior.

### Apply-time validation and application matrix

- E1 reuses the exact Stage 2A `StructuredCurrent` and
  `FingerprintSnapshot` implementations. Fingerprint mismatch, source identity
  change, invalid product type, or changed product context returns stale/conflict
  and leaves the suggestion approved.
- Structured SAP precedence is checked again: resolved brand/category blocks
  replacement, and non-whitespace description blocks `PROPOSE_NEW`.
- Brand and category `MATCH_EXISTING` apply only to a still-missing field and
  only after exact ID/code revalidation against an active canonical row.
  Missing, inactive, mismatched, or ambiguous targets conflict; no taxonomy is
  created. `KEEP_EXISTING` and `NO_MATCH` do not mutate.
- Brand/category `PROPOSE_NEW` and unsupported target actions are conflicts.
  Unsupported semantics remain informational and never create attributes,
  families, variants, or other schema records.
- Description `PROPOSE_NEW` applies only to NULL/empty/whitespace current
  descriptions after UTF-8 and 500-rune validation through the existing
  `NormalizeProposedDescription` contract. Existing Stage 2B/2C canonical
  trimming is persisted; no HTML sanitizer or rewriting was added.
- The explicit `ApplyPlan` contains only `brand_id`, `category_id`,
  `description`, and `ChangedFields`. No product object or broad product update
  is passed to persistence.

### Lifecycle, no-op, idempotency, and audit

- Only `approved -> applied` is accepted. `applied` is a successful idempotent
  result with `AlreadyApplied=true`; it does not rewrite the product, reviewer,
  applier, or audit state. All other statuses are rejected.
- Approved no-op suggestions containing only `KEEP_EXISTING`, `NO_MATCH`, or
  informational unsupported semantics skip the product update, transition to
  `applied`, set `applied_at`, preserve `reviewer_id`/`reviewed_at`, and write a
  zero-change audit event.
- The existing guarded lifecycle query sets `status`, `applied_at`, and
  `updated_at` only, preserving reviewer and proposal/fingerprint fields.
- The application audit is inserted in the same transaction with event
  `product_enrichment.applied`, tenant organization, suggestion/product IDs,
  applier in `performed_by`, changed fields, old/new brand/category IDs when
  changed, and `description_changed`. Full description text and provider
  secrets/payloads are not stored.
- Product-write, lifecycle, audit, or commit failure cannot leave a committed
  partial application: the transaction rolls back and the suggestion remains
  approved.

### E1 files changed

- Domain: `packages/core/enrichment/application.go`.
- Store/transaction: `packages/core/repository/product_enrichment_application_store.go`.
- SQL: `packages/core/queries/product_enrichment_application.sql`.
- Generated SQLC: `packages/core/repository/product_enrichment_application.sql.go`.
- Tests: `packages/core/enrichment/application_test.go`.
- Worklog: `agents.md`.

No handler, route, `apps/cloud-server/main.go`, permission behavior, SAP
mapping, Stage 2D review semantics, product schema, migration, or Atlas
checksum was changed. E1 writes only the three approved product columns,
normal `updated_at`, the approved suggestion lifecycle fields, and
`audit_logs`.

### Verification and status

- SQLC v1.30.0 was located and real generation completed successfully with
  `cd packages/core && sqlc generate`; the generated companion was not hand
  edited.
- `gofmt` completed on all human-authored E1 Go files and `git diff --check`
  passed.
- Passed `go test ./...` in `packages/core`, `apps/cloud-server`, and
  `apps/sap-agent` using `C:\Program Files\Go\bin\go.exe`.
- No Atlas command or live/external database operation was required because E1
  has no schema migration. No live database was contacted.

Stage 2E status: E1 complete and verified. Stage 2E overall remains incomplete.

Next Action: E2 - add the explicit authenticated
`POST /api/product-enrichment/suggestions/:id/apply` endpoint using tenant-bound
RepoKey and exact `product_enrichment:apply` permission. Do not implement E2 or
E3 as part of E1.

## Stage 2E E2 - Explicit Apply Endpoint

E2 adds the explicit authenticated tenant-local application endpoint over the
verified E1 application service. Approval remains a separate Stage 2D action;
the approval handler does not call E1.

### Endpoint and security chain

- Exact route: `POST /api/product-enrichment/suggestions/:id/apply`.
- The route is registered in the same `/api` protected group as Stage 2D:
  `JWTAuthMiddleware -> TenantBindingMiddleware -> TenantMiddleware -> handler`.
- The handler obtains the tenant-local repository and pool from middleware
  context (`middleware.RepoKey` / `TenantPoolFromContext`). It has no
  `masterRepo` application path or secondary unscoped lookup.
- The authenticated user ID comes only from trusted JWT auth context. The
  handler loads that user through the selected tenant repository and derives
  `organization_id` from the tenant-local `users` row.
- Authorization requires exactly `product_enrichment:apply`. Review
  permission, role names, `products:manage`, and administrative labels are not
  substitutes. Review and apply permissions remain independently provisioned.

### Request, E1 invocation, and response

- The request body is empty; client fields for tenant, organization, product,
  applier/reviewer, target values, field selection, status, product type, or
  force/override behavior are rejected.
- The handler constructs a request-local E1 service from the tenant repository
  and pool, then invokes exactly:
  `ApplyApprovedSuggestion(ctx, organizationID, suggestionID, userID)`.
- The safe response contains only `suggestion_id`, `product_id`, `status`,
  `applied_at`, `changed_fields`, and `already_applied`. Changed fields are
  restricted to `brand_id`, `category_id`, and `description`.
- E1 `AlreadyApplied` results return `200 OK` with `status=applied` and
  `already_applied=true`; the handler performs no second mutation or audit.

### Error and isolation behavior

- Malformed suggestion IDs return `400`.
- Missing tenant-local suggestions return `404`.
- Non-approved, stale, invalid/prohibited, canonical-target, and conditional
  application conflicts return `409` with sanitized application codes.
- Persistence, transaction, audit, and unexpected failures return sanitized
  `500`; raw SQL, provider/OpenAI, JWT, DSN, and database details are omitted.
- Missing authentication is handled by the existing middleware as `401`;
  missing tenant-local user rows fail safely as `401`; missing apply permission
  is `403`.
- Tenant binding and tenant repository selection remain the first boundary.
  Organization scope is server-derived and enforced again by E1, so tenant and
  organization collision/guessing cannot select another business database or
  organization.
- There is no force apply, OpenAI call, Stage 2C invocation, review auto-apply,
  schema change, query change, SQLC regeneration, or direct product mutation in
  the handler/routing/cloud wiring. E1 remains the only application write
  layer.

### E2 files changed

- Handler: `packages/core/handler/product_enrichment_application.go`.
- Routing: `packages/core/routing/product_enrichment_application.go`.
- Cloud wiring: `apps/cloud-server/main.go`.
- Tests: `packages/core/handler/product_enrichment_application_test.go` and
  `packages/core/routing/product_enrichment_application_test.go`.
- Shared permission constant: `packages/core/enrichment/application.go`.
- Worklog: `agents.md`.

### Verification and status

- `gofmt` completed on all modified Go files and `git diff --check` passed.
- Passed focused handler/routing tests, `go test ./enrichment`,
  `./repository`, `./handler`, `./middleware`, and `./usecase` in
  `packages/core`.
- Passed `go test ./...` in `packages/core`, `apps/cloud-server`, and
  `apps/sap-agent` using `C:\Program Files\Go\bin\go.exe`.
- No SQL, SQLC-generated file, schema, migration, or Atlas checksum changed;
  no SQLC or Atlas operation was required. No live/external database was
  contacted.

Stage 2E status: E2 complete and verified. Stage 2E overall remains
incomplete pending E3.

Next Action: E3 - final end-to-end Stage 2E / tenant-security /
deployment-readiness verification and MVP completion gate. Do not mark the
full MVP complete until E3 passes.

## Stage 2E E3 — Final MVP Verification

Review date: 2026-08-21.

### Scope and final architecture

E3 was a verification-only gate. No source logic, generated artifact, schema,
migration, deployment, database, OpenAI call, commit, or push was performed.
The only E3 file modification is this worklog.

The verified active flow is:

`SAP Agent -> tenant/org-bound M2M JWT -> TenantBindingMiddleware ->
TenantMiddleware -> tenant pool/RepoKey -> tenant-local SAPMigrationUseCase ->
commit -> tenant-local Stage 2A enqueue -> pending suggestion -> F2 active-tenant
supervisor -> tenant-local worker -> shared OpenAI adapter -> strict Stage 2B
validation -> in_review -> tenant-local human review -> approved/rejected ->
explicit apply permission -> E1 suggestion lock then product lock -> fingerprint
and structured-SAP revalidation -> narrow product update -> applied audit ->
commit.`

SAP migration and enrichment are separate after commit. Enqueue failure is
best-effort and cannot roll back a committed SAP migration. The worker never
mutates products. Review never invokes OpenAI or E1. Approval never auto-applies.

### Tenant data ownership invariant

- Master DB is used for tenant registry/control-plane discovery, tenant
  connection metadata, active state, and migration coordination.
- Each active tenant DB owns organizations, users, roles/permissions, products,
  brands/categories, SAP staging, product enrichment suggestions, and audit logs.
- For active tenant business operations, Product DB = Suggestion DB = Worker DB
  = Review DB = Application DB = Audit DB: all are the same selected tenant
  pool/repository.
- `masterEnrichmentTenantRegistry` reads only active tenant identities. The F2
  factory resolves each slug through the manager, constructs a fresh
  `repository.New(tenantPool)`, and creates the worker from that repository.
- Review and application handlers obtain `RepoKey` and `TenantPoolFromContext`;
  no master fallback or secondary unscoped lookup exists in their stores.
- Legacy master business rows can exist because the shared migration directory
  is also usable for the control-plane database, but no enrichment worker,
  review route, apply route, copy path, or inferred tenant mapping processes
  them. They are quarantined/inert.

### Authentication and SAP boundary

- Standard user JWTs contain the signed `tenant_slug`; protected routes require
  `JWTAuthMiddleware -> TenantBindingMiddleware -> TenantMiddleware` and exact
  `x-tenant-id` equality.
- SAP migration is machine-only. The JWT and registered client must agree on
  `tenant_slug` and positive `organization_id`; the token type must be `machine`
  and the client must be active/whitelisted.
- `x-tenant-id` is a tenant slug, never an organization ID. Optional
  `x-organization-id` and payload organization values are consistency checks;
  the trusted M2M organization claim is authoritative.
- The selected tenant organization is validated in that tenant DB before the
  migration handler constructs the tenant-local use case. No default
  organization `1` remains in this path.
- The SAP use case begins and commits its transaction using the selected tenant
  pool. Its post-commit coordinator uses the same request-local tenant repo.

### F2 worker verification

- The supervisor enumerates active tenants from master only, on every cycle.
  Disabled tenants are not returned/are skipped, and newly active tenants are
  rediscovered without restart.
- Each tenant gets an isolated pool, repository, store, worker pass, and logger.
  Setup and record failures are isolated so one tenant does not stop others.
- Due suggestions, product snapshots, UoM context, and candidate dictionaries
  are loaded from the tenant repository. Numeric ID collisions across tenant
  databases are therefore harmless.
- OpenAI is the shared provider adapter only; it carries no tenant identity.
  Strict request/response validation precedes persistence. Outcomes are only
  tenant-local `in_review`, `retryable`, or `failed`.
- No master suggestion query or product mutation is reachable from the worker.

### AI safety and strict contract

- Valid immutable product types are exactly `standard`, `raw_material`,
  `fixed_asset`, and `finished_good`; `product_type` is absent from proposal
  DTOs and output schema.
- Structured SAP brand/category/description values remain authoritative.
  Resolved brand/category require `KEEP_EXISTING`; populated description cannot
  be replaced.
- The contract and parser reject authoritative targets including SKU/ItemCode,
  product name as a mutation target, barcode ownership, inventory/stock,
  warehouse/store, price, tax, UoM/base UoM/conversions, supplier, active,
  sellable, purchasable, SAP identity, and product type. Unsupported semantics
  are informational JSON only and are never applied.
- Candidate matching is exact ID/code matching against server-supplied active
  canonical candidates. Model-authored names are not identity and no fuzzy
  fallback exists.
- Description proposals are trimmed, valid UTF-8, non-empty, and at most 500
  Unicode runes.

### Review verification

- Review routes are behind JWT authentication, tenant binding, tenant-local
  repository selection, tenant-local user lookup, and exact
  `product_enrichment:review` authorization.
- Organization is derived from the authenticated tenant-local user; the client
  cannot provide organization, tenant, reviewer, product, or proposal values.
- Review is whole-suggestion only. Bodies must be empty; proposals cannot be
  edited, taxonomy cannot be created, and `PROPOSE_NEW` brand/category actions
  cannot be approved.
- Current product context and the canonical Stage 2A fingerprint are checked
  before approval. Stale approval is blocked. Reject remains available for
  stale or unresolved/`PROPOSE_NEW` suggestions.
- Approve/reject use a tenant transaction for the guarded status transition and
  review audit. Concurrent approve/reject requests have one winning transition.

### Apply verification

- The only application route is authenticated, tenant-bound
  `POST /api/product-enrichment/suggestions/:id/apply`.
- The handler accepts an empty body only. It accepts no tenant, organization,
  product ID, applier ID, brand/category/description, field selection, force,
  override, ignore-stale, or status input.
- The handler derives user and organization from the tenant-local user row,
  checks exactly `product_enrichment:apply`, and invokes only
  `ApplyApprovedSuggestion(ctx, organizationID, suggestionID, userID)`.
- E1 locks the suggestion first and product second in one tenant transaction;
  checks approved state; reloads current context; reuses `FingerprintSnapshot`;
  revalidates structured precedence and canonical targets; builds an explicit
  narrow plan; performs the conditional write; transitions approved to applied;
  inserts audit; and commits. Any failure rolls back.
- `KEEP_EXISTING`, `NO_MATCH`, and unsupported informational-only proposals are
  zero-product-change applications that still transition approved to applied
  and write a zero-change audit.
- A second committed apply sees `AlreadyApplied`, returns HTTP 200, performs no
  product update, no second transition, and no duplicate audit.
- There is no force path and no OpenAI/review invocation during apply.

### Final product write surface

The Stage 2E application SQL is `ApplyProductEnrichmentFields`. Its only
product assignments are `brand_id`, `category_id`, `description`, and normal
`updated_at`. It is organization- and product-scoped and uses conditional
null/blank guards. No broad whole-product update is reachable from E1/E2.

Stage 2E cannot write SKU, name, product type, base UoM, metadata,
track-inventory, active/sellable/purchasable flags, variants, prices,
inventory, tax, UoM/conversions, barcodes, or suppliers.

### Staleness and canonical targets

Apply reuses the canonical Stage 2A `FingerprintSnapshot`; there is no second
application fingerprint. ItemName/source-name, product type, structured brand,
structured category, description, and fingerprinted UoM/conversion context are
protected. Inventory/pricing-only state is absent from the fingerprint and does
not falsely stale a suggestion.

Stale, invalid, missing, inactive, or ID/code-mismatched canonical targets
leave the product untouched, leave the suggestion approved, emit no application
audit, and map through E2 to HTTP 409. Brand/category `PROPOSE_NEW` conflicts;
there is no creation or silent retargeting.

### Atomicity, audit, concurrency, and isolation

- Review status transition plus review audit are one tenant transaction.
- Application product mutation, applied lifecycle transition, and application
  audit are one tenant transaction.
- Application audit records tenant organization, suggestion/product IDs,
  authenticated applier, approved-to-applied status, changed fields, old/new
  brand/category IDs where applicable, and `description_changed`. It does not
  store API keys, JWTs, prompts, raw provider responses, DSNs/passwords, full
  description text, or provider secrets.
- Suggestion-first/product-second lock ordering, guarded lifecycle SQL, product
  conditional guards, and fingerprint revalidation provide the required race
  behavior. Normal tests cover application concurrency seams and idempotency.
- Tenant A and tenant B may contain identical organization/product/suggestion/
  user numeric IDs; each operation remains on its tenant pool. Within one
  tenant DB, all suggestion/product queries carry the server-derived
  organization ID, giving scoped-not-found behavior for another organization.

### Permission separation and route inventory

The exact production authorization points are:

- Stage 2D review handler: `product_enrichment:review`.
- Stage 2E E2 application handler: `product_enrichment:apply`.

The two migrations provision the permission rows separately and grant neither to
any role/user. Role names, `products:manage`, review permission, and admin
labels are not substitutes. Any inert master permission copies cannot authorize
tenant business work because authorization uses the tenant repository.

Verified routes, all under the protected `/api` group:

- `GET /api/product-enrichment/suggestions`
- `GET /api/product-enrichment/suggestions/:id`
- `POST /api/product-enrichment/suggestions/:id/approve`
- `POST /api/product-enrichment/suggestions/:id/reject`
- `POST /api/product-enrichment/suggestions/:id/apply`

There is no public enrichment route, alias, bulk approve/apply, apply-all,
force-apply, automatic-apply, or GET apply route.

### Migration chain and tenant migrator

The final chain is present and ordered:

1. `20260820000000.sql` — enrichment foundation.
2. `20260820010000.sql` — retry durability.
3. `20260821000000.sql` — `product_enrichment:review`.
4. `20260821010000.sql` — `product_enrichment:apply`.

`cmd/migrate-tenants` reads the master registry, optionally migrates master,
then migrates active tenant connection strings independently. `-master=false`
skips master migration but still reads the registry. Tenant failures are logged,
isolated, summarized, and produce a non-zero exit; no fallback to master exists.
The F2 worker never runs migrations.

`CreateTenant` registers metadata only; it does not provision or migrate the new
database and can register an active tenant before operational migration. The
required runbook remains: register/provision DB -> run tenant migrations ->
verify migration state -> configure permissions -> activate/use business
features.

### SQLC, Atlas, formatting, tests, and secret scan

- SQLC v1.30.0 was located. Read-only `sqlc compile -f packages/core/sqlc.yaml`
  passed; no regeneration was performed.
- Atlas `v1.3.2-4bf5fb9-canary` was located. Read-only migration validation and
  hash checks passed; no migration was applied and `atlas.sum` was not rewritten.
- `C:\Program Files\Go\bin\gofmt.exe -l` over all E2 human-authored Go files
  returned no files.
- Uncached tests passed: core focused enrichment/repository/handler/
  middleware/usecase packages, core `go test -count=1 ./...`, cloud-server
  `go test -count=1 ./...`, and SAP-agent `go test -count=1 ./...`.
- `go test -race -count=1 ./enrichment` was attempted. Windows Go reports
  `-race requires cgo`; `CGO_ENABLED=0` and no GCC toolchain is present. No
  extra tooling was installed. This is recorded as a tooling limitation, not
  an MVP code blocker, because normal concurrency/idempotency tests passed.
- The uncommitted diff and E2 untracked files were scanned for OpenAI keys,
  bearer tokens, JWT secrets, M2M tokens, DSNs/passwords, SAP credentials, and
  private keys. No suspicious credential was found.
- `git diff --check` passed.

### Complete static flow trace

Tenant A: `M2M A/org1 -> signed tenant-a and org1 -> tenant A registry lookup
-> tenant A DB -> product 95 -> post-commit suggestion 1 pending -> active
tenant A worker -> tenant A product/context/candidates -> OpenAI -> strict
validation -> suggestion 1 in_review -> tenant A reviewer JWT/RepoKey/org1 /
review permission -> approved -> tenant A applier JWT/RepoKey/org1 / apply
permission -> suggestion lock -> product 95 lock -> revalidation -> allowed
narrow update -> applied -> tenant audit -> commit`.

Tenant B may use the same numeric IDs, but every read/write above uses tenant B's
separate pool. Master performs control-plane discovery only and never owns an
active product, review, application, or audit transaction.

### Deployment checklist

1. Deploy tenant-bound user JWT changes.
2. Re-login users so old unbound JWTs disappear.
3. Reissue SAP M2M credentials with `tenant_slug` and `organization_id`.
4. Deploy F1 tenant-local SAP routing.
5. Run approved tenant migrations.
6. Verify all four enrichment/RBAC migrations per intended tenant.
7. Deploy the F2 tenant worker supervisor.
8. Configure/enable OpenAI only where intended.
9. Assign `product_enrichment:review` explicitly to reviewers.
10. Assign `product_enrichment:apply` explicitly to appliers.
11. Deploy the Stage 2D review API.
12. Deploy E1/E2 application paths.
13. Verify one tenant-local suggestion reaches `in_review`.
14. Review it.
15. Explicitly apply it.
16. Verify only allowed product fields changed.
17. Verify tenant-local audit records.
18. Confirm a second tenant with colliding IDs is unaffected.
19. Keep legacy master suggestions quarantined.
20. Monitor worker and application errors after rollout.

No deployment step was executed during E3.

### Production distinction and optional work

`CODE MVP COMPLETE` is supported by the source audit, SQLC/Atlas checks,
formatting check, and uncached normal Go tests. `PRODUCTION DEPLOYMENT VERIFIED`
is not claimed. No live tenant DB, external database, migration application,
OpenAI call, or production smoke test was performed:

`NOT PERFORMED — DEPLOYMENT/ENVIRONMENT VALIDATION`.

Optional follow-up work remains: deterministic alias/rule learning; reviewer and
application UI; observability/metrics; live OpenAI smoke testing; tenant
provisioning automation; tenant DSN rotation hardening; stale-approved
re-enrichment UX; PROPOSE_NEW taxonomy resolution; field-level review; reviewer
notes/rejection reasons; product attributes/family schema; candidate/taxonomy
re-enrichment strategy; legacy master recovery if authoritative mapping becomes
available; and separate migration/gRPC security hardening if still open.

### E3 acceptance result

CRITICAL findings: none.

IMPORTANT findings: none.

MINOR findings: normal Go race instrumentation is unavailable on this host;
new-tenant provisioning remains an explicit operational prerequisite and is not
automatic. Neither blocks the code MVP under the approved E3 criteria.

Stage 2E: COMPLETE

Core AI Product Enrichment MVP: CODE COMPLETE

Next Action: deployment/environment validation and optional follow-up
enhancements. Do not state that production deployment is verified until it is
actually performed.

## DeepSeek Product Enrichment Provider

### Completion

- The provider-neutral architecture remains unchanged:
  `ProductEnrichmentWorker -> ProductEnrichmentProvider`.
- Provider selection is configuration-driven in `apps/cloud-server/main.go`.
  `openai` keeps the existing OpenAI adapter; `deepseek` selects the new
  `packages/core/enrichment/deepseekadapter` adapter; unknown providers fail
  safely.
- The DeepSeek adapter uses a small adapter-local HTTP client for the
  OpenAI-compatible `POST /chat/completions` endpoint. The configured base URL
  is normalized and `/chat/completions` is appended. The default is
  `https://api.deepseek.com`.
- Configuration names are `DEEPSEEK_API_KEY`, `DEEPSEEK_BASE_URL`, and
  `DEEPSEEK_MODEL`. The default model is `deepseek-v4-flash`; model selection
  remains configuration-driven. The key is required only when enrichment is
  enabled and the selected provider is `deepseek`.
- DeepSeek requests use `response_format: {"type":"json_object"}` and an
  explicit JSON-only instruction. JSON mode is not treated as schema or
  security validation.
- DeepSeek output is extracted from the Chat Completions envelope and always
  passes through the existing Stage 2B strict parser, candidate dictionary
  validation, structured SAP precedence, prohibited-target validation, and
  provider-neutral result construction. No malformed JSON repair or model
  supplied provider identity is accepted.
- Business safety is identical to OpenAI. Product type is never an output
  target; only `standard`, `raw_material`, `fixed_asset`, and `finished_good`
  remain valid immutable context. SKU/ItemCode, SAP identity, barcodes,
  inventory, pricing, tax, UoM conversions, supplier identity, and status
  flags remain prohibited. Structured brand/category values remain
  authoritative.
- DeepSeek failures use provider-neutral classifications: 401 and other
  configuration/request HTTP failures are permanent; timeout, connection,
  408, 429, and 5xx failures are retryable; empty/malformed output,
  candidate mismatch, correlation mismatch, structured precedence violations,
  and prohibited output fail closed as Stage 2B response errors.
- Only the allowlisted provider-safe request context is sent. Tenant DSNs,
  database credentials, JWT/M2M tokens, API keys, inventory, pricing, tax,
  supplier operational data, barcode ownership, and arbitrary metadata are
  excluded. API keys are environment-only and are never logged, committed, or
  stored in suggestion metadata.
- Tests use fake HTTP transport or `httptest.Server`; no live DeepSeek or
  OpenAI request was made. Verified cases include endpoint/model/auth header,
  JSON mode/instructions, strict parser failures, unknown fields,
  product_type/UoM conversion/prohibited output, structured category/brand
  overrides, correlation mismatch, Pantene-style enrichment, HIKvision,
  Huawei, Epson, fixed-asset immutability, 401/429/500, timeout, connection
  failure, configuration gating, and provider factory selection.

### Verification

- `go test -count=1 ./...` passed in `packages/core`.
- `go test -count=1 ./...` passed in `apps/cloud-server`.
- `go test -count=1 ./...` passed in `apps/sap-agent`.
- Focused core enrichment, OpenAI adapter, DeepSeek adapter, and config tests
  passed uncached.
- `gofmt` completed for all changed Go files and `git diff --check` passed.
- Static diff secret scan found no real API keys, bearer tokens, JWT/M2M
  tokens, DSNs/passwords, SAP credentials, or private keys.
- No schema, SQLC, Atlas, migration, SAP/F1, Stage 2D/E1/E2, review/apply,
  audit, lifecycle, or product schema changes were made.

### Files changed

- `packages/core/enrichment/deepseekadapter/provider.go`
- `packages/core/enrichment/deepseekadapter/provider_test.go`
- `packages/core/config/config.go`
- `packages/core/config/config_test.go`
- `apps/cloud-server/main.go`
- `apps/cloud-server/main_test.go`
- `agents.md`

Core enrichment architecture remains CODE COMPLETE.

Architect-requested DeepSeek provider support: COMPLETE.

Next Action: optional controlled live DeepSeek smoke test using
`DEEPSEEK_API_KEY` configured locally/server-side, followed by deployment and
environment validation. No live request or deployment was performed here.

## DeepSeek Live Stage 2B Contract Retest Preparation — 2026-08-21

- DeepSeek first live connectivity: VERIFIED.
- DeepSeek first live Stage 2B contract: FAILED.
- Root cause: the model returned the noncanonical description member `text`
  instead of the strict Stage 2B member `value`; diagnostics also exposed
  unresolved-brand/missing-category action guidance gaps and a smoke fixture
  that supplied category candidates without constructing an authoritative
  `Snapshot.Category`.
- Stage 2B remains unchanged and fail-closed. No parser alias, response repair,
  or business-rule relaxation was added.
- DeepSeek production guidance now states the exact top-level and nested JSON
  member names, canonical description `value` member, omission rules for
  action-inapplicable optional members, exact candidate restriction for
  MATCH_EXISTING, and conditional actions for resolved/missing brand,
  category, and description. Product type and unsupported semantics remain
  immutable/informational respectively.
- Deterministic DeepSeek tests cover the explicit instruction contract, valid
  canonical output, rejection of `description.text`, invalid KEEP_EXISTING for
  unresolved brand/missing category, candidate restrictions, unsupported
  semantics, and product-type immutability.
- The ignored Pantene smoke fixture now constructs the authoritative category
  in `EnrichmentSourceSnapshot.Category`; its diagnostics therefore derive
  `STRUCTURED_CATEGORY_PRESENT=true` from the same request fields Stage 2B
  validates. Structured brand remains intentionally missing and description
  remains intentionally empty.
- DeepSeek live contract after correction: PENDING RETEST.
- Verification passed: targeted DeepSeek adapter tests, core enrichment tests,
  core `go test -count=1 ./...`, cloud-server `go test -count=1 ./...`,
  sap-agent `go test -count=1 ./...`, gofmt, and `git diff --check`.
- No live request, commit, or push was performed.

## DeepSeek Explicit Non-Thinking Correction — 2026-08-21

- DeepSeek connectivity: VERIFIED.
- DeepSeek JSON-mode connectivity: VERIFIED.
- The first contract issue, `description.text`, was corrected in production
  instructions while Stage 2B remained unchanged.
- The corrected-contract live request reached semantic/type validation and
  exposed a separate `WRONG_VALUE_TYPE` finding; that finding remains open.
- A subsequent default-thinking request exhausted `max_tokens` before final
  content was emitted, with thousands of reasoning characters and
  `finish_reason=length`.
- Product enrichment is bounded structured extraction/classification rather
  than an open-ended reasoning task. The strict Stage 2B parser remains the
  deterministic validation boundary; disabling thinking does not weaken it.
- The DeepSeek adapter now explicitly sends `thinking: {"type":"disabled"}`.
  `max_tokens` remains `2048`; no `reasoning_effort` or unrelated sampling
  parameters were added.
- Stage 2B remains unchanged. The next live non-thinking contract retest is
  pending; DeepSeek live contract verification is not yet claimed.
- No live request, commit, or push was performed.

## DeepSeek Evidence Array Contract Correction — 2026-08-21

- DeepSeek connectivity: VERIFIED.
- DeepSeek thinking-disabled live behavior: VERIFIED.
- The canonical `description.value` correction was verified by the prior live
  response.
- The canonical action rules produced the correct live semantics.
- Canonical candidate matching produced the correct live brand target.
- The remaining live failure was identified exactly as a scalar-vs-array type
  mismatch for `brand.evidence`, `category.evidence`,
  `description.evidence`, and `unsupported_semantics[*].evidence`.
- DeepSeek instructions now require every present `evidence` member to be a
  JSON array of strings, including a one-element array for one fact, and
  explicitly prohibit scalar evidence strings.
- Stage 2B remains unchanged and fail-closed; scalar evidence is still
  rejected, with no adapter-side coercion or response repair.
- `thinking: {"type":"disabled"}`, `max_tokens: 2048`, and
  `response_format: {"type":"json_object"}` remain unchanged.
- Offline adapter tests cover one and multiple evidence strings, permitted
  omission, and scalar rejection for all four evidence-bearing locations.
- Final DeepSeek contract retest: PENDING.
- No live request, commit, or push was performed here.

## DeepSeek D2 — Live Contract Verification

- DeepSeek authentication/connectivity VERIFIED.
- Model `deepseek-v4-flash` live VERIFIED.
- Thinking explicitly disabled.
- `max_tokens=2048`.
- `response_format=json_object`.
- HTTP 200.
- `finish_reason=stop`.
- No `reasoning_content`.
- Valid JSON.
- All canonical Stage 2B top-level object types matched.
- `TYPE_MISMATCH_COUNT=0`.
- Strict canonical brand candidate identity matched.
- The evidence `[]string` contract matched across brand, category, description, and unsupported semantics.
- Stage 2B strict parser PASS.
- No parser aliases, coercion, or repair were required.
- No database, product, or suggestion mutation occurred during the smoke test.
- No secret was persisted.
- `DEEPSEEK_API_KEY` remains environment-only.

### Live verification progression

1. Initial `description.text` contract violation.
2. Corrected to `description.value`.
3. Default thinking caused excessive reasoning and one length exhaustion.
4. Thinking explicitly disabled.
5. Evidence scalar mismatch discovered.
6. Evidence array-of-strings guidance corrected.
7. Final live request passed Stage 2B with zero type mismatches.

DeepSeek live connectivity: VERIFIED

DeepSeek live strict-contract compatibility: VERIFIED

DeepSeek provider implementation: COMPLETE

D1: COMPLETE

D2: COMPLETE

Remaining phase: D3 — deployment/environment validation only.

Core AI Product Enrichment MVP: CODE COMPLETE

Production/environment: NOT YET VERIFIED

Future optimization is optional and is not a blocker.

The ignored scratch harness may remain temporarily for deployment diagnostics;
it is not production code.

No new DeepSeek/OpenAI call, database connection, database alteration,
deployment, commit, or push was performed for this record. Only this worklog
file was updated.

## D3 - Deployment / Environment Validation

### D3 scope and result

- Review date: 2026-08-21.
- D3 was limited to deployment/configuration validation, tenant migration and
  RBAC readiness, runtime wiring, secret handling, isolation review, and
  operational runbook evidence. No feature implementation, migration
  application, live database mutation, provider call, commit, or push was
  performed.
- The worktree was clean at the start and remains limited to this worklog
  update. HEAD was `abc94f1` on `feature-agent` tracking `origin/feature-agent`.
- Core MVP source and test status remains CODE COMPLETE. The deployment gate is
  blocked by pre-existing operational secret-output code described below; the
  exact safe migrator command is documented and does not use that wrapper.

### Final provider configuration

Runtime names are sourced from `packages/core/config/config.go`:

- `ENRICHMENT_ENABLED` defaults to `false`.
- `ENRICHMENT_PROVIDER` selects `openai` or `deepseek`; unknown values fail
  configuration validation and provider construction safely.
- DeepSeek requires `DEEPSEEK_API_KEY` and `DEEPSEEK_MODEL` only when
  enrichment is enabled with `ENRICHMENT_PROVIDER=deepseek`.
- `DEEPSEEK_BASE_URL` defaults to `https://api.deepseek.com`. The adapter
  accepts only an absolute HTTP(S) URL and rejects embedded credentials, query
  parameters, and fragments. Production configuration should use HTTPS and a
  deliberately approved endpoint.
- The default DeepSeek model is `deepseek-v4-flash`.
- DeepSeek sends `thinking: {"type":"disabled"}`, `max_tokens: 2048`,
  `response_format: {"type":"json_object"}`, and `stream: false`.
- OpenAI remains selectable with `OPENAI_API_KEY` and
  `OPENAI_ENRICHMENT_MODEL`; the OpenAI key is not required for a DeepSeek
  deployment.
- Shared runtime controls use the existing names
  `OPENAI_ENRICHMENT_TIMEOUT`, `OPENAI_ENRICHMENT_WORKER_INTERVAL`,
  `OPENAI_ENRICHMENT_BATCH_SIZE`, and `OPENAI_ENRICHMENT_MAX_RETRIES`.

Sanitized DeepSeek deployment example:

```env
ENRICHMENT_ENABLED=true
ENRICHMENT_PROVIDER=deepseek
DEEPSEEK_API_KEY=<secret-manager-or-environment>
DEEPSEEK_BASE_URL=https://api.deepseek.com
DEEPSEEK_MODEL=deepseek-v4-flash
```

The current process environment reported
`DEEPSEEK_API_KEY_PRESENT=false`, `OPENAI_API_KEY_PRESENT=false`, and
`MASTER_DB_URL_PRESENT=false`; values were never printed.

### DeepSeek D2 evidence

The existing D2 evidence is accepted and was not retested in D3: authentication
and connectivity succeeded; model `deepseek-v4-flash` was accepted; thinking
was disabled; `max_tokens=2048`; JSON mode was enabled; HTTP was 200;
`finish_reason=stop`; `reasoning_content` was absent; JSON was valid;
`TYPE_MISMATCH_COUNT=0`; strict canonical candidate matching succeeded; the
`evidence []string` contract succeeded; and strict Stage 2B parsing passed.
No additional external API call was made in D3.

### Database ownership invariant

- The master database is used for tenant registry/control-plane identity,
  tenant connection metadata, active-tenant enumeration, and migration
  coordination only.
- Each selected tenant database owns organizations, users/roles/permissions,
  SAP staging/business rows, products, brands, categories,
  `product_enrichment_suggestions`, review lifecycle, application lifecycle,
  and `audit_logs`.
- For active tenant business work:
  `Product DB = Suggestion DB = Worker DB = Review DB = Application DB = Audit DB`.
- F1, F2, Stage 2D, and Stage 2E construct request/worker repositories from the
  selected tenant pool. No active path uses a master product or suggestion
  fallback. Legacy master business rows are quarantined/inert for enrichment.

### Tenant migration chain and readiness

The exact checked-in chain is present and Atlas integrity is valid:

- `20260820000000.sql` - product enrichment suggestion foundation.
- `20260820010000.sql` - retry durability and worker attempt fields.
- `20260821000000.sql` - tenant-local `product_enrichment:review` permission
  metadata, with no role assignment.
- `20260821010000.sql` - tenant-local `product_enrichment:apply` permission
  metadata, with no role assignment.

`atlas migrate validate --dir "file://db/migrations"` passed.
`atlas migrate hash --dir "file://db/migrations"` passed with no worktree
change, and `atlas.sum` contains all four migration entries. SQLC v1.30.0
`sqlc compile -f sqlc.yaml` and `sqlc generate` both passed; generation caused
no repository diff.

The tenant migrator is `apps/cloud-server/cmd/migrate-tenants/main.go`. It:

- requires `MASTER_DB_URL` and reads active tenants from the master registry;
- uses `WHERE is_active = true` and each tenant's `db_conn_str` independently;
- with `-master=false`, skips master migration but still reads the master
  registry and processes tenant databases;
- runs Atlas against each tenant DSN, never substitutes the master DSN for a
  failed tenant;
- logs per-tenant failures, continues to the remaining tenants, and exits 1 if
  any tenant failed; and
- has no dependency on or invocation from the enrichment worker.

Safe read-only status command for one deployment revision:

```powershell
cd apps/cloud-server
go run cmd/migrate-tenants/main.go -master=false -status=true
```

Authorized tenant migration command, not executed in D3:

```powershell
cd apps/cloud-server
go run cmd/migrate-tenants/main.go -master=false
```

The legacy `apps/cloud-server/migrate-tenants.ps1` wrapper is not the safe
deployment command because it prints the full `MASTER_DB_URL`; it must not be
used until corrected outside D3.

### New-tenant operational limitation

`CreateTenant` registers tenant name, slug, connection metadata, active state,
and settings in the master registry. It does not create/provision the tenant
database, run migrations, provision RBAC, or validate enrichment readiness.
Current activation can therefore precede operational migration; this is an
explicit deployment limitation, not silently treated as readiness.

Required new-tenant sequence:

1. Provision/register the tenant database.
2. Run tenant migrations.
3. Validate migration state.
4. Provision explicit RBAC role mappings.
5. Validate tenant connectivity.
6. Activate/use business features.

### RBAC deployment readiness

- `20260821000000.sql` provisions exactly `product_enrichment:review`.
- `20260821010000.sql` provisions exactly `product_enrichment:apply`.
- Both migrations insert capability metadata only; neither assigns a role or
  user automatically.
- Review and apply checks are separate tenant-local permission lookups through
  `role_permissions` and `user_roles`.
- Review does not imply apply; apply does not require review; `products:manage`
  is not used as a substitute; there is no admin-name or role-name bypass in
  the enrichment handlers.
- Deployment must explicitly assign the two permissions in each intended
  tenant. No live permission rows were queried:
  `LIVE_PERMISSION_ROWS=UNVERIFIED`.

### JWT and SAP M2M deployment requirements

- Standard user JWTs contain signed `tenant_slug`.
- `TenantBindingMiddleware` requires the `x-tenant-id` header to equal that
  signed claim; the header is only a consistency assertion.
- JWTs without `tenant_slug` fail closed on protected user routes. Users must
  re-login after the tenant-bound JWT rollout.
- No localStorage/header tenant authority is reintroduced.
- SAP machine JWTs require `tenant_slug`, positive `organization_id`,
  `token_type=machine`, and `is_m2m=true`.
- The registered M2M client organization must equal the signed organization;
  `x-tenant-id` and `x-organization-id` are consistency assertions.
- There is no organization default of 1 in the SAP auth path, and payload
  organization data cannot override the trusted claim.
- Legacy unbound M2M credentials are rejected by the migration-specific
  middleware. Reissue one tenant/org-bound SAP M2M credential per integration;
  no M2M secret was displayed.

### F1 runtime trace

`SAP Agent -> tenant/org-bound M2M JWT -> signed tenant_slug and
organization_id -> master tenant registry lookup -> selected tenant pool ->
tenant-local organization validation -> tenant-local SAPMigrationUseCase ->
transaction -> commit -> tenant-local Stage 2A coordinator -> pending
suggestion`.

The payload organization is checked against the trusted claim. The post-commit
enqueue is best effort and cannot roll back a committed SAP transaction. No
master product or suggestion write is present in this path.

### F2 / DeepSeek runtime trace

`master active-tenant enumeration -> selected tenant pool -> fresh tenant
repository/store -> due suggestion claim -> ProductEnrichmentProvider ->
configured OpenAI or DeepSeek adapter -> strict Stage 2B -> tenant-local
lifecycle update`.

The supervisor re-enumerates active tenants every cycle, skips inactive tenants,
rediscovers newly active tenants, creates a fresh worker per tenant, and
isolates setup/worker failures. The provider is stateless/shared and contains
no tenant credentials or tenant identity. The worker never applies product
mutations. DeepSeek uses disabled thinking, JSON mode, and the strict parser;
retryable/permanent/provider-contract failures are classified without logging
credentials.

### Review and apply runtime traces

Review:

`JWT -> TenantBindingMiddleware -> TenantMiddleware -> RepoKey -> tenant-local
user -> server-derived organization -> product_enrichment:review -> tenant-local
suggestion/current product -> stale validation -> approve/reject -> tenant-local
audit`.

Review never auto-applies a suggestion.

Apply:

`POST /api/product-enrichment/suggestions/:id/apply -> JWT -> tenant binding ->
tenant RepoKey -> tenant-local user -> server-derived organization ->
product_enrichment:apply -> E1 ApplyApprovedSuggestion -> tenant transaction ->
suggestion lock -> product lock -> fingerprint/stale check -> precedence and
canonical-target validation -> narrow update -> applied transition -> audit ->
commit`.

Tenant, organization, product, applier, brand, category, description, force,
and override are not client-controlled. Apply locks and revalidates the
tenant-local suggestion and product and does not auto-create taxonomy records.

### Final product write surface and route inventory

Stage 2E product SQL writes only `products.brand_id`, `products.category_id`,
`products.description`, and `products.updated_at`, plus suggestion lifecycle
fields and tenant-local `audit_logs`. It does not write SKU, name, product type,
metadata, inventory, prices, tax, UoM, barcodes, suppliers,
`track_inventory`, or active/sellable/purchasable flags.

Authenticated enrichment routes are:

- `GET /api/product-enrichment/suggestions`
- `GET /api/product-enrichment/suggestions/:id`
- `POST /api/product-enrichment/suggestions/:id/approve`
- `POST /api/product-enrichment/suggestions/:id/reject`
- `POST /api/product-enrichment/suggestions/:id/apply`

No public review, public apply, bulk apply, auto apply, force apply, or GET
apply route exists.

### Deployment order

1. Build and run the final repository tests/checks.
2. Provision secrets through the deployment secret manager.
3. Deploy tenant-bound JWT behavior.
4. Require user re-login.
5. Reissue tenant/org-bound SAP M2M credentials.
6. Deploy F1 tenant-local SAP routing.
7. Run approved migrations against every intended tenant with
   `-master=false`.
8. Validate the four enrichment/RBAC migration states per tenant.
9. Explicitly assign `product_enrichment:review`.
10. Explicitly assign `product_enrichment:apply`.
11. Deploy F2 supervisor and configured provider selection.
12. Enable Stage 2D review routes.
13. Enable Stage 2E apply routes.
14. Set `ENRICHMENT_ENABLED=true` intentionally for the approved scope.
15. Monitor the first tenant processing cycle and audit events.

### First-tenant staging validation runbook

This is a plan only; D3 did not execute business mutation. Use one safe test
product in one staging tenant and separate reviewer-only and applier test users.

1. Before processing, record tenant slug, organization, product ID, source
   item code/name, `product_type`, brand/category/description, SKU, inventory,
   pricing, tax, UoM/conversions, barcodes, suppliers, status flags, and the
   product fingerprint.
2. Confirm the expected enrichment gap is brand, description, and/or missing
   category; confirm the product is SAP-sourced with an approved product type.
3. Run the approved staging SAP batch and verify the suggestion is created in
   the selected tenant DB with `pending`, the expected organization/product,
   source fingerprint, and no master business write.
4. With enrichment enabled and provider set to DeepSeek, observe
   `pending -> processing -> in_review`; verify provider, model metadata, and
   no protected fields in the suggestion proposal.
5. Use the reviewer-only account to list/detail and approve. Verify the
   reviewer has `product_enrichment:review` and cannot apply.
6. Use the applier account with `product_enrichment:apply` to apply the
   approved suggestion. Verify `applied`, `applied_at`, changed fields, and a
   tenant-local audit record.
7. After processing, compare the recorded snapshot. Only brand/category/
   description and allowed lifecycle/audit fields may differ; product type,
   SKU, inventory, pricing, tax, UoM, conversions, barcodes, suppliers, and
   status flags must be unchanged.
8. Verify a different tenant has no new suggestion, status change, product
   mutation, or audit row from this run.

### Colliding-tenant isolation validation

Use two physical tenant databases where Tenant A and Tenant B can contain the
same organization/product/suggestion numeric IDs (for example org 1/product
95/suggestion 1), or equivalent intentionally colliding staging fixtures.
Authenticate each request with its matching signed tenant claim and header.
Verify list/detail/review/apply requests against A never read B and vice versa;
an ID that exists only in the other tenant returns the tenant-local not-found
behavior. Verify the supervisor creates separate pools/repositories and that
workers process each tenant's same numeric suggestion independently. Confirm
audits and product changes remain in the physical tenant that owns the request.

### Observability

Current MVP logs expose tenant slug, suggestion ID, normalized error class, and
worker/supervisor failure boundaries. Provider failures distinguish retryable,
permanent, timeout, network, and strict response failures; status transitions
are durable in tenant suggestions; review/apply errors are sanitized and audit
events are durable. Provider keys, JWTs, M2M tokens, DSNs, and response bodies
are not intentionally logged by the enrichment runtime. Metrics/alerts and a
centralized audit/observability service remain optional follow-up work.

### Go, SQLC, Atlas, and diff verification

- `C:\Program Files\Go\bin\go.exe version`: Go 1.26.7.
- `packages/core`: `go test -count=1 ./...` passed.
- `apps/cloud-server`: `go test -count=1 ./...` passed.
- `apps/sap-agent`: `go test -count=1 ./...` passed.
- `C:\Program Files\Go\bin\gofmt.exe -l` over the D3-relevant human-authored
  Go files reported no files.
- `git diff --check` passed before the D3 worklog update and is required again
  after it.
- SQLC v1.30.0 compile and generation passed with no generated diff.
- Atlas validation and hash passed with no migration-history change.
- No live database was contacted.

### Secret scan

- Current tracked source/diff contains no production-looking DeepSeek API key,
  OpenAI API key, `sk-` key, JWT signing secret, SAP credential, M2M bearer
  token, private key, or tenant secret from this D3 work.
- Example M2M/JWT/DSN values are placeholders or local-development fixtures;
  no value was printed in this validation.
- Existing tracked local/dev DSN material was detected in development/example
  files, including `apps/pos-client/.env` and `apps/pos-client/.env.dev`;
  production secrets must not be placed there.
- IMPORTANT operational security finding: `apps/cloud-server/migrate-tenants.ps1:28`
  prints the full `MASTER_DB_URL`, and
  `apps/cloud-server/cmd/check-tenant/main.go:90,115-122` uses insufficient
  length-based masking for a DSN. These files were not changed in D3. Do not
  use the wrapper/diagnostic output in deployment logs until corrected.

### Live environment status

`LIVE_ENVIRONMENT_READ_ONLY_CHECKS=NOT_PERFORMED`.
The environment had no `MASTER_DB_URL`, no provider key, and no legitimate
staging database connection available to this run. Therefore tenant registry,
tenant migration versions, live permission rows, table readiness, and actual
SAP/worker/review/apply execution remain unverified. D2's previously recorded
DeepSeek live contract evidence is not equivalent to target deployment
verification.

### D3 blocker classification

CRITICAL

- No active enrichment-runtime cross-tenant or product-write CRITICAL defect
  was found in this static gate.

IMPORTANT

- Correct or retire the legacy migration wrapper's full-DSN output and the
  check-tenant DSN masking before using either operational tool in deployment
  or shared logs. This is a code/security correction intentionally not made in
  D3.
- Provision DeepSeek secret/configuration in the target environment, or select
  OpenAI with its own key; neither is present locally.
- Apply the four tenant migrations and validate each intended tenant.
- Assign review/apply permissions explicitly per tenant.
- Reissue SAP M2M credentials and require user re-login for the JWT rollout.
- Deploy the intended revision and execute the first-tenant staging runbook.

MINOR

- Remove or quarantine tracked local/dev DSN fixtures and replace weak
  diagnostics with structured DSN redaction during normal security hardening.
- Add metrics/alerts for supervisor cycles and provider status rates.

### Status matrix and final distinction

`CORE_MVP_CODE_COMPLETE=true`

`DEPLOYMENT_READY=false` (the operational secret-output correction is still
required; the safe Go migrator path is documented).

`DEPLOYMENT_ENVIRONMENT_VERIFIED=false`

`PRODUCTION_END_TO_END_VERIFIED=false`

`agents.md` was updated by this D3 record only. No other file was intentionally
changed, no live migration or business mutation was performed, and no commit or
push was made.

Final verdict: `D3_BLOCKED_CODE_CORRECTION_REQUIRED`

### D3 Deployment Utility Secret Hardening

- Review date: 2026-08-21.
- Original `apps/cloud-server/migrate-tenants.ps1` output interpolated and
  printed the raw `MASTER_DB_URL` value.
- Original `apps/cloud-server/cmd/check-tenant/main.go` used length-based
  truncation, which could preserve database credentials and query material.
- The PowerShell wrapper now prints only `MASTER_DB_URL configured.` and still
  reads the variable for migration behavior.
- `check-tenant` now uses proper PostgreSQL URL parsing and reconstructs only
  validated host/port metadata plus `credentials=<redacted>`. Userinfo,
  passwords, database paths, and all query parameters are omitted.
- Empty, malformed, unsupported, keyword/value, invalid-host, and invalid-port
  inputs return `<redacted>`; sanitization never falls back to the original
  DSN.
- Connection-construction errors in `check-tenant` no longer include raw error
  text that could echo a DSN. Existing tenant discovery, status reporting,
  connection attempts, and migration behavior were otherwise unchanged.
- Focused `check-tenant` tests cover URL credentials, escaped passwords,
  query options including credential-like parameters, malformed/empty input,
  and keyword/value input. The focused package test passed.
- `packages/core`, `apps/cloud-server`, and `apps/sap-agent` each passed
  `go test -count=1 ./...` using Go from `C:\Program Files\Go\bin`.
- `gofmt` was run on the modified Go files and `git diff --check` passed.
- Static PowerShell review confirms `MASTER_DB_URL` is read but never emitted.
- No live database was connected, no migration was run, no deployment was
  performed, no secret value was displayed, and no commit or push was made.
- Affected files: `apps/cloud-server/migrate-tenants.ps1`,
  `apps/cloud-server/cmd/check-tenant/main.go`,
  `apps/cloud-server/cmd/check-tenant/main_test.go`, and `agents.md`.
- D3 code blocker: `RESOLVED`.
- Core MVP: `CODE COMPLETE`.
- Deployment readiness: `READY FOR D3 REVALIDATION`.
- `DEPLOYMENT_ENVIRONMENT_VERIFIED` and
  `PRODUCTION_END_TO_END_VERIFIED` remain unclaimed because they require
  actual environment execution.

## D3 Revalidation After Deployment Utility Hardening

- Revalidation date: 2026-08-21.
- This was a revalidation-only gate. No feature, frontend, enrichment logic,
  deployment, live migration, tenant database mutation, provider API call,
  commit, or push was performed.

### Original D3 blockers

- `MIGRATE_TENANTS_DSN_EXPOSURE=RESOLVED`: `apps/cloud-server/migrate-tenants.ps1`
  reports only that `MASTER_DB_URL` is configured and never interpolates the
  value into output.
- `CHECK_TENANT_DSN_MASKING=RESOLVED`:
  `apps/cloud-server/cmd/check-tenant/main.go` reconstructs only validated
  host/port metadata and `credentials=<redacted>`. Userinfo, passwords,
  database paths, query parameters, malformed input, unsupported DSNs,
  invalid hosts/ports, and connection-construction errors fail closed without
  emitting the original DSN.
- Focused sanitizer tests and the full cloud-server suite passed. No current
  utility path was found that emits a raw credential-bearing DSN.

### Secret and logging policy

- No production-looking DeepSeek/OpenAI key, `sk-` credential, database
  password, full DSN, JWT signing secret, M2M token, SAP credential, or private
  key was found in the current tracked source/diff review.
- Clearly fake test credentials and tracked local/development DSN fixtures
  remain separate development material; no secret value was printed.
- Runtime logging remains sanitized for DSNs, provider credentials, tokens,
  and provider response bodies.

### Final code readiness

- Core architecture and tests for tenant-bound JWT, tenant/org-bound SAP M2M,
  tenant-local SAP migration, Stage 2A enqueue/fingerprint, strict Stage 2B,
  F2 worker, Stage 2D review, Stage 2E deterministic application, separate
  review/apply RBAC, and DeepSeek provider selection remain complete.
- No active source/code blocker was found in this revalidation.
- DeepSeek D2 evidence remains accepted and was not retested or called again:
  HTTP 200, `finish_reason=stop`, no `reasoning_content`, valid JSON,
  `TYPE_MISMATCH_COUNT=0`, and `STAGE2B_RESULT=PASS`.
- Current DeepSeek configuration names are `ENRICHMENT_ENABLED`,
  `ENRICHMENT_PROVIDER`, `DEEPSEEK_API_KEY`, `DEEPSEEK_BASE_URL`, and
  `DEEPSEEK_MODEL`; the adapter sends disabled thinking, `max_tokens=2048`,
  JSON-object response format, and no hardcoded key. OpenAI remains selectable
  and unknown providers fail safely.

### Data ownership and runtime trace

- MASTER is limited to tenant registry/control-plane identity, connection
  metadata, active-tenant discovery, and migration coordination.
- Each TENANT DB owns organizations, users/RBAC, SAP business data, products,
  brands/categories, `product_enrichment_suggestions`, review/application
  lifecycle, and `audit_logs`.
- For active tenant business work:
  `Product DB = Suggestion DB = Worker DB = Review DB = Application DB = Audit DB`.
  No active master business fallback is used.
- SAP trace: SAP Agent -> trusted tenant/org M2M -> tenant pool -> tenant-local
  transaction -> commit -> Stage 2A suggestion.
- Worker trace: master active-tenant discovery -> tenant repository -> due
  suggestion -> configured provider/DeepSeek -> strict Stage 2B -> tenant-local
  `in_review`/`retryable`/`failed`.
- Review and apply derive tenant, organization, product, reviewer/applier,
  and canonical targets server-side. Review never auto-applies.

### Migration readiness

- The required chain is present: `20260820000000.sql`, `20260820010000.sql`,
  `20260821000000.sql`, and `20260821010000.sql`.
- Atlas validation and hash passed without applying a migration or changing
  migration history. SQLC compile and generate passed with v1.30.0 and no
  generated repository diff.
- `-master=false` skips master migration while still reading the master
  tenant registry and migrating tenant databases independently.
- Safe read-only status command from the current source:
  `cd apps/cloud-server; go run cmd/migrate-tenants/main.go -master=false -status=true`.
- The authorized tenant migration command remains operational-only and was not
  run in D3:
  `cd apps/cloud-server; go run cmd/migrate-tenants/main.go -master=false`.

### RBAC, JWT, and M2M readiness

- `20260821000000.sql` provisions `product_enrichment:review` metadata only;
  `20260821010000.sql` provisions `product_enrichment:apply` metadata only.
- No automatic production role grant exists. Review does not imply apply, and
  apply does not require review. Without a live tenant DB:
  `LIVE_PERMISSION_ROWS=UNVERIFIED`.
- Interactive JWTs require signed `tenant_slug`; old tokens without that claim
  fail closed on protected routes, so users must re-login after rollout.
- SAP M2M JWTs require `tenant_slug`, positive `organization_id`,
  `token_type=machine`, and `is_m2m=true`. Credentials must be reissued per
  tenant/org; there is no organization default of 1.

### Product writes and route security

- Stage 2E product writes remain limited to `brand_id`, `category_id`,
  `description`, and `updated_at`, plus suggestion/audit lifecycle writes.
  There is no reachable enrichment write for SKU, name, `product_type`,
  metadata, inventory, prices, tax, UoM, barcodes, suppliers,
  `track_inventory`, or active/sellable/purchasable flags.
- The authenticated routes are the list/detail GET routes and POST
  approve/reject/apply routes under `/api/product-enrichment/suggestions`.
  No public alias, bulk apply, auto apply, force apply, or GET apply route
  exists.

### New-tenant and live-environment status

- `CreateTenant` still registers control-plane metadata only; it does not
  provision, migrate, seed, validate, or assign RBAC in a new tenant DB.
- Required operational order remains: provision/register DB -> run tenant
  migrations -> verify state -> assign RBAC -> validate connectivity ->
  activate/use business features.
- No legitimate target environment credentials/configuration were available;
  `LIVE_ENVIRONMENT_READ_ONLY_CHECKS=NOT_PERFORMED`. No live database was
  contacted, no migration was applied, and no production E2E path was run.

### Verification gate

- `C:\Program Files\Go\bin\go.exe test -count=1 ./...` passed in
  `packages/core`, `apps/cloud-server`, and `apps/sap-agent`.
- `C:\Program Files\Go\bin\gofmt.exe -l` reported no modified
  human-authored Go files.
- `git diff --check` passed before this worklog update; final status/stat
  checks are required after it.

### D3 decision

D3:
COMPLETE

Core AI Product Enrichment MVP:
CODE COMPLETE

Deployment readiness:
READY

Deployment environment:
NOT VERIFIED

Production end-to-end:
NOT VERIFIED

CORE_MVP_CODE_COMPLETE=true
DEPLOYMENT_READY=true
DEPLOYMENT_ENVIRONMENT_VERIFIED=false
PRODUCTION_END_TO_END_VERIFIED=false

Remaining operational requirements are target secret provisioning, tenant
migrations, live migration-state validation, explicit per-tenant review/apply
RBAC assignment, tenant/org-bound SAP M2M reissue, user re-login, staging
execution, and deployment monitoring. These do not constitute current source
code blockers.

Next Action:
Inspect architect-referenced localhost Go pages for review/approval/apply
frontend integration.

## UI0.5 — Enrichment Read Authorization for Apply Workflow

- UI0 discovered that apply-only users could call the apply endpoint only if
  they already knew a suggestion ID, because list/detail reads required
  `product_enrichment:review`.
- Corrected read authorization so `product_enrichment:review` preserves the
  existing review read scope, while `product_enrichment:apply` without review
  can list and view only `approved` and `applied` suggestions.
- Reviewer list requests without `status` still default to `in_review`.
- Apply-only list requests without `status` default to `approved`.
- Apply-only detail requests for non-approved/non-applied suggestions return a
  scoped not-found response.
- Review/apply write permissions are unchanged: approve/reject require
  `product_enrichment:review`; apply requires `product_enrichment:apply`.
- JWT, tenant-local user loading, tenant repository selection, server-derived
  organization scoping, and no-master-fallback behavior are unchanged.
- Focused handler and routing tests cover reviewer/apply-only/both/neither
  read scopes, defaults, approved/applied detail access, scoped denial,
  write-permission separation, and cross-organization isolation.
- Swagger generation was not changed; documentation remains a future UI3
  follow-up.

Next Action:
UI1 — Angular suggestions list/detail implementation in
`apps/pos-client/frontend`.
