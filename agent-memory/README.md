# agent-memory

Long-term memory (Tier 4) for the [agentic-minions](https://github.com/mlarkin00/agentic-minions) fleet.

This repository is the persistent knowledge store that every minion reads at
session start and that the reflection pass writes to at task end. Git is the
storage engine — versioning, review, rollback, and audit come for free.

See [`docs/designs/2026-04-16-multi-tier-memory-design.md`](https://github.com/mlarkin00/agentic-minions/blob/main/docs/designs/2026-04-16-multi-tier-memory-design.md)
in the agentic-minions repo for the full architecture.

## Layout

```
MEMORY.md              Tier-1 global index (≤100 entries, ≤25KB)
topics/                Tier-2 topic files (one topic per file, frontmatter + body)
transcripts/           Tier-3 raw JSONL (append-only, search-only)
schemas/               JSON Schema for frontmatter + index validation
```

## Write rules

- Only the `maintaining-agent-memory` role writes here.
- Writes happen at task end via the reflection pass — never inside the hot loop.
- Commit to a topic file **before** adding its pointer to `MEMORY.md`.
- Never delete a superseded topic; set `valid_until` and write a replacement.
- Never dump raw logs, PR narrative, or chain-of-thought.

## Read rules

- `MEMORY.md` is loaded into every session's system prompt.
- Topic bodies are loaded on demand via the `read_memory_topic` harness tool.
- Transcripts are never injected into context — search only.
