---
name: "voidlab-ops-agent"
description: "Connects an agent to the VOIDLAB admin API with agent-token auth and scope checks. Invoke when an agent needs to verify access or operate the VOIDLAB backend safely."
---

# VOIDLAB Ops Agent

Use this skill when an agent needs to connect to the VOIDLAB backend and operate it through the API instead of the admin UI.

## Package Structure

This skill can now be maintained and shared as a complete folder-based package.

Read these reference files as needed:

- `references/overview.md`
- `references/articles.md`
- `references/events.md`
- `references/builders.md`
- `references/media.md`
- `references/publishing-flow.md`
- `references/error-handling.md`
- `references/share.md`

## Goal

This version covers three layers:

- the connection layer for agent-token authentication
- the first operational workflows for article, event, and builder management
- the knowledge-base reading protocol for Space unlock, protected entry reading, and protected asset rendering

It standardizes:

- how to authenticate with an agent token
- how to verify current access before doing work
- which API resources are available to agent tokens
- how to fail clearly on permission boundaries
- how to list articles, create drafts, and update article status safely
- how to list events, create event drafts, and update event status safely
- how to list builders, create builder drafts, and update builder status safely
- how to read a protected knowledge space through Space token verification
- how to read a protected knowledge entry through a short-lived grant
- how Markdown knowledge assets are referenced and resolved
- how to create, list, and rotate knowledge Space access tokens through dedicated agent scopes
- how to import a local Markdown file into a knowledge entry draft payload

## Required Environment

Set these values before making requests:

- `VOIDLAB_API_BASE_URL`: API origin, for example `http://127.0.0.1:18081`
- `VOIDLAB_AGENT_TOKEN`: agent token created from the admin panel

Knowledge reading can also require these values:

- `VOIDLAB_KNOWLEDGE_SPACE_TOKEN`: optional Space access token provided by the user for one protected knowledge space
- `VOIDLAB_KNOWLEDGE_GRANT`: optional cached grant returned by the verify-token endpoint after a successful unlock

Optional fallback:

- if `VOIDLAB_API_BASE_URL` is missing in local development, assume `http://127.0.0.1:18081`

## URL Resolution Rule

Resolve the backend URL in this priority order:

1. explicit URL provided by the user in the current request
2. `VOIDLAB_API_BASE_URL` from environment
3. local development default: `http://127.0.0.1:18081`

Ask the user for the URL only when:

- there is no explicit URL
- `VOIDLAB_API_BASE_URL` is not configured
- and the local default is clearly not the intended target environment

In normal day-to-day use, the user should not need to repeat the URL every time. Configure it once through environment or a stable default, then reuse it.

## Authentication Rules

Use Bearer auth on every protected request:

```bash
Authorization: Bearer $VOIDLAB_AGENT_TOKEN
```

Never prefer a human admin token when an agent token is available. Agent work should stay isolated, scope-limited, and auditable.

Knowledge-space reading is different:

- do not send the agent token to the public knowledge endpoints
- verify a protected knowledge space with the Space access token
- store the returned grant and reuse it for knowledge entry or asset reads in that same Space
- keep the agent token flow and the knowledge grant flow separate in both reasoning and request headers

## First-Run Handshake

Before any write operation, do this sequence:

1. Check service health with `GET /healthz`
2. Check identity with `GET /api/v1/auth/me`
3. Verify resource access with a cheap read on the target module
4. Only then perform the write

Example:

```bash
curl -s "$VOIDLAB_API_BASE_URL/healthz"
curl -s "$VOIDLAB_API_BASE_URL/api/v1/auth/me" \
  -H "Authorization: Bearer $VOIDLAB_AGENT_TOKEN"
curl -s "$VOIDLAB_API_BASE_URL/api/v1/articles" \
  -H "Authorization: Bearer $VOIDLAB_AGENT_TOKEN"
```

## Current Agent Scope Map

Supported scopes in the backend:

- `articles:read`
- `articles:write`
- `events:read`
- `events:write`
- `builders:read`
- `builders:write`
- `knowledge:read`
- `knowledge:write`
- `knowledge_tokens:read`
- `knowledge_tokens:write`
- `media:read`
- `media:write`

Scope behavior:

- `*:read` can access list and detail routes for that resource
- `*:write` can access create, update, status change, and delete routes
- some list/detail routes also accept the matching `*:write` scope

Current limitation:

- knowledge-space admin routes are now available to agent tokens that carry `knowledge:read` or `knowledge:write`
- knowledge access token management is available only to agent tokens that carry `knowledge_tokens:read` or `knowledge_tokens:write`
- agent tokens can operate articles, events, builders, media, knowledge content, and Space access tokens when they have the matching dedicated scopes

## Resource Endpoints

Agent-accessible content endpoints:

- Articles
  - `GET /api/v1/articles`
  - `GET /api/v1/articles/:id`
  - `POST /api/v1/articles`
  - `PUT /api/v1/articles/:id`
  - `PUT /api/v1/articles/:id/status`
  - `DELETE /api/v1/articles/:id`
- Events
  - `GET /api/v1/events`
  - `GET /api/v1/events/:id`
  - `POST /api/v1/events`
  - `PUT /api/v1/events/:id`
  - `PUT /api/v1/events/:id/status`
  - `DELETE /api/v1/events/:id`
- Builders
  - `GET /api/v1/builders`
  - `GET /api/v1/builders/:id`
  - `POST /api/v1/builders`
  - `PUT /api/v1/builders/:id`
  - `PUT /api/v1/builders/:id/status`
  - `DELETE /api/v1/builders/:id`
- Knowledge
  - `GET /api/v1/knowledge/spaces`
  - `POST /api/v1/knowledge/spaces`
  - `PUT /api/v1/knowledge/spaces/:id`
  - `DELETE /api/v1/knowledge/spaces/:id`
  - `GET /api/v1/knowledge/entries`
  - `POST /api/v1/knowledge/entries/import-markdown`
  - `POST /api/v1/knowledge/entries`
  - `PUT /api/v1/knowledge/entries/:id`
  - `DELETE /api/v1/knowledge/entries/:id`
  - `GET /api/v1/knowledge/spaces/:id/assets`
  - `POST /api/v1/knowledge/spaces/:id/assets`
  - `GET /api/v1/knowledge/access-tokens`
  - `POST /api/v1/knowledge/access-tokens`
  - `PUT /api/v1/knowledge/access-tokens/:id/status`
