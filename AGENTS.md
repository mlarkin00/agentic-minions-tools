# agentic-minions — Agent Briefing

## Project Context

A local MCP (Model Context Protocol) server that connects clients to the agentic-minions fleet on GCP. Built with Go and `mcp-go`.

## Operational Commands

- **Build**: `go build -o .bin/agentic-minions .`
- **Run**: `go run .` (Requires `GATEWAY_URL` env var)
- **Test**: `go test ./...`
- **Lint**: `go vet ./...`

## Style & Conventions

- Standard Go conventions (`gofmt`, `go vet`).
- Use `github.com/mark3labs/mcp-go` for MCP implementation.
- Error handling: wrap errors with context.

## Architecture & Constraints

- MCP server proxies to a GCP gateway (Cloud Run).
- Authentication via Application Default Credentials (ADC) or gcloud.
- Key components:
  - `main.go`: Entry point, tool registration.
  - `internal/mcpclient/`: Core logic for communicating with the gateway.
- **Never** hardcode the `GATEWAY_URL`.

## Local Minions

This extension provides specialized sub-agents (minions) that proxy requests to the remote agentic-minions fleet.

- **designing-code**: Elite Senior Software Architect for blueprints and technical specs.
- **generating-code**: Senior Software Engineer for high-fidelity code implementation.
- **reviewing-code**: Lead Software Architect for high-criticality audits and design reviews.
- **pm-assistant**: Senior PM for generating and refining artifacts (PRDs, user stories).
- **pm-mentor**: Senior PM Advisor for strategic guidance and prioritization.

Additional canonical roles available via the gateway (no local proxy file yet): `advising-on-code`, `validating-code`, `maintaining-codebase-health`, `authoring-technical-content`.

## Memory Discipline (MANDATORY)

This project runs a tiered memory system (T0 working · T1 task state · T2 session workspace · T3 project brain · T4 long-term learning). See `agent-memory/README.md`. The mandates below apply to **every** agent session and **every** task. They are non-negotiable.

> **Rollout status (as of 2026-04-17):** T4 store is in rollout. Mandates #1–#3 apply using the seed index; mandates #10–#11 activate when the reflection runner ships.

### Session start — absolute mandates

1. **MEMORY.md MUST be loaded into the working prompt at session start. No exceptions.** The T4 index (`agent-memory/MEMORY.md`, ≤25KB, ≤100 entries) is injected as part of Layer 1 of the prompt. If the fetch fails, proceed with the last cached index and emit a degraded-mode warning — but never start a session without an index in context.
2. **The active role instruction file MUST be loaded** (`AGENTS.md`). This is the T3 project brain and anchors the system prompt.
3. **Consult MEMORY.md entries before planning.** For any pointer whose description matches the task, call the memory-read tool to pull the topic body into context.

### Execution — write rules

4. **Offload any tool output >2K tokens to T2.** Write the raw payload to the session workspace (`workspace/<task_id>/scratch/…`) and replace the in-message body with a one-line manifest reference (path + ≤200-token summary).
5. **Checkpoint T1 after every state-machine transition and every tool result.** Never skip a checkpoint for latency — T1 is the sole recovery path after pod restart.
6. **Never write to T4 in the hot loop.** Long-term memory is written only at task end by the reflection pass. The user response must never block on a memory write.
7. **Honor role memory namespaces.** An agent reads and proposes edits only to topics its role is permitted to touch. Governance enforces this; do not attempt to bypass.

### Conflict & failure modes

8. **Reality primacy.** If a MEMORY.md pointer references a file, symbol, or fact that no longer matches the environment, the environment wins. Prune the pointer and flag the discrepancy for the reflection queue. Never hallucinate around a stale memory.
9. **Degraded-mode operation.** T4 fetch failure, topic read miss, or governance denial MUST NOT halt the task. Log the failure, continue with the best available context, and surface the degradation in task output.

### Task end — reflection pass

10. **Enqueue the reflection pass on `TASK_DONE` or `TASK_FAILED`.** The pass runs out-of-band, extracts only facts / preferences / corrections / mistakes, and proposes diffs against T4 topics. It does not store raw logs, chain-of-thought, or PR narrative.
11. **Only the maintainer role writes to T4.** Proposed diffs land as PRs (auto-merged for low-risk types, human-reviewed otherwise). The hot-path agent never commits directly.

### Never do these

- Never load tool output >2K tokens directly into the next LLM turn — always offload to T2 first.
- Never retry a T4 write inline — queue it to disk and let the reflection runner retry.
- Never skip loading MEMORY.md "because the task looks simple." The index is cheap; missing context is expensive.
- Never edit or delete existing T4 topic files in place when superseding — write a new topic and set the old one's `valid_until` (bi-temporal model).
