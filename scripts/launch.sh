#!/usr/bin/env sh
# MCP launcher for Claude Code plugins.
#
# Claude Code installs plugins via git clone (no release assets), so we can't
# ship a prebuilt binary in the usual sense. Instead this wrapper builds once
# into the plugin root; Go's build cache makes subsequent starts near-instant
# and auto-rebuilds on source updates (e.g. after `/plugin update`).
set -eu

root="${CLAUDE_PLUGIN_ROOT:-$(CDPATH='' cd -- "$(dirname -- "$0")/.." && pwd)}"
cd "$root"

# Build stdout must not pollute MCP's stdio stream — send build output to stderr.
go build -o agentic-minions . 1>&2

exec ./agentic-minions
