# Auth And CLI Access Plan

## Goal

Add a real identity and authorization model for LLM-Wiki that works across:

- local human CLI usage
- local agent or MCP usage
- browser-based OAuth sign-in
- remote and headless environments such as CI, cloud VMs, and Kubernetes
- service-to-service integrations that need `ns`-scoped fine-grained tokens
- multi-`ns` collaboration where normal users can self-serve into a personal `ns` and then create more

## Why Now

The current model is still bootstrap-grade:

- CLI is only a thin HTTP wrapper with `base-url` and `ns`
- HTTP and MCP integrations still rely on `X-LLM-Wiki-NS`
- there is no first-class user, session, or service principal identity
- there is no central token issuance or revocation model
- `ns` creation still reads too much like an admin-only provisioning concern
- CLI login is not yet optimized around the browser loopback callback path

That is enough for local development, but not enough for hosted multi-`ns` use.

## Design Principles

- Keep `cmd/cli` thin. Credential discovery and login UX can live in the CLI, but authorization decisions must stay server-side.
- Treat `ns` selection as a consequence of authenticated identity and granted access, not as a free-form client header.
- Treat a personal default `ns` as part of user onboarding, not as a separate manual provisioning step.
- Separate human sessions from service tokens.
- Support interactive and headless login flows from the start.
- Prefer browser login with localhost callback for interactive CLI sessions; only fall back to device flow when needed.
- Make all access paths converge on the same server-side auth context model.
- Prefer short-lived access tokens plus refresh or re-issuance paths over long-lived bearer secrets.
- Every token must be attributable to a principal and auditable.

## Target Principal Model

Every authenticated request should resolve into one principal record in server context.

Core fields:

- `principal_type`: `user`, `service`, `agent_session`, or `admin`
- `principal_id`
- `ns`
- `space_ids`
- `team_ids`
- `scopes`
- `session_id` or `token_id`
- `actor_type`: immediate caller type used for audit
- `actor_id`

Audit and revision metadata such as `author_type` and `author_id` should default from that auth context instead of being trusted from raw user input alone.

## Access Surfaces

### 1. CLI explicit token

Support:

- `lw --token <token> ...`
- `LLM_WIKI_TOKEN=... lw ...`

Use case:

- CI
- quick local scripts
- debugging
- temporary delegated access

### 2. CLI config file under `~/.llm-wiki/`

Support profile-based config discovery such as:

- `~/.llm-wiki/config.yaml`
- `~/.llm-wiki/profiles/default.json`

Suggested shape:

```yaml
current_profile: default
profiles:
  default:
    base_url: https://llm-wiki.example.com
    auth:
      mode: oauth
      access_token: ""
      refresh_token: ""
      expires_at: ""
    ns: tenant_a
  ci:
    base_url: https://llm-wiki.example.com
    auth:
      mode: token
      token_file: /var/run/secrets/llm-wiki/token
    ns: tenant_a
```

Recommended credential precedence:

1. explicit CLI flags
2. environment variables
3. active profile in `~/.llm-wiki/`
4. built-in development defaults

This keeps local and cloud usage predictable.

CLI login should update this profile automatically after successful auth, including:

- `base_url`
- selected or default `ns`
- access token metadata
- refresh token metadata when present
- optional user display information for profile introspection

### 3. Browser OAuth 2.0 login

Add `lw login` with an experience similar to Codex:

- `Sign in with Browser`
- `Sign in with Device Code`
- `Provide access token`

Recommended browser flow:

- Authorization Code + PKCE
- loopback redirect such as `http://127.0.0.1:<random>/auth/callback`
- fallback to device authorization flow for headless terminals

CLI behavior:

- open browser automatically when possible
- bind a localhost listener for the callback instead of asking the user to paste codes in the common path
- persist tokens in `~/.llm-wiki/`
- store server metadata and the chosen `ns` profile
- refresh access tokens automatically when refresh token is present

Provider rollout for the first hosted flow:

- admin configures required Google and GitHub OAuth client credentials
- login screen shows enabled providers automatically
- first successful OAuth login can create the internal user if policy allows
- first successful OAuth login can also create the user's personal default `ns` if it does not exist yet

