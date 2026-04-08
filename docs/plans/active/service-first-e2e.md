# Service-First E2E Plan

## Goal

Reframe LLM-Wiki around one core service and drive the next work by one rule:

First make the full end-to-end user flow work for a real hosted user, then iterate on depth, polish, and broader integrations.

## Product Direction

- LLM-Wiki service is the core product.
- CLI, HTTP, MCP, hosted install docs, and a future protocol are access methods.
- LLM-Wiki keeps its own canonical backend storage model.
- External tools such as Obsidian, Feishu, Notion, or IDE plugins are downstream adapters or sync targets.

## Primary E2E Story

The first success path should be:

1. Admin configures required OAuth provider credentials.
2. A new user signs in with Google or GitHub.
3. The service creates the user automatically if needed.
4. The service creates a personal default tenant automatically if needed.
5. The user lands in a usable workspace without admin hand-holding.
6. `llm-wiki auth login` opens a browser, listens on a localhost callback port, and persists `~/.llm-wiki/` config automatically.
7. The CLI can immediately call the same service without repeatedly passing manual tenant or URL flags.
8. The same authenticated user can create another tenant or workspace and invite collaborators.

If that path is not smooth, feature breadth should not take priority over new surfaces.

## Phase Breakdown

### Phase 1: Make the first-run hosted flow real

- keep first-boot admin setup
- add admin-managed OAuth provider config for Google and GitHub
- support first-login auto user creation
- support first-login auto personal-tenant creation
- define default tenant naming rules from username or email
- define membership and role defaults for the creator
- make browser login the default CLI path with localhost callback
- keep device code as explicit fallback for headless environments
- persist CLI base URL, tenant selection, and tokens in `~/.llm-wiki/`

### Phase 2: Make the workspace model real

- allow one user to create multiple tenants or workspaces
- add invite and membership flows
- define tenant roles and membership lifecycle
- remove assumptions that only admins create tenants
- make tenant selection in UI and CLI follow authenticated membership grants

### Phase 3: Stabilize the public contract

- normalize service semantics across HTTP, CLI, MCP, and install skill guidance
- define a protocol shape so third parties can implement clients
- make tenant and namespace context derived from authenticated identity and explicit selection, not arbitrary headers
- document canonical auth, tenancy, and revision flows once

### Phase 4: Add downstream adapters

- build Obsidian as the first adapter target
- define export, sync, or mirrored editing semantics
- add additional targets such as Feishu, Notion, and IDE plugins after the adapter contract is stable

## Near-Term Execution Order

1. tighten the auth and tenant model around OAuth-backed human identity
2. make CLI browser login and local profile persistence frictionless
3. make personal-tenant auto provisioning deterministic
4. add multi-tenant membership and invites
5. extract and document the stable service contract
6. build the first Obsidian adapter or plugin

## Non-Goals For This Round

- polishing every admin page before the main hosted flow works
- building a full first-party authorization server if external OAuth client mode is enough
- broad downstream sync support before the first adapter contract is validated
- over-optimizing MCP or protocol breadth before login, tenancy, and write paths are stable
