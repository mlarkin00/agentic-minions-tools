# topics/

Tier-2 topic files. One topic per file. Each file starts with YAML frontmatter
that MUST validate against [`../schemas/topic.schema.json`](../schemas/topic.schema.json).

## Frontmatter fields

| Field         | Required | Notes                                                      |
| ------------- | -------- | ---------------------------------------------------------- |
| `name`        | yes      | Kebab-case, matches filename stem                          |
| `description` | yes      | One sentence; used for fuzzy match at read time            |
| `type`        | yes      | `user` · `project` · `feedback` · `reference`              |
| `roles`       | yes      | List of roles permitted to read this topic                 |
| `valid_from`  | yes      | ISO date the fact became true                              |
| `valid_until` | no       | ISO date the fact was superseded. `null` or omit = current |
| `supersedes`  | no       | Filename of the topic this one replaces                    |

## Body rules

- Keep under ~4K tokens so a single `read_memory_topic` call fits the budget.
- Lead with the fact or rule.
- For `feedback` and `project` types, include `**Why:**` and `**How to apply:**`
  lines so future readers can judge edge cases.
- Never include raw logs or chain-of-thought — only distilled facts.