### 4. Service tokens for cloud and K8s

Support server-issued fine-grained tokens for:

- batch jobs
- MCP bridges
- in-cluster agents
- folder-specific or `ns`-specific automation
- external systems with bounded permissions

These should not reuse human refresh tokens.

## Recommended Token Types

### Human session tokens

- issued after browser OAuth or device flow
- short-lived access token
- refresh token stored by the CLI
- principal type is `user`

### Personal access tokens

- optional, for advanced users who want non-OAuth CLI use
- manually created and named
- long-ish lifetime but revocable
- narrower than full session scopes by default

### Service access tokens

- issued to a service principal
- can be static with rotation or minted dynamically by a broker
- principal type is `service`
- never imply wildcard `ns` access unless explicitly granted

### Delegated agent session tokens

- short-lived tokens minted by LLM-Wiki for a human-approved automation run
- principal type is `agent_session`
- must carry the initiating human or service principal in audit metadata

## Scope Model

Start with coarse scopes, then add resource constraints.

Base scopes:

- `ns.read`
- `namens.read`
- `folders.write`
- `documents.read`
- `documents.write`
- `documents.archive`
- `revisions.read`
- `mcp.invoke`
- `tokens.issue`
- `tokens.revoke`
- `admin.ns`

Resource restrictions:

- `ns`-bound
- optionally restricted to selected folders
- optionally restricted to selected API folder resources
- optional read-only or write-only mode

Token examples:

- human CLI token for one `ns` with full wiki write access
- service token only allowed to read `org/*` and write `drafts/*`
- MCP token that can invoke tools but cannot archive documents

## Server-Side Auth Architecture

Add a dedicated auth subsystem with these responsibilities:

1. Resolve bearer token or session credential from HTTP and MCP requests.
2. Validate signature or introspect opaque token.
3. Materialize a normalized auth context.
4. Enforce `ns`, folder, and scope checks before handler logic runs.
5. Inject caller identity into audit and revision creation paths.

Suggested internal packages:

- `internal/auth`: auth context types, middleware, validators
- `internal/authz`: policy evaluation and scope checks
- `internal/identity`: principal, membership, service account, and token models

Suggested HTTP changes:

- `Authorization: Bearer <token>` becomes the primary auth input
- `X-LLM-Wiki-NS` becomes optional legacy compatibility only
- if both are present, the authenticated token grant wins and mismatches are rejected

## OAuth Deployment Model

There are two realistic deployment paths.

### Path A: LLM-Wiki as OAuth client to an external IdP

Use when users already have an external identity system:

- Auth0
- Okta
- Keycloak
- Google Workspace backed identity
- GitHub-backed developer identity

LLM-Wiki responsibilities:

- redirect user to IdP
- exchange code for tokens
- map upstream identity to an internal principal
- evaluate `ns` memberships locally
- auto-create the user on first login when policy allows
- auto-create the personal default `ns` on first login when policy allows

### Path B: LLM-Wiki as its own authorization server for first-party CLI

Use when wanting a Codex-like first-party sign-in experience.

LLM-Wiki responsibilities:

- login UI
- authorization endpoints
- device flow endpoints
- token issuance and refresh
- service token issuance UI or API

Recommended rollout:

- v1: support external IdP integration plus first-party token issuance
- v2: add fully first-party OAuth authorization server only if product needs it

This avoids overbuilding the identity stack too early.

## Workspace And Tenant Onboarding

The target hosted behavior should be closer to an `ns`-scoped collaboration product:

- a user can sign in without an admin pre-creating an `ns` for them
- first login creates a personal default `ns` named from username or email
- the creator becomes owner or admin of that `ns`
- the same user can later create additional `ns` scopes
- membership grants, not free-form `ns` ids, control access thereafter

Naming guidance for the first `ns`:

- prefer a normalized username slug when present
- otherwise derive from email local-part plus collision handling
- preserve a user-facing display name separately from the immutable id

## CLI Command Additions

Recommended commands:

- `lw login`
- `lw logout`
- `lw auth status`
- `lw auth token`
- `lw auth profiles list`
- `lw auth profiles use <name>`
- `lw auth profiles set <name>`