- Media
  - `GET /api/v1/media`
  - `POST /api/v1/media/upload`

Public knowledge endpoints use a different access model and do not use the agent token:

- `GET /api/v1/public/knowledge/spaces`
- `GET /api/v1/public/knowledge/spaces/:slug/toc`
- `POST /api/v1/public/knowledge/spaces/:slug/verify-token`
- `GET /api/v1/public/knowledge/spaces/:slug/entries/:entrySlug`
- `GET /api/v1/public/knowledge/spaces/:slug/assets/:assetID`

Admin-only endpoints are not available to agent tokens:

- `GET/POST/PUT /api/v1/agent-tokens`
- `GET/POST/PUT /api/v1/users`
- `GET/PUT /api/v1/site-configs`
- `GET /api/v1/audit-logs`
- all lead operations under `/api/v1/leads`
- dashboard and other admin-only groups that require human roles

## Knowledge Workflow

Use the knowledge workflow when the task is one of these:

- list the current published knowledge spaces
- inspect a Space directory before deciding what to read
- unlock a protected Space with a user-provided Space token
- read a protected knowledge entry after verification
- render or fetch protected Markdown assets referenced by a knowledge entry

Current capability boundary:

- agent tokens can manage knowledge spaces, entries, and knowledge assets when they have `knowledge:write`
- agent tokens can inspect knowledge admin lists when they have `knowledge:read`
- agent tokens can inspect knowledge access tokens when they have `knowledge_tokens:read`
- agent tokens can create or update knowledge access token status when they have `knowledge_tokens:write`

### Knowledge Token Model

Knowledge access uses three different objects:

1. `Knowledge access token`
   - provided by a human user
   - can be one of:
     - `basic`: unlocks one Space
     - `pro`: unlocks multiple specified Spaces
     - `vip`: unlocks all published Spaces
   - used only with `POST /api/v1/public/knowledge/spaces/:slug/verify-token`
2. `grant`
   - returned by the backend after verification
   - reused for protected entry reads and protected asset reads
3. `knowledge-asset://ASSET_ID`
   - Markdown placeholder for a protected knowledge image or file
   - must be resolved to the public knowledge asset endpoint for the current Space

Keep these strictly separate from the agent token:

- `VOIDLAB_AGENT_TOKEN` is for protected admin APIs
- knowledge access token and grant are for public knowledge unlock flow

### Knowledge Space List

List published knowledge spaces:

```bash
curl -s "$VOIDLAB_API_BASE_URL/api/v1/public/knowledge/spaces"
```

Use this before attempting to unlock or read a Space so the agent can confirm:

- available `slug`
- title
- short description
- whether the Space looks like the intended target

### Knowledge Space Directory

Read the published Space directory before reading entries:

```bash
curl -s "$VOIDLAB_API_BASE_URL/api/v1/public/knowledge/spaces/SPACE_SLUG/toc"
```

Important behavior:

- the directory endpoint intentionally does not return full `content_markdown`
- it only returns the Space metadata and entry metadata needed for navigation
- directory access is public for `public` and `directory_only` spaces

### Verify Knowledge Space Token

When a Space is protected and the user provides a knowledge access token, verify it like this:

```bash
curl -s -X POST "$VOIDLAB_API_BASE_URL/api/v1/public/knowledge/spaces/SPACE_SLUG/verify-token" \
  -H "Content-Type: application/json" \
  -d '{
    "token": "USER_PROVIDED_ACCESS_TOKEN"
  }'
```

Expected success behavior:

- response returns a `grant`
- response also describes the unlock range
- the grant can be:
  - single-Space
  - multi-Space
  - all published Spaces
- store that grant and reuse it for entry reads and asset reads in every allowed Space

If verification fails:

- report that the access token is invalid or inactive for the current Space
- do not keep retrying with the same token

### Read Protected Knowledge Entry

If the entry is not a preview, read it with the grant:

```bash
curl -s "$VOIDLAB_API_BASE_URL/api/v1/public/knowledge/spaces/SPACE_SLUG/entries/ENTRY_SLUG" \
  -H "X-Knowledge-Grant: KNOWLEDGE_GRANT"
```

Alternative query-string form is also accepted:

```bash
curl -s "$VOIDLAB_API_BASE_URL/api/v1/public/knowledge/spaces/SPACE_SLUG/entries/ENTRY_SLUG?grant=KNOWLEDGE_GRANT"
```

Behavior rules:

- prefer the header form when making direct HTTP calls
- if the entry is a preview, the backend may allow reading it without a grant
- if the entry is locked and there is no valid grant, expect `403`

### Resolve Protected Markdown Assets

Knowledge Markdown can contain protected assets using this placeholder format:

```md
![Architecture](knowledge-asset://12)
```

Resolve it to:

```text
/api/v1/public/knowledge/spaces/SPACE_SLUG/assets/12?grant=KNOWLEDGE_GRANT
```

Rules:

- only resolve `knowledge-asset://ASSET_ID` placeholders this way
- if the Markdown image source is already `http`, `https`, or `data:`, leave it unchanged
- if the source is not a knowledge asset placeholder, treat it like a normal image URL
- use the same Space grant that unlocked the entry

### Knowledge Reading Policy

- always read the Space directory before reading a protected entry
- do not ask for a Space token unless the entry is actually locked
- cache the grant per `spaceSlug` for the current session
- do not leak raw Space tokens in summaries after a successful unlock
- when summarizing a knowledge read, report `space`, `entry`, and whether it was `preview` or `grant-protected`

## Knowledge Admin Workflow

Use the knowledge admin workflow when the task is one of these:

- create a new knowledge space
- update a knowledge space title, description, visibility mode, intro, or status
- create or update a Markdown knowledge entry
- import a local Markdown file before creating or updating a knowledge entry
- list current knowledge spaces or entries before editing
- upload a protected knowledge asset and insert it into Markdown
- list current knowledge Space access tokens
- create a new knowledge Space access token
- enable or disable an existing knowledge Space access token

Required scopes:

