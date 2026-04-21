---
name: local-build-hygiene
description: Mandatory rule — local `go build` binaries must be deleted immediately after verification; intentional binaries live only in `dist/`.
type: feedback
roles: [coding-design, maintaining-agent-memory]
valid_from: 2026-04-17
---

When verifying a change with `go build` (or any local build command) during a task, the produced binary MUST be deleted immediately after the build/test step completes. The only binaries that belong in the working tree are intentional artifacts committed under `dist/` (or an equivalent, clearly-named output directory).

**Why:** Stray binaries pollute `git status`, inflate `.gitignore` maintenance (see the anchored-pattern fix for `/claude-agent`, `/gateway`, `/mcp-server`, `/agentic-minions-mcp`), and risk being committed by accident into a code review. The repo's CI path relies on Cloud Build producing artifacts in Artifact Registry — no local binary is ever deployed.

**How to apply:**

- Prefer `go build -o /dev/null ./cmd/<bin>` (or `go vet` / `go test -c -o /dev/null`) when the goal is just compile-verification.
- When you do need the binary (e.g. to inspect output), put it under `dist/` (already ignored via `dist/` in `.gitignore`) or a temp dir, not the cmd package or repo root.
- If a `go build ./cmd/<bin>` produced a root-level binary during this task, remove it before reporting done. A task is not complete while there is an untracked build artifact in the tree.
- Applies equally to the MCP server, gateway, and reflection binaries — and to any future tooling.
