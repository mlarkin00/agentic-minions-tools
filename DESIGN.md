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