- `knowledge:read` for listing spaces, entries, and knowledge assets
- `knowledge:write` for create, update, delete, and asset upload
- `knowledge_tokens:read` for listing knowledge Space access tokens
- `knowledge_tokens:write` for creating or updating knowledge Space access tokens

### Knowledge Space Field Contract

Knowledge space create and update use this payload shape:

```json
{
  "title": "string",
  "slug": "string",
  "description": "string",
  "cover_label": "string",
  "icon": "string",
  "theme_tint": "string",
  "visibility_mode": "directory_only",
  "directory_summary": "string",
  "intro_markdown": "string",
  "token_hint": "string",
  "cover_url": "string",
  "status": "draft"
}
```

Minimum required fields:

- `title`
- `slug`

### Knowledge Entry Field Contract

Knowledge entry create and update use this payload shape:

```json
{
  "space_id": 1,
  "title": "string",
  "slug": "string",
  "section_name": "string",
  "sort_order": 10,
  "estimated_read_minutes": 8,
  "public_summary": "string",
  "content_markdown": "string",
  "cover_url": "string",
  "is_preview": false,
  "status": "draft"
}
```

Minimum required fields:

- `space_id`
- `title`
- `slug`

### Knowledge Asset Upload

Upload a protected image or file to a Space:

```bash
curl -s -X POST "$VOIDLAB_API_BASE_URL/api/v1/knowledge/spaces/SPACE_ID/assets" \
  -H "Authorization: Bearer $VOIDLAB_AGENT_TOKEN" \
  -F "file=@/absolute/path/to/image.png"
```

Expected response includes:

- `asset`
- `markdown_url`
- `markdown_snippet`
- `public_url`

Prefer inserting the returned `markdown_snippet` directly into `content_markdown`.

Example:

```md
![system-architecture](knowledge-asset://12)
```

### Knowledge Markdown Import

Import a local Markdown file and let the backend derive the entry draft fields:

```bash
curl -s -X POST "$VOIDLAB_API_BASE_URL/api/v1/knowledge/entries/import-markdown" \
  -H "Authorization: Bearer $VOIDLAB_AGENT_TOKEN" \
  -F "file=@/absolute/path/to/doc.md"
```

Expected response includes a draft-shaped payload such as:

- `title`
- `slug`
- `section_name`
- `estimated_read_minutes`
- `public_summary`
- `content_markdown`

Use it as a prefill step before `POST /api/v1/knowledge/entries` or `PUT /api/v1/knowledge/entries/:id`.

### Knowledge Access Token Field Contract

Knowledge access token create uses this payload shape:

```json
{
  "space_id": 1,
  "name": "string",
  "expires_at": "2026-12-31 23:59"
}
```

Minimum required fields:

- `access_level`
- `name`

Level-specific requirements:

- `basic`
  - requires exactly one target Space
  - prefer `space_id`
- `pro`
  - requires one or more target Spaces
  - prefer `space_ids`
- `vip`
  - requires no Space binding
  - do not send `space_id` unless the backend explicitly requires a current context lookup first

Knowledge access token status updates use this payload shape:

```json
{
  "is_active": true
}
```

### Knowledge Access Token Operations

List current knowledge access tokens:

```bash
curl -s "$VOIDLAB_API_BASE_URL/api/v1/knowledge/access-tokens?space_id=SPACE_ID" \
  -H "Authorization: Bearer $VOIDLAB_AGENT_TOKEN"
```

Create a new `basic` single-Space token:

```bash
curl -s -X POST "$VOIDLAB_API_BASE_URL/api/v1/knowledge/access-tokens" \
  -H "Authorization: Bearer $VOIDLAB_AGENT_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "access_level": "basic",
    "space_id": 1,
    "name": "founder-playbook-q4",
    "expires_at": "2026-12-31 23:59"
  }'
```

Create a `pro` multi-Space token:

```bash
curl -s -X POST "$VOIDLAB_API_BASE_URL/api/v1/knowledge/access-tokens" \
  -H "Authorization: Bearer $VOIDLAB_AGENT_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "access_level": "pro",
    "space_ids": [1, 2, 5],
    "name": "founder-system-bundle",
    "expires_at": "2026-12-31 23:59"
  }'
```

Create a `vip` all-published token:

```bash
curl -s -X POST "$VOIDLAB_API_BASE_URL/api/v1/knowledge/access-tokens" \
  -H "Authorization: Bearer $VOIDLAB_AGENT_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "access_level": "vip",
    "name": "voidlab-vip-pass",
    "expires_at": ""
  }'
```

Enable or disable an existing knowledge access token:

```bash
curl -s -X PUT "$VOIDLAB_API_BASE_URL/api/v1/knowledge/access-tokens/TOKEN_ID/status" \
  -H "Authorization: Bearer $VOIDLAB_AGENT_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "is_active": false
  }'
```

Knowledge access token policy:

- use `knowledge_tokens:read` and `knowledge_tokens:write` instead of overloading `knowledge:write`
- treat the plain token returned on create as a secret and do not repeat it in later summaries
- prefer disabling a token through status change when the user wants to revoke access
- if the target Space scope is unclear, identify whether the user wants `basic`, `pro`, or `vip` before creating the token

### Knowledge Operation Policy

- always list or read current spaces before creating related entries if the target `space_id` is not explicit
- for entry updates, read the current record first and merge changes before sending the full payload
- keep `content_markdown` as the source of truth for tutorials, docs, and note-style content
- for image insertion, upload first, then insert the returned `knowledge-asset://ASSET_ID` placeholder into Markdown
- use the dedicated `knowledge_tokens:*` scopes when the task is token lifecycle management

## Article Workflow

Use the article workflow when the task is one of these:

- list existing articles before deciding what to edit
- create a new article draft
- publish, unpublish, or archive an existing article

Required scopes:

- `articles:read` for listing and reading details
- `articles:write` for create, update, status change, and delete

### Article Field Contract

Article create and update use this payload shape:

```json
{
  "title": "string",
  "slug": "string",
  "summary": "string",
  "category": "string",
  "audience": "string",
  "tags": ["string"],
  "cover_url": "string",
  "content": "string",
  "source_name": "string",
  "source_url": "string",
  "featured": false,
  "status": "draft"
}
```

