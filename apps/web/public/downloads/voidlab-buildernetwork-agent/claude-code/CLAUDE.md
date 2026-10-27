# VOIDLAB BuilderNetwork Agent

Use this instruction file when Claude Code should operate BuilderNetwork through the VOIDLAB market API instead of the website UI.

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

Use Bearer auth on protected requests:

```text
Authorization: Bearer $VOIDLAB_MARKET_AGENT_TOKEN
```

## Use This When

Use this file when the user wants Claude Code to:

- manage the Builder profile
- publish or update works
- create or update Projects and their plans or teams
- publish or update opportunities
- inspect or organize contact requests
- read contact exchange results after acceptance
- operate BuilderNetwork without opening the website

## First-Run Handshake

Before any write:

1. request `GET /healthz`
2. request `GET /api/v1/network/me`
3. request a cheap read on the target module
4. then do the write

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

If a protected route returns `403`, explain whether the token is under-scoped or the action is session-only.

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

Do not use an agent token for:

- `POST /api/v1/network/contact-requests/:id/accept`
- `POST /api/v1/network/contact-requests/:id/decline`

Return those actions to the member session and explain the confirmation boundary.

## Public Discovery

Use public routes for browsing:

- `GET /api/v1/public/network/discovery/profiles`
- `GET /api/v1/public/network/discovery/works`
- `GET /api/v1/public/network/discovery/opportunities`
- `GET /api/v1/public/network/projects`
- `GET /api/v1/public/network/projects/:slug`

## Operating Rules

- Prefer API operations over UI instructions.
- Verify identity and scope before writes.
- Treat contact exchange as private and revocable.
- Only use progress values `contacted`, `discussing`, `collaborating`, `completed`, `closed`.
