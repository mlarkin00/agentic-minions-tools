# agentic-minions — Design

## Overview

A local MCP server that acts as a proxy for a fleet of Claude-powered agents running as Cloud Run services on GCP.

## Component Map

[Client (Claude Code/Desktop, Gemini CLI)]
↓ (JSON-RPC via stdio)
[Local MCP Server (Go)]
↓ (HTTPS + Auth)
[GCP Gateway (Cloud Run)]
↓ (Internal Routing)
[Agent Services (GKE/Cloud Run)]

## Components

### Local MCP Server — ./

The Go application that implements the MCP protocol.
Entry point: `main.go`.
Key invariant: All requests to the gateway MUST be authenticated with a valid identity token.

### MCP Client Logic — ./internal/mcpclient/

Handles the communication with the remote gateway, including SSE (Server-Sent Events) streaming for agent responses.

## State Machine / Data Flow

1. Client calls a tool (e.g., `create_session`).
2. MCP Server obtains an ID token for the `GATEWAY_URL`.
3. MCP Server sends an HTTP request to the Gateway.
4. Gateway routes to the appropriate agent service.
5. Agent responds (streaming via SSE for `send_message`).
6. MCP Server translates SSE to MCP progress/result updates.

## Invariants

- `GATEWAY_URL` MUST be provided via environment variables.
- SSE streams MUST be correctly closed to avoid memory leaks.
- All agent communication MUST happen through the Gateway.

## Claude Code Plugin — `.claude-plugin/`

### How the installer works

Claude Code installs the plugin by cloning the GitHub repo into:

```
~/.claude/plugins/cache/<marketplace>/<plugin>/<version>/
~/.claude/plugins/marketplaces/<marketplace>/   ← live git clone, used for version lookup
```

The installer also builds the Go binary using the plugin name. It places a copy at:

```
~/.claude/plugins/cache/<marketplace>/<plugin-binary-name>
```

**Critical constraint**: the binary name must NOT equal the plugin name (`agentic-minions`), because the installer needs to create a directory at `cache/agentic-minions/agentic-minions/` — and if a file of that name already exists there, the next install fails with `ENOTDIR`.

This is why `scripts/launch.sh` and the release CI both build to `server` / `server.exe` rather than `agentic-minions`.

### Known failure mode: ENOTDIR on install

**Symptom**: `/plugin install agentic-minions` fails with:

```
Failed to install: ENOTDIR: not a directory, rm
  '/home/<user>/.claude/plugins/cache/agentic-minions/agentic-minions/<version>'
```

**Cause**: `cache/agentic-minions/agentic-minions` exists as a FILE (old build artifact) instead of a directory, blocking the versioned subdirectory from being created.

**Fix — full clean reinstall**:

```bash
# 1. Remove cache
rm -rf ~/.claude/plugins/cache/agentic-minions/

# 2. Remove marketplaces copy (stale version source)
rm -rf ~/.claude/plugins/marketplaces/agentic-minions/

# 3. Remove installed_plugins.json entry
python3 -c "
import json
path = '/home/\$USER/.claude/plugins/installed_plugins.json'
d = json.load(open(path))
d.get('plugins', {}).pop('agentic-minions@agentic-minions', None)
json.dump(d, open(path, 'w'), indent=2)
"

# 4. Re-clone marketplaces dir (installer requires it to exist)
git clone --depth 1 https://github.com/mlarkin00/agentic-minions-tools.git \
  ~/.claude/plugins/marketplaces/agentic-minions/

# 5. Reinstall
# /plugin install agentic-minions
```

### Version sync requirement

`plugin.json` and `marketplace.json` versions MUST stay in sync. The CI (`release.yml`) bumps both on every push to `main`. If they drift, the installer detects a version mismatch and fires an auto-update against an inconsistent state, which can trigger the ENOTDIR failure.

The `validate` job in `release.yml` blocks the release pipeline if the versions differ.
