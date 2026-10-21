---
name: "voidlab-buildernetwork-agent"
description: "Operates BuilderNetwork through the VOIDLAB market API with agent-token auth. Invoke when a user wants an agent to manage profiles, works, opportunities, requests, or contact exchange without opening the website."
---

# VOIDLAB BuilderNetwork Agent

Use this skill when an agent needs to operate BuilderNetwork through the API instead of the website UI.

This skill is designed for end users who want to plug their own agent into VOIDLAB BuilderNetwork and let that agent:

- read and update their Builder profile
- create or update works
- create or update opportunities
- review incoming and outgoing contact requests
- read authorized contact exchange results
- track collaboration progress after a request is accepted

## When To Invoke

Invoke this skill when the user asks to:

- use their own agent to operate BuilderNetwork
- publish a work or opportunity without opening the website
- review or organize BuilderNetwork requests through API calls
- check whether a market token has enough permission
- connect a third-party host such as Trae, Codex, or Claude Code to BuilderNetwork

Do not invoke this skill for admin console operations outside BuilderNetwork.

## Required Environment

Set these values before making requests:

- `VOIDLAB_MARKET_BASE_URL`: site origin, default production value is `https://voidlab.club`
- `VOIDLAB_MARKET_AGENT_TOKEN`: BuilderNetwork agent token created by the member from `/market/workspace`

Optional local default:

- if `VOIDLAB_MARKET_BASE_URL` is missing during local development, assume `http://127.0.0.1:18081`

## User Onboarding

The operations agent distributes registration invites; this member agent operates the account after registration. Keep these credentials separate.

1. Open the registration link supplied by the operations agent, usually `https://voidlab.club/market/register?invite=INVITE_CODE`.
2. Complete registration. The invite is prefilled and verified by the registration page.
3. Sign in and open `/market/workspace`.
4. In `Agent / API 接入`, create a personal BuilderNetwork Agent Token and select only the scopes this agent needs.
5. Configure that token as `VOIDLAB_MARKET_AGENT_TOKEN`.

The member creates and owns this token. Never ask the operations agent for a member token, and never send passwords or tokens back to operations.

## URL Rule

Build protected API URLs as:

```text
{VOIDLAB_MARKET_BASE_URL}/api/v1/...
```

Examples:

- `https://voidlab.club/api/v1/network/me`
- `https://voidlab.club/api/v1/network/works`
- `https://voidlab.club/api/v1/network/opportunities`

## Authentication Rule

Use Bearer auth on every protected BuilderNetwork request:

```bash
Authorization: Bearer $VOIDLAB_MARKET_AGENT_TOKEN
```

Never ask the user for a website password when an agent token is enough.

## First-Run Handshake

Before any write operation, do this sequence:

1. `GET /healthz`
2. `GET /api/v1/network/me`
3. a cheap read for the target resource
4. then perform the write

Example:

```bash
curl -s "$VOIDLAB_MARKET_BASE_URL/healthz"
curl -s "$VOIDLAB_MARKET_BASE_URL/api/v1/network/me" \
  -H "Authorization: Bearer $VOIDLAB_MARKET_AGENT_TOKEN"
curl -s "$VOIDLAB_MARKET_BASE_URL/api/v1/network/works" \
  -H "Authorization: Bearer $VOIDLAB_MARKET_AGENT_TOKEN"
```

## Scope Map

Supported BuilderNetwork agent scopes:

- `network_profiles:read`
- `network_profiles:write`
- `network_works:read`
- `network_works:write`
- `network_opportunities:read`
- `network_opportunities:write`
- `network_contact_requests:read`
- `network_contact_requests:write`
- `network_contact_exchange:read`
- `network_contact_exchange:write`

Scope behavior:

- `*:read` is for list, detail, and lightweight inspection routes
- `*:write` is for create, update, withdraw, revoke, and other write actions on that module
- if a token does not have the required scope, expect `HTTP 403`

## Agent-Accessible Endpoints

### Profile

- `GET /api/v1/network/me`
- `PUT /api/v1/network/me`
- `GET /api/v1/network/me/contacts`
- `PUT /api/v1/network/me/contacts`

### Works

