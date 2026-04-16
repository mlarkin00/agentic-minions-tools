# agentic-minions-mcp

A local MCP (Model Context Protocol) server that connects Claude Code, Claude
Desktop, Gemini CLI, or any MCP-compatible client to the agentic-minions fleet
running on GCP.

## Prerequisites

- **Go 1.25+** installed
- **gcloud CLI** installed and authenticated
- Access to the `mslarkin-agents` GCP project (or your own deployment)

Authenticate so the MCP server can obtain ID tokens for Cloud Run:

```sh
gcloud auth application-default login
```

## Install

### Claude Code

The repo ships a `.mcp.json` that Claude Code picks up automatically when
you open this project:

```jsonc
// .mcp.json (already committed)
{
  "mcpServers": {
    "agentic-minions": {
      "command": "go",
      "args": ["run", "."],
      "env": {
        "GATEWAY_URL": "https://gateway-845186993936.us-central1.run.app",
      },
    },
  },
}
```

No manual setup needed — open the project in Claude Code and the
`agentic-minions` tools appear automatically.

### Claude Desktop

Add to `~/Library/Application Support/Claude/claude_desktop_config.json`
(macOS) or `%APPDATA%\Claude\claude_desktop_config.json` (Windows):

```json
{
  "mcpServers": {
    "agentic-minions": {
      "command": "go",
      "args": ["run", "."],
      "cwd": "/path/to/agentic-minions-mcp",
      "env": {
        "GATEWAY_URL": "https://gateway-845186993936.us-central1.run.app"
      }
    }
  }
}
```

### Gemini CLI

Add to `~/.gemini/settings.json`:

```json
{
  "mcpServers": {
    "agentic-minions": {
      "command": "go",
      "args": ["run", "."],
      "cwd": "/path/to/agentic-minions-mcp",
      "env": {
        "GATEWAY_URL": "https://gateway-845186993936.us-central1.run.app"
      }
    }
  }
}
```

### Pre-built binary (optional)

```sh
go build -o mcp-server .
```

Then replace `"command": "go"` and `"args": ["run", "."]` with
`"command": "/path/to/mcp-server"` and `"args": []`.

## Configuration

| Environment variable | Required | Description                                 |
| -------------------- | -------- | ------------------------------------------- |
| `GATEWAY_URL`        | Yes      | Base URL of the gateway service (Cloud Run) |

Authentication is automatic via:

1. **Service account credentials** (if `GOOGLE_APPLICATION_CREDENTIALS` is set)
2. **Application Default Credentials** (from `gcloud auth application-default login`)
3. **gcloud CLI fallback** (calls `gcloud auth print-identity-token`)

## MCP Tools

| Tool             | Description                                    |
| ---------------- | ---------------------------------------------- |
| `create_session` | Start a new conversation with an agent role    |
| `send_message`   | Send a prompt and receive the agent's response |
| `list_sessions`  | List active sessions for a user/role           |
| `delete_session` | Clean up a session                             |

### Example usage (Claude Code)

```
> create a session with the coding-design agent

> ask it to design a rate limiter in Go

> list my active sessions

> delete the session when done
```

### Available agent roles

| Role            | Description                                                   | Backend                       |
| --------------- | ------------------------------------------------------------- | ----------------------------- |
| `coding-design` | System design, code architecture, implementation, code review | Claude Opus 4.6 via Vertex AI |

## Troubleshooting

**"GATEWAY_URL environment variable is required"**
The MCP server requires `GATEWAY_URL` in its env block. Check your MCP config.

**"gcloud auth print-identity-token failed"**
Run `gcloud auth application-default login` to refresh credentials.

**"create session failed (403)"**
Your Google identity doesn't have access to the Cloud Run gateway. Verify
with: `gcloud auth print-identity-token | head -c 20` (should print a token).

**"send message failed (502/504)"**
The backend agent pod may be starting up. GKE pods take ~30s on cold start.
Retry after a moment.
