# TODO

## P1 — Important / Unblocking

- [ ] **[P1]** Verify sub-agent triggering for `coding-design-minion` and `security-auditor-minion` in a fresh session.
- [x] **[P1]** Update MCP tool descriptions to use canonical role names — `main.go` now enumerates the 9 valid roles in `create_session` and references them from `send_message`/`list_sessions`/`delete_session`. Local proxy agents in `agents/` and `.gemini/agents/` renamed to canonical roles (`designing-code`, `generating-code`, `reviewing-code`); `pm-assistant` and `pm-mentor` retained. AGENTS.md and README.md updated.
- [ ] **[P1]** Implement `go vet` and `golangci-lint` in a CI workflow — Ensure code quality is maintained automatically.
- [ ] **[P1]** Add more comprehensive integration tests — Verify end-to-end connectivity with a mock gateway.

## P2 — Nice-to-Have

- [ ] **[P2]** Improve error messages for auth failures — Provide more actionable advice when `gcloud` tokens expire.
- [ ] **[P2]** Support custom timeout configuration — Allow users to override the default HTTP timeouts for slow agent responses.