- `GET /api/v1/network/works`
- `POST /api/v1/network/works`
- `PUT /api/v1/network/works/:id`

### Opportunities

- `GET /api/v1/network/opportunities`
- `POST /api/v1/network/opportunities`
- `PUT /api/v1/network/opportunities/:id`

### Contact Requests

- `POST /api/v1/network/contact-requests`
- `GET /api/v1/network/contact-requests`
- `POST /api/v1/network/contact-requests/:id/withdraw`
- `PUT /api/v1/network/contact-requests/:id/progress`

### Contact Exchange

- `GET /api/v1/network/contact-exchanges/:id`
- `POST /api/v1/network/contact-exchanges/:id/revoke`

## Session-Only Boundary

The following actions currently require a member session and should not be attempted with an agent token:

- `POST /api/v1/network/contact-requests/:id/accept`
- `POST /api/v1/network/contact-requests/:id/decline`

If the user asks an agent to accept or decline a request, explain that the current product keeps those actions under the member session for safety and confirmation.

## Public Discovery Endpoints

These routes do not require an agent token and can be used for discovery:

- `GET /api/v1/public/network/profiles`
- `GET /api/v1/public/network/works`
- `GET /api/v1/public/network/opportunities`
- `GET /api/v1/public/network/discovery/profiles`
- `GET /api/v1/public/network/discovery/works`
- `GET /api/v1/public/network/discovery/opportunities`

Use the public routes when the task is browsing or discovery rather than member-owned operations.

## Operating Rules

- Prefer the API over describing website clicks.
- Check scope and identity before writes.
- Keep contact handling conservative: only read exchange data after acceptance.
- Treat revocation as authoritative: if an exchange has been revoked, do not assume previously revealed fields are still valid.
- When updating collaboration progress, only use these values: `contacted`, `discussing`, `collaborating`, `completed`, `closed`.
- If an endpoint rejects the request because the token is under-scoped, tell the user which scope is missing instead of retrying blindly.

## Example Workflows

### Publish A Work

1. Read `/network/me`
2. Check `/network/works`
3. `POST /network/works` with title, summary, body, tags, and status

### Publish An Opportunity

1. Read `/network/me`
2. Check `/network/opportunities`
3. `POST /network/opportunities` with title, summary, opportunity type, roles, and status

### Review Requests

1. Read `/network/contact-requests?direction=incoming`
2. If the user wants to accept or decline, hand off to the member session
3. After acceptance, read `/network/contact-exchanges/:id`
4. If needed, update `/network/contact-requests/:id/progress`

## Error Handling

- `401 unauthorized`: token missing, invalid, inactive, or expired in context
- `403 forbidden`: token exists but lacks the required scope, or the action is session-only
- `404 not found`: resource not visible to the current user or no longer available
- `400 validation`: request body is invalid, status value is unsupported, or required fields are missing

When a write fails, prefer explaining the exact boundary:

- missing token
- wrong base URL
- missing scope
- session-only action
- validation error in payload

## Minimal Curl Examples

Read current profile:

```bash
curl -s "$VOIDLAB_MARKET_BASE_URL/api/v1/network/me" \
  -H "Authorization: Bearer $VOIDLAB_MARKET_AGENT_TOKEN"
```

Create a work:

```bash
curl -s "$VOIDLAB_MARKET_BASE_URL/api/v1/network/works" \
  -H "Authorization: Bearer $VOIDLAB_MARKET_AGENT_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "title": "Realtime AI Research Dashboard",
    "summary": "A live dashboard for agent-generated insight tracking.",
    "body": "Built with Go, Vue, and scheduled data collection.",
    "tags": ["AI", "Dashboard", "Automation"],
    "status": "published"
  }'
```

Create an opportunity:

```bash
curl -s "$VOIDLAB_MARKET_BASE_URL/api/v1/network/opportunities" \
  -H "Authorization: Bearer $VOIDLAB_MARKET_AGENT_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "title": "Looking for an AI infra engineer",
    "summary": "Need help building a production-grade agent workflow.",
    "opportunity_type": "collaboration",
    "looking_for_roles": ["Engineer"],
    "required_skills": ["Go", "Kubernetes"],
    "remote_allowed": true,
    "status": "open"
  }'
```