Minimum required fields:

- `title`
- `slug`

Defaults and rules:

- if `status` is omitted on create, the backend defaults it to `draft`
- valid create statuses are `draft` and `published`
- valid update statuses are `draft`, `published`, and `archived`, but transitions are restricted
- `content` can now be a Markdown document body; for news and insight content, prefer writing structured Markdown instead of one flat paragraph

### Article Status Flow

Allowed transitions:

- `draft -> published`
- `published -> draft`
- `published -> archived`
- `archived -> draft`

Avoid using the full article update endpoint for status-only changes. Prefer:

- `PUT /api/v1/articles/:id/status`

### Article List

List current articles before creating or editing content:

```bash
curl -s "$VOIDLAB_API_BASE_URL/api/v1/articles" \
  -H "Authorization: Bearer $VOIDLAB_AGENT_TOKEN"
```

### Create Article Draft

When the user gives a title but no slug, derive a short kebab-case slug before sending the request.

For current VOIDLAB news content:

- prefer Markdown in `content`
- usually start the body from `##` sections rather than `#`
- preserve bullets and numbered lists when the source is already a Markdown document

Example:

```bash
curl -s -X POST "$VOIDLAB_API_BASE_URL/api/v1/articles" \
  -H "Authorization: Bearer $VOIDLAB_AGENT_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "title": "New Builder Roundup",
    "slug": "new-builder-roundup",
    "summary": "Weekly builder update draft.",
    "category": "insights",
    "audience": "builders",
    "tags": ["weekly", "builders"],
    "content": "Initial draft content.",
    "status": "draft"
  }'
```

### Update Article Status

Publish an article:

```bash
curl -s -X PUT "$VOIDLAB_API_BASE_URL/api/v1/articles/ARTICLE_ID/status" \
  -H "Authorization: Bearer $VOIDLAB_AGENT_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "status": "published"
  }'
```

Move a published article back to draft:

```bash
curl -s -X PUT "$VOIDLAB_API_BASE_URL/api/v1/articles/ARTICLE_ID/status" \
  -H "Authorization: Bearer $VOIDLAB_AGENT_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "status": "draft"
  }'
```

Archive a published article:

```bash
curl -s -X PUT "$VOIDLAB_API_BASE_URL/api/v1/articles/ARTICLE_ID/status" \
  -H "Authorization: Bearer $VOIDLAB_AGENT_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "status": "archived"
  }'
```

### Update Article Content

Use `PUT /api/v1/articles/:id` only when fields other than status need to change.

Important:

- the update endpoint expects a complete article payload with at least `title` and `slug`
- if editing an existing article, fetch the current record first, merge the intended changes, then send the full payload back
- do not send partial payloads unless you are certain required fields are included

Recommended sequence:

1. `GET /api/v1/articles/:id`
2. merge requested field changes
3. `PUT /api/v1/articles/:id`

Safe merge checklist:

- preserve the current `title` and `slug` unless the user explicitly wants to change them
- preserve `status` unless the task is a real status change
- preserve optional fields such as `cover_url`, `source_name`, `source_url`, `featured`, and `tags` when they should remain unchanged
- if the record is already `published`, a normal full update can keep `status: "published"` as long as the payload is otherwise valid

Read the current article first:

```bash
curl -s "$VOIDLAB_API_BASE_URL/api/v1/articles/ARTICLE_ID" \
  -H "Authorization: Bearer $VOIDLAB_AGENT_TOKEN"
```

Then send the merged full payload:

```bash
curl -s -X PUT "$VOIDLAB_API_BASE_URL/api/v1/articles/ARTICLE_ID" \
  -H "Authorization: Bearer $VOIDLAB_AGENT_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "title": "Agent Auth Smoke Test",
    "slug": "agent-auth-smoke-test",
    "summary": "runtime validation updated by full payload",
    "category": "ops-automation",
    "audience": "internal-team",
    "tags": ["agent", "auth", "full-update"],
    "cover_url": "",
    "content": "Created by agent token validation and updated through the full article endpoint.",
    "source_name": "VOIDLAB Agent",
    "source_url": "https://voidlab.ai",
    "featured": true,
    "status": "published"
  }'
```

Expected success behavior:

- response returns the updated article record
- audit log records `action=update`
- actor is stored as `actor_type=agent` when the request uses an agent token

Common mistake:

- sending only the changed field such as `summary` or `content` without required fields; this can fail validation because `title` and `slug` are still required by the update endpoint

### Article Operation Policy

- always list or read the article first before updating it
- prefer creating drafts before publishing directly unless the user explicitly asks for immediate publish
- for status changes, use the dedicated status endpoint instead of the full update endpoint
- after a successful write, summarize the resulting `id`, `slug`, and `status`
- remember that agent writes are audited as `actor_type=agent`

## Event Workflow

Use the event workflow when the task is one of these:

- list current events before editing or creating a new one
- create a new event draft
- publish, unpublish, or archive an event
- update event logistics such as city, location, type, time, or copy

Required scopes:

- `events:read` for listing and reading details
- `events:write` for create, update, status change, and delete

### Event Field Contract

Event create and update use this payload shape:

```json
{
  "title": "string",
  "slug": "string",
  "summary": "string",
  "city": "string",
  "location": "string",
  "event_type": "string",
  "event_time": "string",
  "cover_url": "string",
  "content": "string",
  "status": "draft"
}
```

Minimum required fields:

- `title`
- `slug`

Defaults and rules:

- if `status` is omitted on create, the backend defaults it to `draft`
- valid create statuses are `draft` and `published`
- valid update statuses use the same content status flow as articles
- `event_time` is currently stored as a plain string, so preserve its format consistently across updates

### Event Status Flow

Allowed transitions:

- `draft -> published`
- `published -> draft`
- `published -> archived`
- `archived -> draft`

For status-only changes, prefer:

- `PUT /api/v1/events/:id/status`

### Event List

List current events before editing or publishing:

```bash
curl -s "$VOIDLAB_API_BASE_URL/api/v1/events" \
  -H "Authorization: Bearer $VOIDLAB_AGENT_TOKEN"
```

### Create Event Draft

When the user gives an event title but no slug, derive a short kebab-case slug before sending the request.

