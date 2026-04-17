# agentic-minions

You have access to a fleet of Claude-powered agents via the agentic-minions
MCP tools. Each agent specializes in a role (e.g. `coding-design` for system
design, code architecture, implementation, and code review).

## Workflow

1. **Create a session** with `create_session` — specify the agent `role` and
   a `user_id`. You can optionally specify an `agent_name`. You receive a `session_id`.
2. **Send messages** with `send_message` — pass the `role`, `user_id`,
   `session_id`, and your `message`. You can optionally specify an `agent_name`.
   The agent streams its response via SSE and returns the final answer.
3. **List sessions** with `list_sessions` to see active conversations.
4. **Delete sessions** with `delete_session` when done.

## Available roles

| Role            | Specialty                                                     |
| --------------- | ------------------------------------------------------------- |
| `coding-design` | System design, code architecture, implementation, code review |

## Tips

- Sessions are persistent — you can send multiple messages in the same session
  to build on prior context.
- Use `user_id` consistently (e.g. the user's email) so the agent can track
  sessions per user.
- The backend agents run on GKE and may take ~30s on cold start. If you get a
  502/504, retry after a moment.

## Troubleshooting

### Installation Error: 404/Tag Undefined

If you see an error like `Failed to fetch release data for ... at tag undefined: Request failed with status code 404`, it usually means the extension has not yet had a tagged release published on GitHub. Ensure that the first release has been created by the automated workflow.
