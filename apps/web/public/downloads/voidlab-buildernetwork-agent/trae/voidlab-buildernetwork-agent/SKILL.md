---
name: "voidlab-buildernetwork-agent"
description: "Operates BuilderNetwork through the VOIDLAB market API with agent-token auth. Invoke when a user wants an agent to manage profiles, works, opportunities, requests, or contact exchange without opening the website."
---

# VOIDLAB BuilderNetwork Agent

Use this skill when an agent needs to operate BuilderNetwork through the API instead of the website UI.

## Required Environment

- `VOIDLAB_MARKET_BASE_URL`
- `VOIDLAB_MARKET_AGENT_TOKEN`

Production default:

```text
VOIDLAB_MARKET_BASE_URL=https://voidlab.club
```

## Registration And Token Setup

The operations agent gives the user a registration link with an invitation code. Register through that link, then sign in and open `/market/workspace`. Create your own BuilderNetwork Agent Token in `Agent / API 接入`, grant only the scopes needed for your tasks, and configure it as `VOIDLAB_MARKET_AGENT_TOKEN`.

The operations agent must not create, request, or receive the member's Agent Token or password.

Protected routes are built as:

```text
{VOIDLAB_MARKET_BASE_URL}/api/v1/...
```

## Authentication

Use Bearer auth:

```bash
Authorization: Bearer $VOIDLAB_MARKET_AGENT_TOKEN
```

## Invoke When

Invoke this skill when the user wants to:

- manage their Builder profile
- publish or update works
- create or update Projects and their plans or teams
- publish or update opportunities
- review requests through the API
- read contact exchange results after acceptance
- connect their own agent to BuilderNetwork instead of opening the website

## First-Run Handshake

Before any write:

1. `GET /healthz`
2. `GET /api/v1/network/me`
3. do a cheap read on the target module
4. then write

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

If the token lacks the required scope, expect `HTTP 403`.

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

## Session-Only Boundary

These actions are currently member-session only and should not be attempted with an agent token:

- `POST /api/v1/network/contact-requests/:id/accept`
- `POST /api/v1/network/contact-requests/:id/decline`

## Public Discovery

For browsing, prefer public routes:

- `GET /api/v1/public/network/discovery/profiles`
- `GET /api/v1/public/network/discovery/works`
- `GET /api/v1/public/network/discovery/opportunities`
- `GET /api/v1/public/network/projects`
- `GET /api/v1/public/network/projects/:slug`

## Operating Rules

- Prefer API operations over website walkthroughs.
- Verify identity and scope before writes.
- Treat contact exchange as private and revocable.
- Only use these progress values: `contacted`, `discussing`, `collaborating`, `completed`, `closed`.
- Explain missing scope or session-only boundaries clearly.