Example:

```bash
curl -s -X POST "$VOIDLAB_API_BASE_URL/api/v1/events" \
  -H "Authorization: Bearer $VOIDLAB_AGENT_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "title": "Builder Mixer Shanghai",
    "slug": "builder-mixer-shanghai",
    "summary": "Community event draft.",
    "city": "Shanghai",
    "location": "Xuhui Hub",
    "event_type": "Salon",
    "event_time": "2026-08-31T11:00:00Z",
    "content": "Draft event description.",
    "status": "draft"
  }'
```

### Update Event Status

Publish an event:

```bash
curl -s -X PUT "$VOIDLAB_API_BASE_URL/api/v1/events/EVENT_ID/status" \
  -H "Authorization: Bearer $VOIDLAB_AGENT_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "status": "published"
  }'
```

Move a published event back to draft:

```bash
curl -s -X PUT "$VOIDLAB_API_BASE_URL/api/v1/events/EVENT_ID/status" \
  -H "Authorization: Bearer $VOIDLAB_AGENT_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "status": "draft"
  }'
```

Archive a published event:

```bash
curl -s -X PUT "$VOIDLAB_API_BASE_URL/api/v1/events/EVENT_ID/status" \
  -H "Authorization: Bearer $VOIDLAB_AGENT_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "status": "archived"
  }'
```

### Update Event Content

Use `PUT /api/v1/events/:id` when logistics or event copy need to change.

Important:

- the update endpoint expects a complete event payload with at least `title` and `slug`
- fetch the current record first, merge intended changes, then send the full payload back
- preserve `status` unless the task is a real status change

Recommended sequence:

1. `GET /api/v1/events/:id`
2. merge requested field changes
3. `PUT /api/v1/events/:id`

Safe merge checklist:

- preserve the current `title` and `slug` unless the user explicitly wants to change them
- preserve `status` unless the task is a real status change
- preserve `city`, `location`, `event_type`, and `event_time` unless the user explicitly changes them
- preserve `cover_url` if the current event already uses one

Read the current event first:

```bash
curl -s "$VOIDLAB_API_BASE_URL/api/v1/events/EVENT_ID" \
  -H "Authorization: Bearer $VOIDLAB_AGENT_TOKEN"
```

Then send the merged full payload:

```bash
curl -s -X PUT "$VOIDLAB_API_BASE_URL/api/v1/events/EVENT_ID" \
  -H "Authorization: Bearer $VOIDLAB_AGENT_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "title": "Agent Event Smoke Test",
    "slug": "agent-event-smoke-test",
    "summary": "runtime event validation updated by full payload",
    "city": "Shanghai",
    "location": "Xuhui Innovation Hub",
    "event_type": "Builder Salon",
    "event_time": "2026-09-01T12:30:00Z",
    "cover_url": "",
    "content": "Created and updated through the event agent token workflow.",
    "status": "published"
  }'
```

Expected success behavior:

- response returns the updated event record
- audit log records `action=update` or `action=update_status`
- actor is stored as `actor_type=agent` when the request uses an agent token

Common mistake:

- sending only the changed field such as `location` or `event_time` without the rest of the required payload; this can fail validation because `title` and `slug` are still required

### Event Operation Policy

- always list or read the event first before updating it
- prefer creating drafts before publishing directly unless the user explicitly asks for immediate publish
- for status changes, use the dedicated status endpoint instead of the full update endpoint
- after a successful write, summarize the resulting `id`, `slug`, `status`, and `event_time`
- remember that agent writes are audited as `actor_type=agent`

## Builder Workflow

Use the builder workflow when the task is one of these:

- list current builder cards before editing or creating a new one
- create a new builder draft
- publish, unpublish, or archive a builder card
- update builder profile fields such as title, role, city, intro, story, or collaboration setup

Required scopes:

- `builders:read` for listing and reading details
- `builders:write` for create, update, status change, and delete

### Builder Field Contract

Builder create and update use this payload shape:

```json
{
  "name": "string",
  "slug": "string",
  "title": "string",
  "city": "string",
  "role": "string",
  "intro": "string",
  "story": "string",
  "expertise": ["string"],
  "focus_areas": ["string"],
  "collaboration_modes": ["string"],
  "cover_url": "string",
  "contactable": true,
  "featured": false,
  "status": "draft"
}
```

Minimum required fields:

- `name`
- `slug`

Defaults and rules:

- if `status` is omitted on create, the backend defaults it to `draft`
- valid create statuses are `draft` and `published`
- valid update statuses use the same content status flow as articles and events
- array fields such as `expertise`, `focus_areas`, and `collaboration_modes` should be sent as full arrays on update

### Builder Status Flow

Allowed transitions:

- `draft -> published`
- `published -> draft`
- `published -> archived`
- `archived -> draft`

For status-only changes, prefer:

- `PUT /api/v1/builders/:id/status`

### Builder List

List current builders before editing or publishing:

```bash
curl -s "$VOIDLAB_API_BASE_URL/api/v1/builders" \
  -H "Authorization: Bearer $VOIDLAB_AGENT_TOKEN"
```

### Create Builder Draft

When the user gives a builder name but no slug, derive a short kebab-case slug before sending the request.

Example:

```bash
curl -s -X POST "$VOIDLAB_API_BASE_URL/api/v1/builders" \
  -H "Authorization: Bearer $VOIDLAB_AGENT_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Jane Builder",
    "slug": "jane-builder",
    "title": "Community Operator",
    "city": "Shanghai",
    "role": "Builder",
    "intro": "Draft builder intro.",
    "story": "Draft builder story.",
    "expertise": ["Community", "AI Ops"],
    "focus_areas": ["AI", "Partnerships"],
    "collaboration_modes": ["Advisory", "Workshop"],
    "cover_url": "",
    "contactable": true,
    "featured": false,
    "status": "draft"
  }'
```

### Update Builder Status

Publish a builder:

```bash
curl -s -X PUT "$VOIDLAB_API_BASE_URL/api/v1/builders/BUILDER_ID/status" \
  -H "Authorization: Bearer $VOIDLAB_AGENT_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "status": "published"
  }'
```

Move a published builder back to draft:

