# VOIDLAB BuilderNetwork Agent

Use this workspace instruction file when Codex should operate BuilderNetwork through the VOIDLAB market API instead of the website UI.

## Required Environment

- `VOIDLAB_MARKET_BASE_URL`
- `VOIDLAB_MARKET_AGENT_TOKEN`

Production default:

```text
VOIDLAB_MARKET_BASE_URL=https://voidlab.club
```

## Registration And Token Setup

The operations agent provides a BuilderNetwork registration link containing an invitation code. Register through the link, sign in, and open `/market/workspace`. Create your own Agent Token in `Agent / API 接入`, select only the required scopes, and configure it as `VOIDLAB_MARKET_AGENT_TOKEN`.

The operations agent only issues registration invitations. Never share the member password or Agent Token with the operator.

Protected routes are built as:

```text
{VOIDLAB_MARKET_BASE_URL}/api/v1/...
```

## Authentication

Use Bearer auth on every protected request:

```text
Authorization: Bearer $VOIDLAB_MARKET_AGENT_TOKEN
```

## Use This When

Use this file when the user wants Codex to:

- manage their Builder profile
- publish or update works
- create or update Projects and their plans or teams
- publish or update opportunities
- inspect or organize contact requests
- read authorized contact exchange results
- operate BuilderNetwork without opening the website

## First-Run Handshake

Before any write operation:

1. call `GET /healthz`
2. call `GET /api/v1/network/me`
3. call a cheap read on the target resource
4. only then perform the write

## Scope Map

- `network_profiles:read`
- `network_profiles:write`
- `network_works:read`
- `network_works:write`
- `network_projects:read`
- `network_projects:write`
- `network_opportunities:read`
- `network_opportunities:write`
- `network_contact_requests:read`
- `network_contact_requests:write`
- `network_contact_exchange:read`
- `network_contact_exchange:write`

If a request returns `403`, explain which scope is missing or whether the action is session-only.

## Agent-Accessible Endpoints

- `GET /api/v1/network/me`
- `PUT /api/v1/network/me`
- `GET /api/v1/network/me/contacts`
- `PUT /api/v1/network/me/contacts`
- `GET /api/v1/network/works`
- `POST /api/v1/network/works`
- `PUT /api/v1/network/works/:id`
- `GET /api/v1/network/projects`
- `GET /api/v1/network/projects/:id`
- `POST /api/v1/network/projects`
- `PUT /api/v1/network/projects/:id`
- `GET /api/v1/network/opportunities`
- `POST /api/v1/network/opportunities`
- `PUT /api/v1/network/opportunities/:id`
- `GET /api/v1/network/contact-requests`
- `POST /api/v1/network/contact-requests`
- `POST /api/v1/network/contact-requests/:id/withdraw`
- `PUT /api/v1/network/contact-requests/:id/progress`
- `GET /api/v1/network/contact-exchanges/:id`
- `POST /api/v1/network/contact-exchanges/:id/revoke`

## Session-Only Actions

Do not attempt these with an agent token:

- `POST /api/v1/network/contact-requests/:id/accept`
- `POST /api/v1/network/contact-requests/:id/decline`

Hand those actions back to the member session and explain the safety boundary.

## Public Discovery

Use public routes for browsing and discovery:

- `GET /api/v1/public/network/discovery/profiles`
- `GET /api/v1/public/network/discovery/works`
- `GET /api/v1/public/network/discovery/opportunities`
- `GET /api/v1/public/network/projects`
- `GET /api/v1/public/network/projects/:slug`

## Operating Rules

- Prefer direct API work over describing website clicks.
- Verify scope before writes.
- Treat contact exchange as revocable private data.
- Only use progress values `contacted`, `discussing`, `collaborating`, `completed`, `closed`.
