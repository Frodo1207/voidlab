# VOIDLAB BuilderNetwork Agent Pack

This package lets your users connect their own agent to BuilderNetwork and operate the market through the API instead of opening the website UI.

## What Is Included

- `trae/voidlab-buildernetwork-agent/SKILL.md`
- `codex/AGENTS.md`
- `claude-code/CLAUDE.md`

All three files describe the same BuilderNetwork operating protocol, adapted to the common host conventions for Trae, Codex, and Claude Code.

## What This Pack Enables

With a BuilderNetwork agent token, a user agent can:

- read and update the member profile
- create and update works
- create and update opportunities
- read incoming and outgoing requests
- read authorized contact exchange results
- update collaboration progress

## Current Boundary

For safety, accepting or declining a request is still member-session only:

- `POST /api/v1/network/contact-requests/:id/accept`
- `POST /api/v1/network/contact-requests/:id/decline`

Agents should hand those actions back to the member session.

## Required Values

Set these in the target host environment:

- `VOIDLAB_MARKET_BASE_URL`
- `VOIDLAB_MARKET_AGENT_TOKEN`

Production default:

```text
VOIDLAB_MARKET_BASE_URL=https://voidlab.club
```

## Get Your Agent Token

1. Receive a BuilderNetwork invitation link from the community operator.
2. Register from that link; the invitation code is prefilled on the registration page.
3. Sign in and open `/market/workspace`.
4. In `Agent / API 接入`, create your own Agent Token with the minimum scopes needed.
5. Configure the token as `VOIDLAB_MARKET_AGENT_TOKEN` in your host.

The operator only issues registration invitations. Your Agent Token belongs to you and is created by you after registration. Do not share your password or token with the operator.

## Install

### Trae

1. Create `.trae/skills/voidlab-buildernetwork-agent/`
2. Put `SKILL.md` from `trae/voidlab-buildernetwork-agent/` into that folder

### Codex

1. Copy `codex/AGENTS.md` into your workspace as `AGENTS.md`
2. Keep the environment values available to the host

### Claude Code

1. Copy `claude-code/CLAUDE.md` into your workspace as `CLAUDE.md`
2. Keep the environment values available to the host

## Download Files

- Trae: `/downloads/voidlab-buildernetwork-agent/trae/voidlab-buildernetwork-agent/SKILL.md`
- Codex: `/downloads/voidlab-buildernetwork-agent/codex/AGENTS.md`
- Claude Code: `/downloads/voidlab-buildernetwork-agent/claude-code/CLAUDE.md`