```bash
curl -s -X PUT "$VOIDLAB_API_BASE_URL/api/v1/builders/BUILDER_ID/status" \
  -H "Authorization: Bearer $VOIDLAB_AGENT_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "status": "draft"
  }'
```

Archive a published builder:

```bash
curl -s -X PUT "$VOIDLAB_API_BASE_URL/api/v1/builders/BUILDER_ID/status" \
  -H "Authorization: Bearer $VOIDLAB_AGENT_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "status": "archived"
  }'
```

### Update Builder Content

Use `PUT /api/v1/builders/:id` when profile content or structured arrays need to change.

Important:

- the update endpoint expects a complete builder payload with at least `name` and `slug`
- fetch the current record first, merge intended changes, then send the full payload back
- preserve `status` unless the task is a real status change
- preserve array fields unless the user explicitly wants to replace them

Recommended sequence:

1. `GET /api/v1/builders/:id`
2. merge requested field changes
3. `PUT /api/v1/builders/:id`

Safe merge checklist:

- preserve the current `name` and `slug` unless the user explicitly wants to change them
- preserve `status` unless the task is a real status change
- preserve `contactable` and `featured` unless the user explicitly changes them
- preserve `expertise`, `focus_areas`, and `collaboration_modes` unless the user explicitly changes those arrays
- preserve `cover_url` if the current builder already uses one

Read the current builder first:

```bash
curl -s "$VOIDLAB_API_BASE_URL/api/v1/builders/BUILDER_ID" \
  -H "Authorization: Bearer $VOIDLAB_AGENT_TOKEN"
```

Then send the merged full payload:

```bash
curl -s -X PUT "$VOIDLAB_API_BASE_URL/api/v1/builders/BUILDER_ID" \
  -H "Authorization: Bearer $VOIDLAB_AGENT_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Agent Builder Smoke Test",
    "slug": "agent-builder-smoke-test",
    "title": "Community Systems Operator",
    "city": "Shanghai",
    "role": "Builder-Operator",
    "intro": "runtime builder validation updated by full payload",
    "story": "Created and updated through the builder agent token workflow.",
    "expertise": ["Community", "AI Ops", "Ecosystem"],
    "focus_areas": ["AI", "Partnerships", "Builder Network"],
    "collaboration_modes": ["Advisory", "Workshop", "Co-building"],
    "cover_url": "",
    "contactable": true,
    "featured": true,
    "status": "published"
  }'
```

Expected success behavior:

- response returns the updated builder record
- audit log records `action=update` or `action=update_status`
- actor is stored as `actor_type=agent` when the request uses an agent token

Common mistake:

- sending only one changed field such as `intro` or `role` without the rest of the required payload; this can fail validation because `name` and `slug` are still required

### Builder Operation Policy

- always list or read the builder first before updating it
- prefer creating drafts before publishing directly unless the user explicitly asks for immediate publish
- for status changes, use the dedicated status endpoint instead of the full update endpoint
- after a successful write, summarize the resulting `id`, `slug`, `status`, and `featured`
- remember that agent writes are audited as `actor_type=agent`

## Standard Actions

Use these standard actions when the user gives a direct operating instruction instead of asking for low-level API details.

### Intake Rule

Before executing any standard action:

1. identify the target resource: `article`, `event`, or `builder`
2. identify the intent: `list`, `create draft`, `update`, `publish`, `move to draft`, or `archive`
3. check whether the necessary fields are already present
4. if key required fields are missing, ask only for the missing fields
5. if the task is an update, read the current record first and merge changes

### Default Assumptions

Use these defaults unless the user explicitly says otherwise:

- create as `draft`
- preserve current `slug`
- preserve current `status` during full updates
- preserve optional fields and arrays when the user did not ask to change them
- summarize the final object using `id`, `slug`, and `status`

### Standard Action: Create Article Draft

Invoke when the user says things like:

- create an article draft
- add a new insight
- save this as an article

Minimum input:

- `title`

Recommended input:

- `summary`
- `category`
- `audience`
- `content`

Execution steps:

1. derive `slug` from title if missing
2. fill omitted optional fields with empty strings or empty arrays where appropriate
3. send `POST /api/v1/articles` with `status: "draft"`
4. report the created `id`, `slug`, and `status`

### Standard Action: Publish Article

Invoke when the user says things like:

- publish this article
- put this live
- change article status to published

Minimum input:

- article `id` or a uniquely identifiable existing article

Execution steps:

1. read or identify the target article
2. call `PUT /api/v1/articles/:id/status`
3. report the updated `id`, `slug`, and `status`

### Standard Action: Create Event Draft

Invoke when the user says things like:

- create an event draft
- add a new meetup
- create an event card

Minimum input:

- `title`

Recommended input:

- `city`
- `location`
- `event_type`
- `event_time`
- `summary`

Execution steps:

1. derive `slug` from title if missing
2. fill omitted optional fields with empty strings
3. send `POST /api/v1/events` with `status: "draft"`
4. report the created `id`, `slug`, `status`, and `event_time`

### Standard Action: Publish Event

Invoke when the user says things like:

- publish this event
- put this event live
- open event registration

Minimum input:

- event `id` or a uniquely identifiable existing event

Execution steps:

1. read or identify the target event
2. call `PUT /api/v1/events/:id/status`
3. report the updated `id`, `slug`, `status`, and `event_time`

### Standard Action: Create Builder Draft

Invoke when the user says things like:

- create a builder card
- add a builder profile
- save this person as a builder

Minimum input:

- `name`

Recommended input:

- `title`
- `city`
- `role`
- `intro`

Execution steps:

1. derive `slug` from name if missing
2. fill omitted arrays with empty arrays
3. send `POST /api/v1/builders` with `status: "draft"`
4. report the created `id`, `slug`, `status`, and `featured`

### Standard Action: Publish Builder

Invoke when the user says things like:

- publish this builder
- put this profile live
- make this builder visible

Minimum input:

- builder `id` or a uniquely identifiable existing builder

Execution steps:

1. read or identify the target builder
2. call `PUT /api/v1/builders/:id/status`
3. report the updated `id`, `slug`, `status`, and `featured`

### Standard Action: Full Update

Invoke when the user wants to revise existing content instead of only changing status.