Minimal first cut:

- `lw login`
- `lw logout`
- `lw auth status`

## MCP And Remote Agent Alignment

MCP must not stay on legacy tenant-header-only auth long-term.

Recommended model:

- remote MCP over HTTP uses bearer tokens
- stdio MCP bridge accepts `--token`, `--token-file`, or inherited env
- MCP server maps the token into the same auth context used by HTTP routes
- tool visibility and tool execution must respect scopes

Example:

- a token with `documents.read` but not `documents.write` can see read tools or read resources only

## Cloud And Kubernetes Patterns

### Human-operated cloud CLI

- `lw login --base-url https://wiki.example.com`
- tokens stored in the user's home directory
- browser flow or device flow based on runtime capability

### K8s workload

- mount a short-lived token file into the pod
- point the CLI or MCP bridge at `--token-file`
- rotate via sidecar, CSI driver, or workload identity broker

### Internal service calling LLM-Wiki

- create a service principal in LLM-Wiki
- issue an `ns`-scoped token with explicit scopes
- rotate on a schedule
- log token ID and service principal ID on every write

### Delegated automation

- human user authorizes an automation run
- broker exchanges that approval for a short-lived delegated token
- token carries both `principal_id` and `delegated_by`

## Data Model Additions

Likely new persistence entities:

- `users`
- `teams`
- `service_principals`
- `tenant_memberships`
- `oauth_accounts`
- `auth_sessions`
- `api_tokens`
- `api_token_grants`

Fields each token record should keep:

- `id`
- `ns`
- `principal_type`
- `principal_id`
- `display_name`
- `scope_set`
- `resource_constraints`
- `issued_at`
- `expires_at`
- `revoked_at`
- `last_used_at`
- `created_by`

Only store token hashes server-side for static bearer tokens.

## Rollout Phases

### Phase 1: auth plumbing

- add auth context and middleware
- accept bearer token on HTTP routes
- keep the legacy tenant header only as temporary fallback
- start rejecting missing `ns` context on protected routes

### Phase 2: service tokens

- add service principal and token tables
- add issue, list, and revoke APIs
- wire scopes into HTTP and MCP authorization

### Phase 3: OAuth-backed user onboarding

- add OAuth provider config and account linkage
- add first-login user provisioning
- add first-login personal-`ns` creation
- make membership bootstrap deterministic

### Phase 4: CLI credential discovery and browser login

- add `--token`, `--token-file`, env support, and `~/.llm-wiki/` profiles
- add browser PKCE login with localhost callback
- remove the need to always pass `--ns` manually when token already binds an `ns`

### Phase 5: headless human fallback

- add device code flow
- add refresh token handling

### Phase 6: resource-aware policy

- move from `ns`-wide scopes to folder and document ACL integration
- reduce trust in user-supplied `author_*` fields

## Key Decisions Proposed

- Keep the CLI thin, but give it first-class credential discovery and login UX.
- Make bearer auth the primary access path for HTTP and MCP.
- Treat `ns` as a grant on the token, not a free-text parameter.
- Support both interactive OAuth login and explicit token use.
- Introduce service principals and short-lived delegated tokens for cloud automation.

## Risks

- Building a full first-party OAuth server too early will slow product progress.
- Keeping legacy tenant-header auth too long will leak into downstream integrations and be harder to remove later.
- Long-lived static service tokens without scope and resource constraints will create an avoidable security problem.
- If revision authorship continues to trust raw request fields, audit history will be weak even after auth lands.

## Recommended First Implementation Slice

The highest-leverage next slice is:

1. Add bearer-token auth context to the server.
2. Add OAuth account linkage and first-login user provisioning.
3. Add personal default-`ns` auto provisioning and membership bootstrap.
4. Move CLI to browser-first login with localhost callback and `~/.llm-wiki/` profile persistence.
5. Keep device code as explicit fallback for constrained environments.
6. Add service principals and `ns`-scoped service tokens after the human hosted flow is solid.

That gets the real hosted user journey working first, while still leaving service-token plumbing on the same auth substrate.
