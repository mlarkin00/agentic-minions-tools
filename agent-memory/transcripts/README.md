# transcripts/

Tier-3 raw task transcripts. One JSONL file per task, named
`YYYY-MM-DD-<task_id>.jsonl`. Append-only — never edited after commit.

## Purpose

Transcripts are the forensic record: every harness event (tool call, state
transition, model response, error) captured verbatim. They are **never**
injected into context. The `search_transcript` harness tool greps them on
demand for post-hoc investigation.

## Line schema

Each line is a JSON object with at minimum:

```json
{"ts":"2026-04-16T22:40:00Z","task_id":"t_abc","kind":"TOOL_CALL","role":"coding-design","body":{...}}
```

`kind` matches the harness `EventKind` enum (17 values — see
`internal/harness/observe/emitter.go` in agentic-minions).

## Retention

90 days in-repo, then cold-archive to
`gs://agentic-minions-memory-archive/transcripts/` and delete from the repo.
A scheduled job performs the archive; see §11 of the design doc.