Execution rule:

- always read the current record first
- merge only the requested changes
- send the full payload required by that resource
- keep the current `status` unless the user explicitly requested a status change

### Standard Action: List Resource

Invoke when the user says things like:

- show me current articles
- list events
- what builder cards do we have

Execution steps:

1. call the corresponding list endpoint
2. summarize records briefly with the most useful identity fields
3. do not dump unnecessary raw payload unless the user explicitly asks for it

## Natural Language Mapping

Map everyday operating language to the standard actions above instead of waiting for the user to speak in API terms.

### Chinese Command Mapping

Article intent examples:

- `帮我发一篇文章草稿` -> `Create Article Draft`
- `新增一篇资讯` -> `Create Article Draft`
- `把这篇文章发布掉` -> `Publish Article`
- `把这篇文章撤回草稿` -> `move article to draft`
- `改一下这篇文章的标题和摘要` -> `Full Update` on article
- `看看现在有哪些文章` -> `List Resource` for articles

Event intent examples:

- `新增一个活动` -> `Create Event Draft`
- `帮我建一个 meetup 草稿` -> `Create Event Draft`
- `把这个活动发布` -> `Publish Event`
- `把这个活动改回草稿` -> `move event to draft`
- `改一下这个活动的时间和地点` -> `Full Update` on event
- `列一下当前活动` -> `List Resource` for events

Builder intent examples:

- `新增一个 builder 卡片` -> `Create Builder Draft`
- `帮我录入这个人` -> `Create Builder Draft`
- `把这个 builder 发出去` -> `Publish Builder`
- `把这个 builder 改回草稿` -> `move builder to draft`
- `更新这个 builder 的介绍和标签` -> `Full Update` on builder
- `看下目前有哪些 builder` -> `List Resource` for builders

Knowledge intent examples:

- `新增一个知识库项目` -> `Create Knowledge Space`
- `新建一篇知识文档` -> `Create Knowledge Entry`
- `更新这个知识库的介绍` -> `Full Update` on knowledge space
- `更新这篇知识文档的 Markdown 内容` -> `Full Update` on knowledge entry
- `给这篇文档插一张图` -> `Upload Knowledge Asset` then `Full Update` on knowledge entry
- `列一下现在有哪些知识库项目` -> `List Resource` for knowledge spaces
- `给这个知识库生成一个令牌` -> `Create Knowledge Access Token`
- `把这个知识库令牌停掉` -> `Update Knowledge Access Token Status`

Knowledge intent examples:

- `看看知识库有哪些项目` -> `List Knowledge Spaces`
- `打开这个知识空间目录` -> `Read Knowledge Space Directory`
- `帮我解锁这个知识库` -> `Verify Knowledge Space Token`
- `读取这篇知识文档` -> `Read Protected Knowledge Entry`
- `这篇知识文档里的图为什么不显示` -> `Resolve Protected Markdown Assets`

### English Command Mapping

Article examples:

- `create an article draft` -> `Create Article Draft`
- `publish this article` -> `Publish Article`
- `update this article summary` -> `Full Update` on article
- `list current articles` -> `List Resource` for articles

Event examples:

- `create an event draft` -> `Create Event Draft`
- `publish this event` -> `Publish Event`
- `update the event time and venue` -> `Full Update` on event
- `list current events` -> `List Resource` for events

Builder examples:

- `create a builder profile` -> `Create Builder Draft`
- `publish this builder` -> `Publish Builder`
- `update the builder intro` -> `Full Update` on builder
- `list builder cards` -> `List Resource` for builders

Knowledge examples:

- `create a knowledge space` -> `Create Knowledge Space`
- `create a knowledge entry` -> `Create Knowledge Entry`
- `update this knowledge space intro` -> `Full Update` on knowledge space
- `upload an image for this knowledge doc` -> `Upload Knowledge Asset`
- `list current knowledge spaces` -> `List Resource` for knowledge spaces
- `create a token for this knowledge space` -> `Create Knowledge Access Token`
- `create a basic token for this space` -> `Create Knowledge Access Token` with `access_level=basic`
- `create a bundle token for these spaces` -> `Create Knowledge Access Token` with `access_level=pro`
- `create a vip token for all knowledge spaces` -> `Create Knowledge Access Token` with `access_level=vip`
- `disable this knowledge token` -> `Update Knowledge Access Token Status`

Knowledge examples:

- `list current knowledge spaces` -> `List Knowledge Spaces`
- `open this knowledge space directory` -> `Read Knowledge Space Directory`
- `unlock this knowledge space` -> `Verify Knowledge Space Token`
- `read this knowledge entry` -> `Read Protected Knowledge Entry`
- `resolve protected knowledge images` -> `Resolve Protected Markdown Assets`

## Slot Extraction Guide

When the user speaks naturally, extract slots in this order.

### Article Slots

Primary slots:

- `title`
- `slug`
- `summary`
- `content`

Secondary slots:

- `category`
- `audience`
- `tags`
- `featured`

### Event Slots

Primary slots:

- `title`
- `slug`
- `event_time`
- `location`

Secondary slots:

- `city`
- `event_type`
- `summary`
- `content`

### Builder Slots

Primary slots:

- `name`
- `slug`

Secondary slots:

- `title`
- `city`
- `role`
- `intro`
- `story`
- `expertise`
- `focus_areas`
- `collaboration_modes`
- `contactable`
- `featured`

### Knowledge Slots

Primary slots:

- `spaceId`
- `spaceSlug`
- `entrySlug`

Secondary slots:

- `spaceToken`
- `grant`
- `assetID`
- `content_markdown`

## Ambiguity Resolution

When the user gives a short command, resolve ambiguity using these rules:

- if the user says `发布这个`, prefer the most recently discussed resource in the current thread
- if the user provides an `id`, trust the `id` over title or name text
- if the user provides a title or name but no `id`, list or read records to identify the target
- if multiple records could match, ask a narrow disambiguation question with the candidate names
- if the user says `更新` without specifying fields, ask what fields should change instead of guessing
- if the user asks to manage knowledge content through the agent token, first confirm the token includes `knowledge:read` or `knowledge:write`
- if the user asks to manage knowledge access tokens through the agent token, first confirm the token includes `knowledge_tokens:read` or `knowledge_tokens:write`
- when creating a knowledge access token, always extract or confirm:
  - `access_level`: `basic` | `pro` | `vip`
  - target Space scope:
    - one Space for `basic`
    - one or more Spaces for `pro`
    - no explicit Space binding for `vip`
