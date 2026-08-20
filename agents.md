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