- if the user asks for “全局令牌”, “高级令牌”, “VIP 令牌”, or “一个能看全部的令牌”, map to `access_level=vip`
- if the user asks for “专题包令牌”, “多空间令牌”, or “一个能看多个项目的令牌”, map to `access_level=pro`
- if the user asks for “单空间令牌”, “这个项目的令牌”, or “只解锁一个的令牌”, map to `access_level=basic`

## Ready-to-Run Examples

Use these patterns when acting on behalf of the user.

### Example: Chinese Draft Creation

User instruction:

- `帮我新增一个活动，标题叫 AI Builder Meetup，地点徐汇，时间下周三晚上`

Agent behavior:

1. map to `Create Event Draft`
2. extract `title`, `location`, and approximate `event_time` if it is already specific enough
3. if time is still too fuzzy for a concrete payload, ask only for the exact datetime string
4. create the draft
5. report `id`, `slug`, `status`, and `event_time`

### Example: Chinese Publish Command

User instruction:

- `把 Agent Event Smoke Test 发布掉`

Agent behavior:

1. map to `Publish Event`
2. find the matching event by `id` or title
3. call the status endpoint
4. report the resulting `id`, `slug`, `status`, and `event_time`

### Example: Builder Update Command

User instruction:

- `把这个 builder 的 intro 和 focus areas 改一下`

Agent behavior:

1. map to `Full Update` on builder
2. identify the target builder
3. read the current builder record
4. merge only `intro` and `focus_areas`
5. send the full payload
6. report the resulting `id`, `slug`, `status`, and changed fields

### Example: Knowledge Unlock Command

User instruction:

- `帮我解锁 agent-builder-playbook，这里是 token`

Agent behavior:

1. map to `Verify Knowledge Space Token`
2. confirm the target `spaceSlug`
3. call the verify-token endpoint with the user-provided knowledge access token
4. store the returned grant and its unlock scope
5. report the unlocked range without repeating the raw token

### Example: Create A Basic Knowledge Token

User instruction:

- `帮我给 founder playbook 新增一个基础令牌，月底过期`

Agent behavior:

1. map to `Create Knowledge Access Token`
2. identify the target Space and its `space_id`
3. set `access_level=basic`
4. create the token with one bound Space
5. return the one-time plain token carefully and avoid repeating it later

### Example: Create A Pro Bundle Token

User instruction:

- `帮我做一个专题包令牌，给 founder playbook、agent builder、prompt system 这三个知识库都能看`

Agent behavior:

1. map to `Create Knowledge Access Token`
2. identify each target Space and collect `space_ids`
3. set `access_level=pro`
4. create the token with `space_ids`
5. summarize the covered Spaces and return the one-time plain token

### Example: Create A VIP Global Token

User instruction:

- `帮我新增一个全局 VIP 令牌，所有知识库都能看`

Agent behavior:

1. map to `Create Knowledge Access Token`
2. set `access_level=vip`
3. do not require explicit `space_id`
4. create the token for all published Spaces
5. state clearly that it unlocks all published knowledge Spaces

### Example: Knowledge Read Command

User instruction:

- `读一下 founder knowledge os 里面那篇 restructure blogs into course`

Agent behavior:

1. map to `Read Knowledge Space Directory`
2. identify the correct `spaceSlug` and `entrySlug`
3. if the entry is protected and there is no valid grant, ask for a knowledge access token
4. once grant is available, read the entry
5. summarize the document structure and important points

## Response Style

When this skill executes a task:

- use concise operational language
- state the action before the API call in one line
- after success, report only the key fields and important changes
- after failure, explain whether it was `401`, `403`, or `400`, and what the user needs to provide next

## Execution Output Contract

After every successful standard action, report results in this shape:

- action performed
- target resource
- key identifiers: `id`, `slug`, and `status`
- one-line note about important changed fields

Example:

- `Created article draft`
- `id: 12`
- `slug: ai-builder-roundup`
- `status: draft`
- `Updated fields: title, summary, content`

## Clarification Boundary

Ask the user only when one of these is true:

- there is no reliable way to identify the target record
- a create action is missing its minimum required field
- the requested change would overwrite an array field but the intended replacement is unclear
- the task needs a scope the current token does not have

Otherwise, execute with reasonable defaults and state those defaults in the result.

## Error Handling

Interpret responses like this:

- `401 unauthorized`: token missing, invalid, or inactive
- `403 forbidden`: token is valid but the scope does not allow this resource
- `400`: request body or state transition is invalid

Knowledge-specific errors:

- `403 knowledge entry is locked`: the Space is protected and there is no valid grant
- `403 knowledge asset is locked`: the Space is protected and the asset request is missing or using the wrong grant
- `404 knowledge space not found`: wrong `spaceSlug` or the Space is not published
- `404 knowledge entry not found`: wrong `entrySlug` inside the current Space
- `404 knowledge asset not found`: wrong `assetID` for the current Space

When blocked by `403`, do not retry blindly. Report the missing scope explicitly.

## Operating Rules

- Prefer the smallest scope set that can finish the task
- For writes, log the intent in plain language before sending the request
- Do not call admin-only configuration endpoints with an agent token
- If a task needs broader access, ask for a new token instead of reusing a human session
- Use the dedicated `knowledge_tokens:*` scopes for knowledge access token lifecycle work
- When reading knowledge content, prefer Space directory first, then entry, then asset resolution

## Minimal Validation Recipe

Use this smoke test to confirm the connection:

```bash
curl -s "$VOIDLAB_API_BASE_URL/api/v1/auth/me" \
  -H "Authorization: Bearer $VOIDLAB_AGENT_TOKEN"

curl -s "$VOIDLAB_API_BASE_URL/api/v1/articles" \
  -H "Authorization: Bearer $VOIDLAB_AGENT_TOKEN"
```

If the target token only has `articles:write`, article writes should succeed while unrelated modules such as `/api/v1/events` should return `403`.
